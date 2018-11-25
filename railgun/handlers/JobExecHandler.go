// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	//"fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/util"
	"net/http"
	//"reflect"
)

type JobExecHandler struct {
	*BaseHandler
}

func (h *JobExecHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

func (h *JobExecHandler) Post(w http.ResponseWriter, r *http.Request, format string, vars map[string]string) (interface{}, error) {

	jobName, ok := vars["name"]
	if !ok {
		return nil, &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}

	job, ok := h.Catalog.GetJob(jobName)
	if !ok {
		return nil, &rerrors.ErrMissingObject{Type: "job", Name: jobName}
	}

	variables := map[string]interface{}{}
	for k, v := range job.Service.Defaults {
		variables[k] = v
	}
	for k, v := range job.Variables {
		variables[k] = v
	}

	_, inputUri, err := dfl.EvaluateString(job.Service.DataStore.Uri, variables, map[string]interface{}{}, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return nil, errors.Wrap(err, "invalid data store uri")
	}

	inputReader, _, err := grw.ReadFromResource(inputUri, job.Service.DataStore.Compression, 4096, false, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error opening resource at uri "+inputUri)
	}

	inputBytes, err := inputReader.ReadAllAndClose()
	if err != nil {
		return nil, errors.Wrap(err, "error reading from resource at uri "+inputUri)
	}

	inputFormat := job.Service.DataStore.Format

	inputType, err := gss.GetType(inputBytes, inputFormat)
	if err != nil {
		return nil, errors.Wrap(err, "error getting type for input")
	}

	inputObject, err := gss.DeserializeBytes(inputBytes, inputFormat, []string{}, "", false, gss.NoSkip, gss.NoLimit, inputType, false)
	if err != nil {
		return nil, errors.Wrap(err, "error deserializing input using format "+inputFormat)
	}

	_, outputObject, err := job.Service.Process.Node.Evaluate(variables, inputObject, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return nil, errors.Wrap(err, "error evaluating process with name "+job.Service.Process.Name)
	}
	return outputObject, nil

}
