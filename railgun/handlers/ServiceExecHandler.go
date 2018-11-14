// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	//"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-try-get/gtg"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/parser"
	"github.com/spatialcurrent/railgun/railgun/util"
	"net/http"
	"strings"
	//"reflect"
)

type ServiceExecHandler struct {
	*BaseHandler
}

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

	obj, err := h.ParseBody(r, format)
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{}
	for k, v := range service.Defaults {
		variables[k] = v
	}

	jobVariables, err := parser.ParseMap(obj, "variables")
	if err != nil {
		return nil, &rerrors.ErrInvalidParameter{Name: "variables", Value: gtg.TryGetString(obj, "variables", "")}
	}
	//jobVariablesValue := reflect.ValueOf(jobVariables)
	//for _, k := range jobVariablesValue.MapKeys() {
	//  variables[fmt.Sprint(k.Interface())] = jobVariablesValue.MapIndex(k).Interface()
	//}
	for k, v := range jobVariables {
		variables[k] = v
	}

	_, inputUri, err := dfl.EvaluateString(service.DataStore.Uri, variables, map[string]interface{}{}, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return nil, errors.Wrap(err, "invalid data store uri")
	}

	var s3_client *s3.S3
	if strings.HasPrefix(inputUri, "s3://") {
		client, err := h.GetAWSS3Client()
		if err != nil {
			return nil, errors.Wrap(err, "error connecting to AWS")
		}
		s3_client = client
	}

	inputReader, _, err := grw.ReadFromResource(inputUri, service.DataStore.Compression, 4096, false, s3_client)
	if err != nil {
		return nil, errors.Wrap(err, "error opening resource at uri "+inputUri)
	}

	inputBytes, err := inputReader.ReadAllAndClose()
	if err != nil {
		return nil, errors.Wrap(err, "error reading from resource at uri "+inputUri)
	}

	inputFormat := service.DataStore.Format

	inputType, err := gss.GetType(inputBytes, inputFormat)
	if err != nil {
		return nil, errors.Wrap(err, "error getting type for input")
	}

	inputObject, err := gss.DeserializeBytes(inputBytes, inputFormat, gss.NoHeader, gss.NoComment, false, gss.NoLimit, inputType, false)
	if err != nil {
		return nil, errors.Wrap(err, "error deserializing input using format "+inputFormat)
	}

	_, outputObject, err := service.Process.Node.Evaluate(variables, inputObject, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return nil, errors.Wrap(err, "error evaluating process with name "+service.Process.Name)
	}
	return outputObject, nil

}
