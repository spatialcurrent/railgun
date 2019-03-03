// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"path/filepath"
	//"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/middleware"
	"github.com/spatialcurrent/railgun/railgun/util"
)

type ServiceDownloadHandler struct {
	*BaseHandler
	Cache *gocache.Cache
}

func (h *ServiceDownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	vars := mux.Vars(r)

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)

	if v := ctx.Value("request"); v != nil {
		if req, ok := v.(middleware.Request); ok {
			req.Vars = vars
			req.Handler = reflect.TypeOf(h).Elem().Name()
			ctx = context.WithValue(ctx, "request", req)
		}
	}

	switch r.Method {
	case "GET":
		once := &sync.Once{}
		h.Catalog.ReadLock()
		defer once.Do(func() { h.Catalog.ReadUnlock() })
		h.SendDebug("read locked for " + r.URL.String())
		filename, obj, err := h.Get(w, r.WithContext(ctx), format, vars)
		once.Do(func() { h.Catalog.ReadUnlock() })
		h.SendDebug("read unlocked for " + r.URL.String())
		if err != nil {
			h.SendError(err)
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		} else {
			err = h.RespondWithObject(w, http.StatusOK, obj, format, filename)
			if err != nil {
				h.SendError(err)
				err = h.RespondWithError(w, err, format)
				if err != nil {
					panic(err)
				}
			}
		}
	default:
		err := h.RespondWithNotImplemented(w, format)
		if err != nil {
			panic(err)
		}
	}

}

