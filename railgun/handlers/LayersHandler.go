// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"github.com/spatialcurrent/railgun/railgun"
	"net/http"
)

type LayersHandler struct {
	*BaseHandler
}

func (h *LayersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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
	case "POST":
		obj, err := h.Post(w, r, format)
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

func (h *LayersHandler) Get(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {
	layersList := h.Config.ListLayers()
	layers := make([]map[string]interface{}, 0, len(layersList))
	for i := 0; i < len(layersList); i++ {
		layers = append(layers, layersList[i].Map())
	}
	return map[string]interface{}{"layers": layers}, nil
}

func (h *LayersHandler) Post(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {

	obj, err := h.ParseBody(r, format)
	if err != nil {
		return make([]byte, 0), err
	}

	l, err := h.Config.ParseLayer(obj)
	if err != nil {
		return make([]byte, 0), err
	}

	err = h.Config.AddLayer(l)
	if err != nil {
		return make([]byte, 0), err
	}

	err = h.Config.Save()
	if err != nil {
		return make([]byte, 0), err
	}

	return map[string]interface{}{"success": true, "object": l.Map()}, nil

}
