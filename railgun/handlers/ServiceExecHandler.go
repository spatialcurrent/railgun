// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/snappy"
	"github.com/gorilla/mux"
	"github.com/mitchellh/go-homedir"
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

var cacheKeyDataStoreFormat = "%s/datastore/%d"
var cacheKeyServiceFormat = "%s/variables"

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
	var inputReader grw.ByteReadCloser
	var inputObject interface{}
	var s3_client *s3.S3
	if inputScheme == "s3" {

		s3_client, err = h.GetAWSS3Client()
		if err != nil {
			return nil, errors.Wrap(err, "error connecting to AWS")
		}
		//s3_client = client

		i := strings.Index(inputPath, "/")
		if i == -1 {
			return nil, errors.New("path missing bucket")
		}

		bucket := inputPath[0:i]
		key := inputPath[i+1:]

		headObjectOutput, err := s3_client.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return nil, errors.Wrap(err, "error heading S3 object")
		}

		cacheKeyDataStore = fmt.Sprintf(cacheKeyDataStoreFormat, service.DataStore.Name, headObjectOutput.LastModified.UnixNano())
		fmt.Println("cache key:", cacheKeyDataStore)

		if object, found := h.Cache.Get(cacheKeyDataStore); found {
			fmt.Println("cache hit for datastore with key " + cacheKeyDataStore)
			inputObject = object
		} else {
			fmt.Println("cache miss for datastore with key " + cacheKeyDataStore)
			inputReader, _, err = grw.ReadS3Object(bucket, key, service.DataStore.Compression, false, s3_client)
			if err != nil {
				return nil, errors.Wrap(err, "error opening resource at uri "+inputUri)
			}
		}

	} else if inputScheme == "" || inputScheme == "file" {
		inputPathExpanded, err := homedir.Expand(inputPath)
		if err != nil {
			return nil, errors.Wrap(err, "error expanding file at path "+inputPath)
		}

		inputFile, err := os.Open(inputPathExpanded)
		if err != nil {
			return nil, errors.Wrap(err, "error opening file at path "+inputPath)
		}

		inputFileInfo, err := inputFile.Stat()
		if err != nil {
			return nil, errors.Wrap(err, "error stating file at path "+inputPath)
		}

		cacheKeyDataStore = fmt.Sprintf(cacheKeyDataStoreFormat, service.DataStore.Name, inputFileInfo.ModTime().UnixNano())

		if object, found := h.Cache.Get(cacheKeyDataStore); found {
			fmt.Println("cache hit for datastore with key " + cacheKeyDataStore)
			inputObject = object
		} else {
			switch service.DataStore.Compression {
			case "snappy":
				inputReader = &grw.Reader{
					Reader: bufio.NewReader(snappy.NewReader(bufio.NewReader(inputFile))),
					File:   inputFile,
				}
			case "gzip":
				gr, err := gzip.NewReader(bufio.NewReader(inputFile))
				if err != nil {
					return nil, errors.Wrap(err, "error creating gzip reader for file \""+inputPath+"\"")
				}
				inputReader = &grw.Reader{
					Reader: bufio.NewReader(gr),
					Closer: gr,
					File:   inputFile,
				}
			case "bzip2":
				inputReader = &grw.Reader{
					Reader: bufio.NewReader(bzip2.NewReader(bufio.NewReader(inputFile))),
					File:   inputFile,
				}
			case "zip":
				inputReader, err = grw.ReadZipFile(inputPathExpanded, false)
				if err != nil {
					return nil, errors.Wrap(err, "error creating gzip reader for file \""+inputPath+"\"")
				}
			case "none", "":
				inputReader = &grw.Reader{Reader: bufio.NewReader(inputFile), File: inputFile}
			}
		}

	}

	if inputObject == nil {

		fmt.Println("Input Object is nil")

		if inputReader == nil {
			inputReader, _, err = grw.ReadFromResource(inputUri, service.DataStore.Compression, 4096, false, s3_client)
			if err != nil {
				return nil, errors.Wrap(err, "error opening resource at uri "+inputUri)
			}
		}

		fmt.Println("Input Reader:", inputReader)

		inputBytes, err := inputReader.ReadAllAndClose()
		if err != nil {
			return nil, errors.Wrap(err, "error reading from resource at uri "+inputUri)
		}

		fmt.Println("Input Bytes:", len(inputBytes))

		inputFormat := service.DataStore.Format

		inputType, err := gss.GetType(inputBytes, inputFormat)
		if err != nil {
			return nil, errors.Wrap(err, "error getting type for input")
		}

		fmt.Println("Deserializing")

		object, err := gss.DeserializeBytes(inputBytes, inputFormat, gss.NoHeader, gss.NoComment, false, gss.NoSkip, gss.NoLimit, inputType, false)
		if err != nil {
			return nil, errors.Wrap(err, "error deserializing input using format "+inputFormat)
		}

		inputObject = object

	}

	fmt.Println("saving to cache")

	if len(cacheKeyDataStore) > 0 {
		h.Cache.Set(cacheKeyDataStore, inputObject, gocache.DefaultExpiration)
	}

	fmt.Println("evaluating")

	variables, outputObject, err := service.Process.Node.Evaluate(variables, inputObject, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return nil, errors.Wrap(err, "error evaluating process with name "+service.Process.Name)
	}

	fmt.Println("saving variables")

	// Set the variables to the cache every time to bump the expiration
	h.Cache.Set(cacheKeyService, &variables, gocache.DefaultExpiration)

	fmt.Println("variables saved")

	return outputObject, nil

}
