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

type WorkspaceHandler struct {
	*BaseHandler
}

func (h *WorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

func (h *WorkspaceHandler) Get(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return make([]byte, 0), &railgunerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	ws, ok := h.Config.GetWorkspace(name)
	if !ok {
		return make([]byte, 0), &railgunerrors.ErrMissingObject{Type: "workspace", Name: name}
	}
	return map[string]interface{}{"workspace": ws.Map()}, nil
}

func (h *WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return make([]byte, 0), errors.Wrap(&railgunerrors.ErrMissingRequiredParameter{Name: "name"}, "error deleting workspace")
	}
	ws, ok := h.Config.GetWorkspace(name)
	if !ok {
		return make([]byte, 0), errors.Wrap(&railgunerrors.ErrMissingObject{Type: "workspace", Name: name}, "error deleting workspace")
	}

	err := h.Config.DeleteWorkspace(name)
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error deleting workspace")
	}

	err = h.Config.Save()
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error saving config")
	}

	data := map[string]interface{}{}
	data["success"] = true
	data["message"] = "workspace with name " + ws.Name + " deleted."
	return data, nil
}
