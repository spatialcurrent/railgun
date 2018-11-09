// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	//"fmt"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-try-get/gtg"
	"github.com/spatialcurrent/railgun/railgun/parser"
	"github.com/spatialcurrent/railgun/railgun/util"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"net/http"
	//"reflect"
)

type ServicesExecHandler struct {
	*BaseHandler
}

func (h *ServicesExecHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)

	switch r.Method {
	case "POST":
		obj, err := h.Post(w, r, format)
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		} else {
			err = h.RespondWithObject(w, obj, format)
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

func (h *ServicesExecHandler) Post(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {

	obj, err := h.ParseBody(r, format)
	if err != nil {
		return nil, err
	}

	serviceName := gtg.TryGetString(obj, "service", "")
	if len(serviceName) == 0 {
		return nil, &rerrors.ErrMissingRequiredParameter{Name: "service"}
	}

	service, ok := h.Catalog.GetService(serviceName)
	if !ok {
		return nil, &rerrors.ErrMissingObject{Type: "service", Name: serviceName}
	}

	vars := map[string]interface{}{}
	for k, v := range service.Defaults {
		vars[k] = v
	}

	variables, err := parser.ParseMap(obj, "variables")
	if err != nil {
		return nil, &rerrors.ErrInvalidParameter{Name: "variables", Value: gtg.TryGetString(obj, "variables", "")}
	}
	for k, v := range variables {
		vars[k] = v
	}

	_, inputUri, err := dfl.EvaluateString(service.DataStore.Uri, vars, map[string]interface{}{}, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return nil, errors.Wrap(err, "invalid data store uri")
	}

	inputReader, _, err := grw.ReadFromResource(inputUri, service.DataStore.Compression, 4096, false, nil)
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

	inputObject, err := gss.DeserializeBytes(inputBytes, inputFormat, gss.NoHeader, "", false, gss.NoLimit, inputType, false)
	if err != nil {
		return nil, errors.Wrap(err, "error deserializing input using format "+inputFormat)
	}

	_, outputObject, err := service.Process.Node.Evaluate(vars, inputObject, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return nil, errors.Wrap(err, "error evaluating process with name "+service.Process.Name)
	}
	return outputObject, nil

}
