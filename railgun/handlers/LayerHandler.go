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

type LayerHandler struct {
	*BaseHandler
}

func (h *LayerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

func (h *LayerHandler) Get(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return nil, &railgunerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	l, ok := h.Config.GetLayer(name)
	if !ok {
		return nil, &railgunerrors.ErrMissingObject{Type: "layer", Name: name}
	}
	return map[string]interface{}{"layer": l.Map()}, nil
}

func (h *LayerHandler) Delete(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return nil, errors.Wrap(&railgunerrors.ErrMissingRequiredParameter{Name: "name"}, "error deleting layer")
	}
	l, ok := h.Config.GetLayer(name)
	if !ok {
		return nil, errors.Wrap(&railgunerrors.ErrMissingObject{Type: "layer", Name: name}, "error deleting layer")
	}

	err := h.Config.DeleteLayer(l.Name)
	if err != nil {
		return nil, errors.Wrap(err, "error deleting layer")
	}

	err = h.Config.Save()
	if err != nil {
		return nil, errors.Wrap(err, "error saving config")
	}

	data := map[string]interface{}{}
	data["success"] = true
	data["message"] = "layer with name " + l.Name + " deleted."
	return data, nil
}
