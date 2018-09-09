// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spatialcurrent/railgun/railgun/railgunerrors"
	"net/http"
)

type JobHandler struct {
	*BaseHandler
}

func (h *JobHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	_, format, _ := railgun.SplitNameFormatCompression(r.URL.Path)

	switch r.Method {
	case "GET":
		obj, err := h.Get(w, r, format)
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		}
		err = h.RespondWithObject(w, obj, format)
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		}
	case "DELETE":
		obj, err := h.Delete(w, r, format)
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		}
		err = h.RespondWithObject(w, obj, format)
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		}
	default:
		err := h.RespondWithNotImplemented(w, format)
		if err != nil {
			panic(err)
		}
	}

}

func (h *JobHandler) Get(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return make([]byte, 0), &railgunerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	j, ok := h.Config.GetJob(name)
	if !ok {
		return nil, &railgunerrors.ErrMissingObject{Type: "job", Name: name}
	}
	return map[string]interface{}{"job": j.Map()}, nil
}

func (h *JobHandler) Delete(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return make([]byte, 0), errors.Wrap(&railgunerrors.ErrMissingRequiredParameter{Name: "name"}, "error deleting job")
	}
	j, ok := h.Config.GetJob(name)
	if !ok {
		return make([]byte, 0), errors.Wrap(&railgunerrors.ErrMissingObject{Type: "job", Name: name}, "error deleting job")
	}

	err := h.Config.DeleteJob(j.Name)
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error deleting job")
	}

	err = h.Config.Save()
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error saving config")
	}

	data := map[string]interface{}{}
	data["success"] = true
	data["message"] = "job with name " + j.Name + " deleted."
	return data, nil
}
