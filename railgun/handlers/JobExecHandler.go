// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	//"fmt"
	"context"
	"net/http"
	"reflect"
	"sync"
	"time"
)

import (
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/middleware"
	"github.com/spatialcurrent/railgun/railgun/util"
)

type JobExecHandler struct {
	*BaseHandler
}

func (h *JobExecHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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
	case "POST":
		once := &sync.Once{}
		h.Catalog.RLock()
		defer once.Do(func() { h.Catalog.RUnlock() })
		h.SendDebug("read locked for " + r.URL.String())
		obj, err := h.Post(w, r.WithContext(ctx), format, vars)
		once.Do(func() { h.Catalog.RUnlock() })
		if err != nil {
			h.Messages <- err
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

func (h *JobExecHandler) Post(w http.ResponseWriter, r *http.Request, format string, vars map[string]string) (object interface{}, err error) {

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

	inputObject, err := h.DeserializeBytes(inputBytes, inputFormat, inputType)
	if err != nil {
		return nil, errors.Wrap(err, "error deserializing input using format "+inputFormat)
	}

	_, outputObject, err := job.Service.Process.Node.Evaluate(variables, inputObject, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return nil, errors.Wrap(err, "error evaluating process with name "+job.Service.Process.Name)
	}
	return outputObject, nil

}
