// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
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
	"github.com/spatialcurrent/go-try-get/gtg"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/parser"
	"github.com/spatialcurrent/railgun/railgun/util"
)

type ServiceExecHandler struct {
	*BaseHandler
	Cache *gocache.Cache
}

var cacheKeyDataStoreFormat = "datastore=%s\nuri=%s\nlastmodified=%d"
var cacheKeyServiceFormat = "service=%s\nvariables"

func (h *ServiceExecHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)

	switch r.Method {
	case "POST":
		obj, err := h.Post(w, r, format, mux.Vars(r))
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		} else {
			err = h.RespondWithObject(w, http.StatusOK, obj, format)
			if err != nil {
				h.Messages <- err
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

func (h *ServiceExecHandler) Post(w http.ResponseWriter, r *http.Request, format string, vars map[string]string) (interface{}, error) {

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

	variables := map[string]interface{}{}

	cacheKeyService := fmt.Sprintf(cacheKeyServiceFormat, serviceName)

	// Load variables from cache
	if cacheVariables, found := h.Cache.Get(cacheKeyService); found {
		if m, ok := cacheVariables.(*map[string]interface{}); ok {
			h.Messages <- "variables found in cache " + cacheKeyService
			for k, v := range *m {
				variables[k] = v
			}
		}
	}

	// Load default variable values from service definition
	for k, v := range service.Defaults {
		variables[k] = v
	}

	// Load variables from request body
	if len(body) > 0 {
		obj, err := h.ParseBody(body, format)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing body")
		}

		jobVariables, err := parser.ParseMap(obj, "variables")
		if err != nil {
			return nil, &rerrors.ErrInvalidParameter{Name: "variables", Value: gtg.TryGetString(obj, "variables", "")}
		}

		for k, v := range jobVariables {
			variables[k] = v
		}
	}

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

			cacheKeyDataStore = fmt.Sprintf(
				cacheKeyDataStoreFormat,
				service.DataStore.Name,
				inputUri,
				lastModified[fmt.Sprintf("s3://%s/%s", bucket, key)].UnixNano())

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

		cacheKeyDataStore = fmt.Sprintf(
			cacheKeyDataStoreFormat,
			service.DataStore.Name,
			inputUri,
			lastModified[inputUri].UnixNano())

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

	wg, _ := errgroup.WithContext(context.Background())
	for i, inputUri := range inputUris {
		i, inputUri := i, inputUri // closure
		wg.Go(func() error {

			if inputObjects[i] == nil {

				inputBytes := make([]byte, 0)
				if r, ok := inputReaders[inputUri]; ok {
					b, err := r.ReadAllAndClose()
					if err != nil {
						return errors.Wrap(err, "error reading from resource at uri "+inputUri)
					}
					inputBytes = b
				} else {
					r, _, err = grw.ReadFromResource(inputUri, service.DataStore.Compression, 4096, false, s3Client)
					if err != nil {
						return errors.Wrap(err, "error opening resource at uri "+inputUri)
					}
					b, err := r.ReadAllAndClose()
					if err != nil {
						return errors.Wrap(err, "error reading from resource at uri "+inputUri)
					}
					inputBytes = b
				}

				//fmt.Println("* read all data for " + inputUri)

				inputFormat := service.DataStore.Format

				inputType, err := gss.GetType(inputBytes, inputFormat)
				if err != nil {
					return errors.Wrap(err, "error getting type for input")
				}

				//fmt.Println("* deserializing data for " + inputUri)

				object, err := gss.DeserializeBytes(inputBytes, inputFormat, gss.NoHeader, gss.NoComment, false, gss.NoSkip, gss.NoLimit, inputType, false)
				if err != nil {
					return errors.Wrap(err, "error deserializing input using format "+inputFormat)
				}

				inputObjects[i] = object

			}

			cacheKeyDataStore = fmt.Sprintf(
				cacheKeyDataStoreFormat,
				service.DataStore.Name,
				inputUri,
				lastModified[inputUri].UnixNano())

			h.Cache.Set(cacheKeyDataStore, inputObjects[i], gocache.DefaultExpiration)

			return nil
		})
	}

	// Wait until all objects have been loaded
	if err := wg.Wait(); err != nil {
		return nil, errors.Wrap(err, "error fetching data")
	}

	//fmt.Println("* aggregating")

	// Aggregate data into 1 slice
	inputSlice := reflect.ValueOf(make([]interface{}, 0))
	for i, inputObject := range inputObjects {
		if kind := reflect.TypeOf(inputObject).Kind(); !(kind == reflect.Array || kind == reflect.Slice) {
			return nil, errors.New("input object for uri " + inputUris[i] + " is not an array or slice")
		}
		inputObjectValue := reflect.ValueOf(inputObject)
		inputObjectLength := inputObjectValue.Len()
		for i := 0; i < inputObjectLength; i++ {
			inputSlice = reflect.Append(inputSlice, inputObjectValue.Index(i))
		}
	}

	//fmt.Println("* evaluating")

	variables, outputObject, err := service.Process.Node.Evaluate(variables, inputSlice.Interface(), dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return nil, errors.Wrap(err, "error evaluating process with name "+service.Process.Name)
	}

	//fmt.Println("* saving variables")

	// Set the variables to the cache every time to bump the expiration
	h.Cache.Set(cacheKeyService, &variables, gocache.DefaultExpiration)

	//fmt.Println("* variables saved")

	//fmt.Println("Output Object: ", outputObject)

	return outputObject, nil

}
