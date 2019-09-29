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
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-reader-writer/pkg/io"
	"github.com/spatialcurrent/go-reader-writer/pkg/splitter"
	"github.com/spatialcurrent/go-stringify/pkg/stringify"
	rerrors "github.com/spatialcurrent/railgun/pkg/errors"
	"github.com/spatialcurrent/railgun/pkg/middleware"
	"github.com/spatialcurrent/railgun/pkg/request"
	"github.com/spatialcurrent/railgun/pkg/util"
)

type ServiceDownloadHandler struct {
	*BaseHandler
	Cache *gocache.Cache
}

func (h *ServiceDownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	qs := request.NewQueryString(r)

	h.SendDebug("QueryString:" + fmt.Sprint(r.URL.Query()))

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
		h.Catalog.RLock()
		defer once.Do(func() { h.Catalog.RUnlock() })
		h.SendDebug("read locked for " + r.URL.String())
		filename, obj, err := h.Get(w, r.WithContext(ctx), format, vars, qs)
		once.Do(func() { h.Catalog.RUnlock() })
		h.SendDebug("read unlocked for " + r.URL.String())
		if err != nil {
			h.SendError(err)
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		} else {
			err = h.RespondWithObject(&Response{
				Writer:     w,
				StatusCode: http.StatusOK,
				Format:     format,
				Filename:   filename,
				Object:     obj,
			})
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

func (h *ServiceDownloadHandler) Get(w http.ResponseWriter, r *http.Request, format string, vars map[string]string, qs request.QueryString) (filename string, object interface{}, err error) {

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

	funcs := dfl.NewFuntionMapWithDefaults()

	for _, fn := range h.Catalog.ListFunctions() {
		for _, alias := range fn.Aliases {
			funcs[alias] = func(fn dfl.Node) func(funcs dfl.FunctionMap, vars map[string]interface{}, ctx interface{}, args []interface{}, quotes []string) (interface{}, error) {
				return func(funcs dfl.FunctionMap, vars map[string]interface{}, ctx interface{}, args []interface{}, quotes []string) (interface{}, error) {
					_, out, err := fn.Evaluate(vars, args, funcs, dfl.DefaultQuotes)
					if err != nil {
						return dfl.Null{}, errors.Wrap(err, "invalid arguments")
					}
					return out, nil
				}
			}(fn.Node)
		}
	}

	requestVars := map[string]interface{}{}
	if service.Transform != nil {
		_, newVariables, err := dfl.EvaluateMap(service.Transform, service.DataStore.Vars, r.URL.Query(), funcs, dfl.DefaultQuotes)
		if err != nil {
			return "", nil, errors.Wrap(err, "invalid service transform")
		}
		newVariables, err = stringify.StringifyMapKeys(newVariables, stringify.NewDefaultStringer())
		if err != nil {
			return "", nil, errors.Wrap(err, "error stringifying map keys")
		}
		if m, ok := newVariables.(map[string]interface{}); ok {
			requestVars = m
		}
	}

	variables := h.AggregateMaps(
		h.GetServiceVariables(h.Cache, serviceName),
		service.Defaults,
		requestVars,
		service.DataStore.Vars,
	)

	h.SendDebug("Variables: " + fmt.Sprint(variables))

	_, inputUri, err := dfl.EvaluateString(service.DataStore.Uri, variables, map[string]interface{}{}, funcs, dfl.DefaultQuotes)
	if err != nil {
		return "", nil, errors.Wrap(err, "invalid data store uri")
	}
	inputScheme, inputPath := splitter.SplitUri(inputUri)

	cacheKeyDataStore := ""
	inputUris := make([]string, 0)
	lastModified := map[string]time.Time{}
	inputReaders := map[string]io.ByteReadCloser{}
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
					uri := fmt.Sprintf("s3://%s/%s", bucket, key)
					valid := true
					if service.DataStore.Filter != nil {
						_, valid, err = dfl.EvaluateBool(service.DataStore.Filter, variables, map[string]interface{}{"uri": uri}, funcs, dfl.DefaultQuotes)
						if err != nil {
							return "", nil, errors.Wrap(err, "error evaluating filter on uri")
						}
					}
					if valid {
						keys = append(keys, key)
						lastModified[uri] = *obj.LastModified
					}
				}
			}
		} else {
			key := inputPath[i+1:]
			headObjectOutput, err := s3Client.HeadObject(&s3.HeadObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(key),
			})
			if err != nil {
				return "", nil, errors.Wrap(err, "error heading S3 object")
			}
			uri := fmt.Sprintf("s3://%s/%s", bucket, key)
			valid := true
			if service.DataStore.Filter != nil {
				_, valid, err = dfl.EvaluateBool(service.DataStore.Filter, variables, map[string]interface{}{"uri": uri}, funcs, dfl.DefaultQuotes)
				if err != nil {
					return "", nil, errors.Wrap(err, "error evaluating filter on uri")
				}
			}
			if valid {
				keys = append(keys, key)
				lastModified[uri] = *headObjectOutput.LastModified
			}
		}

		fmt.Println("Keys:", keys)

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
			inputReader, err := grw.ReadFromFile(&grw.ReadFromFileInput{
				File:       inputFile,
				Alg:        service.DataStore.Compression,
				Dict:       grw.NoDict,
				BufferSize: grw.DefaultBufferSize,
			})
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
					b, err := grw.ReadAllAndClose(&grw.ReadAllAndCloseInput{
						Uri:        inputUri,
						Alg:        service.DataStore.Compression,
						Dict:       grw.NoDict,
						BufferSize: grw.DefaultBufferSize,
						S3Client:   s3Client,
					})
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
		funcs,
		dfl.DefaultQuotes)
	if err != nil {
		return "", nil, errors.Wrap(err, "error evaluating process with name "+service.Process.Name)
	}

	// Set the variables to the cache every time to bump the expiration
	h.SetServiceVariables(h.Cache, serviceName, variables)

	outputObject, err = stringify.StringifyMapKeys(outputObject, stringify.NewDefaultStringer())
	return outputFilename, outputObject, err

}
