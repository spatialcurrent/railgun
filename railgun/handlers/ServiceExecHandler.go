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
	"io/ioutil"
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
	//"github.com/spatialcurrent/go-try-get/gtg"
	"github.com/spatialcurrent/go-adaptive-functions/af"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/middleware"
	"github.com/spatialcurrent/railgun/railgun/request"
	"github.com/spatialcurrent/railgun/railgun/util"
)

type ServiceExecHandler struct {
	*BaseHandler
	Cache *gocache.Cache
}

func (h *ServiceExecHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	qs := request.NewQueryString(r)

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
	case "POST":
		once := &sync.Once{}
		h.Catalog.RLock()
		defer once.Do(func() { h.Catalog.RUnlock() })
		h.SendDebug("read locked for " + r.URL.String())
		obj, err := h.Post(w, r.WithContext(ctx), format, vars, qs)
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
				Filename:   "",
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
	case "OPTIONS":
	default:
		err := h.RespondWithNotImplemented(w, format)
		if err != nil {
			panic(err)
		}
	}

}

func (h *ServiceExecHandler) Post(w http.ResponseWriter, r *http.Request, format string, vars map[string]string, qs request.QueryString) (object interface{}, err error) {

	ctx := r.Context()

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
		return nil, &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}

	service, ok := h.Catalog.GetService(serviceName)
	if !ok {
		return nil, &rerrors.ErrMissingObject{Type: "service", Name: serviceName}
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading from request body")
	}

	jobVariables, err := h.ParseVariables(body, format)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing variables from body")
	}

	bbox := make([]float64, 0)
	if str, err := qs.FirstString("bbox"); err != nil && len(str) > 0 {
		floats, err := af.ToFloat64Array.ValidateRun([]interface{}{strings.Split(str, ",")})
		if err != nil {
			return nil, errors.Wrap(err, "invalid bounding box parameter")
		}
		bbox = floats.([]float64)
	}

	variables := h.AggregateMaps(
		h.GetServiceVariables(h.Cache, serviceName),
		service.Defaults,
		jobVariables,
		map[string]interface{}{
			"bbox": bbox,
		},
		service.DataStore.Vars,
	)

	_, inputUri, err := dfl.EvaluateString(service.DataStore.Uri, variables, map[string]interface{}{}, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return nil, errors.Wrap(err, "invalid data store uri")
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
			return nil, errors.Wrap(err, "error connecting to AWS")
		}

		i := strings.Index(inputPath, "/")
		if i == -1 {
			return nil, errors.New("path missing bucket")
		}

		bucket := inputPath[0:i]

		keys := make([]string, 0)
		if j := strings.Index(inputPath, "*"); j >= 0 {
			listObjectsOutput, err := s3Client.ListObjects(&s3.ListObjectsInput{
				Bucket: aws.String(bucket),
				Prefix: aws.String(inputPath[i+1 : j]),
			})
			if err != nil {
				return nil, errors.Wrap(err, "could not list objects for path "+inputPath)
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
				return nil, errors.Wrap(err, "error heading S3 object")
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

			//fmt.Println("* checking cache with key\n:", cacheKeyDataStore)

			if object, found := h.Cache.Get(cacheKeyDataStore); found {
				//fmt.Println("* cache hit for datastore with key:\n" + cacheKeyDataStore)
				inputObjects = append(inputObjects, object)
			} else {
				//fmt.Println("* cache miss for datastore with key:\n" + cacheKeyDataStore)
				inputObjects = append(inputObjects, nil)
			}

		}

	} else if inputScheme == "" || inputScheme == "file" {

		inputUris = append(inputUris, inputUri)

		inputFile, inputFileInfo, err := grw.ExpandOpenAndStat(inputPath)
		if err != nil {
			return nil, err
		}

		lastModified[inputUri] = inputFileInfo.ModTime()

		cacheKeyDataStore = h.BuildCacheKeyDataStore(
			service.DataStore.Name,
			inputUri,
			lastModified[inputUri])

		if object, found := h.Cache.Get(cacheKeyDataStore); found {
			//fmt.Println("cache hit for datastore with key " + cacheKeyDataStore)
			inputObjects = append(inputObjects, object)
		} else {
			inputReader, err := grw.ReadFromFile(inputFile, service.DataStore.Compression, false, 4096)
			if err != nil {
				return nil, errors.Wrap(err, "error creating grw.ByteReadCloser for file at path \""+inputPath+"\"")
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

				inputType, err := gss.GetType(inputBytes, service.DataStore.Format)
				if err != nil {
					return errors.Wrap(err, "error getting type")
				}

				object, err := h.DeserializeBytes(inputBytes, service.DataStore.Format, inputType)
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
		return nil, errors.Wrap(err, "error fetching data")
	}

	//fmt.Println("* aggregating")

	//fmt.Println("* evaluating")

	variables, outputObject, err := service.Process.Node.Evaluate(
		variables,
		h.AggregateSlices(inputObjects),
		dfl.DefaultFunctionMap,
		dfl.DefaultQuotes)
	if err != nil {
		return nil, errors.Wrap(err, "error evaluating process with name "+service.Process.Name)
	}

	//fmt.Println("* saving variables")

	// Set the variables to the cache every time to bump the expiration
	h.SetServiceVariables(h.Cache, serviceName, variables)

	//fmt.Println("* variables saved")

	//fmt.Println("Output Object: ", outputObject)

	return gss.StringifyMapKeys(outputObject), nil

}
