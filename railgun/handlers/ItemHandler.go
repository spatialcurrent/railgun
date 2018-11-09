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
	"github.com/spatialcurrent/railgun/railgun/util"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"net/http"
	"reflect"
)

type ItemHandler struct {
	*BaseHandler
	Singular string
	Plural   string
	Type     reflect.Type
}

func (h *ItemHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)

	switch r.Method {
	case "GET":
		h.Catalog.Lock()
		obj, err := h.Get(w, r, format)
		h.Catalog.Unlock()
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
	case "DELETE":
		h.Catalog.Lock()
		obj, err := h.Delete(w, r, format)
		h.Catalog.Unlock()
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

func (h *ItemHandler) Get(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return make([]byte, 0), &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	ws, ok := h.Catalog.GetItem(name, h.Type)
	if !ok {
		return make([]byte, 0), &rerrors.ErrMissingObject{Type: h.Singular, Name: name}
	}
	return map[string]interface{}{"item": ws.Map()}, nil
}

func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return nil, errors.Wrap(&rerrors.ErrMissingRequiredParameter{Name: "name"}, "error deleting "+h.Singular)
	}
	
	obj, ok := h.Catalog.GetItem(name, h.Type)
	if !ok {
		return nil, errors.Wrap(&rerrors.ErrMissingObject{Type: h.Singular, Name: name}, "error deleting "+h.Singular)
	}

	err := h.Catalog.DeleteItem(name, h.Type)
	if err != nil {
		return nil, errors.Wrap(err, "error deleting "+h.Singular)
	}
	
	catalogUri := h.Viper.GetString("catalog-uri")
	if len(catalogUri) > 0 {
		err = h.Catalog.SaveToFile(catalogUri)
		if err != nil {
			return nil, errors.Wrap(err, "error saving config")
		}
	}

	data := map[string]interface{}{}
	data["success"] = true
	data["message"] = h.Singular + " with name " + obj.GetName() + " deleted."
	return data, nil
}