func (h *ServiceDownloadHandler) Get(w http.ResponseWriter, r *http.Request, format string, vars map[string]string) (filename string, object interface{}, err error) {

	ctx := r.Context()

	now := time.Now()
	if v := ctx.Value("request"); v != nil {
		if req, ok := v.(middleware.Request); ok {
			now = *req.Start
		}
	}

	defer func() {
		if v := ctx.Value("log"); v != nil {
			if log, ok := v.(*sync.Once); ok {
				log.Do(func() {
					if v := ctx.Value("request"); v != nil {
						if req, ok := v.(middleware.Request); ok {
							req.Error = err
							end := time.Now()
							req.End = &end
							h.SendInfo(req.Map())
						}
					}
				})
			}
		}
	}()

	serviceName, ok := vars["name"]
	if !ok {
		return "", nil, &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}

	outputFilename := fmt.Sprintf("%s_%s.%s", serviceName, now.Format("20060102"), filepath.Ext(r.URL.Path))

	service, ok := h.Catalog.GetService(serviceName)
	if !ok {
		return "", nil, &rerrors.ErrMissingObject{Type: "service", Name: serviceName}
	}

	variables := h.AggregateMaps(
		h.GetServiceVariables(h.Cache, serviceName),
		service.Defaults)

	_, inputUri, err := dfl.EvaluateString(service.DataStore.Uri, variables, map[string]interface{}{}, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return "", nil, errors.Wrap(err, "invalid data store uri")
	}
	inputScheme, inputPath := grw.SplitUri(inputUri)

	cacheKeyDataStore := ""
	inputUris := make([]string, 0)
	lastModified := map[string]time.Time{}
	inputReaders := map[string]grw.ByteReadCloser{}
	inputObjects := make([]interface{}, 0)

	var s3Client *s3.S3
	if inputScheme == "s3" {

		s3Client, err = h.GetAWSS3Client()
		if err != nil {
			return "", nil, errors.Wrap(err, "error connecting to AWS")
		}

		i := strings.Index(inputPath, "/")
		if i == -1 {
			return "", nil, errors.New("path missing bucket")
		}

		bucket := inputPath[0:i]

		keys := make([]string, 0)
		if j := strings.Index(inputPath, "*"); j >= 0 {
			listObjectsOutput, err := s3Client.ListObjects(&s3.ListObjectsInput{
				Bucket: aws.String(bucket),
				Prefix: aws.String(inputPath[i+1 : j]),
			})
			if err != nil {
				return "", nil, errors.Wrap(err, "could not list objects for path "+inputPath)
			}
			for _, obj := range listObjectsOutput.Contents {
				if key := *obj.Key; strings.HasSuffix(key, inputPath[j+1:]) {
					keys = append(keys, key)
					lastModified[fmt.Sprintf("s3://%s/%s", bucket, key)] = *obj.LastModified
				}
			}
		} else {
			key := inputPath[i+1:]
			keys = append(keys, key)
			headObjectOutput, err := s3Client.HeadObject(&s3.HeadObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(key),
			})
			if err != nil {
				return "", nil, errors.Wrap(err, "error heading S3 object")
			}
			lastModified[fmt.Sprintf("s3://%s/%s", bucket, key)] = *headObjectOutput.LastModified
		}

		for _, key := range keys {
			inputUri := fmt.Sprintf("s3://%s/%s", bucket, key)
			inputUris = append(inputUris, inputUri)

			cacheKeyDataStore = h.BuildCacheKeyDataStore(
				service.DataStore.Name,
				inputUri,
				lastModified[fmt.Sprintf("s3://%s/%s", bucket, key)])

			if object, found := h.Cache.Get(cacheKeyDataStore); found {
				inputObjects = append(inputObjects, object)
			} else {
				inputObjects = append(inputObjects, nil)
			}

		}

	} else if inputScheme == "" || inputScheme == "file" {

		inputUris = append(inputUris, inputUri)

		inputFile, inputFileInfo, err := grw.ExpandOpenAndStat(inputPath)
		if err != nil {
			return "", nil, err
		}

		lastModified[inputUri] = inputFileInfo.ModTime()

		cacheKeyDataStore = h.BuildCacheKeyDataStore(
			service.DataStore.Name,
			inputUri,
			lastModified[inputUri])

		if object, found := h.Cache.Get(cacheKeyDataStore); found {
			inputObjects = append(inputObjects, object)
		} else {
			inputReader, err := grw.ReadFromFile(inputFile, service.DataStore.Compression, false, 4096)
			if err != nil {
				return "", nil, errors.Wrap(err, "error creating grw.ByteReadCloser for file at path \""+inputPath+"\"")
			}
			inputReaders[inputUri] = inputReader
			inputObjects = append(inputObjects, nil)
		}

	}

	wg, _ := errgroup.WithContext(ctx)
	for i, inputUri := range inputUris {
		i, inputUri := i, inputUri // closure
		wg.Go(func() error {

			if inputObjects[i] == nil {

				var inputBytes []byte
				if r, ok := inputReaders[inputUri]; ok {
					b, err := r.ReadAllAndClose()
					if err != nil {
						return errors.Wrap(err, "error reading from resource at uri "+inputUri)
					}
					inputBytes = b
				} else {
					b, err := grw.ReadAllAndClose(inputUri, service.DataStore.Compression, s3Client)
					if err != nil {
						return errors.Wrap(err, "error reading from resource at uri "+inputUri)
					}
					inputBytes = b
				}

				object, err := h.DeserializeBytes(inputBytes, service.DataStore.Format)
				if err != nil {
					return errors.Wrap(err, "error deserializing input")
				}
				inputObjects[i] = object

			}

			cacheKeyDataStore := h.BuildCacheKeyDataStore(
				service.DataStore.Name,
				inputUri,
				lastModified[inputUri])

			h.Cache.Set(cacheKeyDataStore, inputObjects[i], gocache.DefaultExpiration)

			return nil
		})
	}

	// Wait until all objects have been loaded
	if err := wg.Wait(); err != nil {
		return "", nil, errors.Wrap(err, "error fetching data")
	}

	variables, outputObject, err := service.Process.Node.Evaluate(
		variables,
		h.AggregateSlices(inputObjects),
		dfl.DefaultFunctionMap,
		dfl.DefaultQuotes)
	if err != nil {
		return "", nil, errors.Wrap(err, "error evaluating process with name "+service.Process.Name)
	}

	// Set the variables to the cache every time to bump the expiration
	h.SetServiceVariables(h.Cache, serviceName, variables)

	return outputFilename, gss.StringifyMapKeys(outputObject), nil

}
