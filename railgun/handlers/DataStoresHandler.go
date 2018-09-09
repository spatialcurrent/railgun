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

type DataStoresHandler struct {
	*BaseHandler
}

func (h *DataStoresHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

func (h *DataStoresHandler) Get(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {
	datastoresList := h.Config.ListDataStores()
	datastores := make([]map[string]interface{}, 0, len(datastoresList))
	for i := 0; i < len(datastoresList); i++ {
		datastores = append(datastores, datastoresList[i].Map())
	}
	return map[string]interface{}{"datastores": datastores}, nil
}

func (h *DataStoresHandler) Post(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {

	obj, err := h.ParseBody(r, format)
	if err != nil {
		return make([]byte, 0), err
	}

	ds, err := h.Config.ParseDataStore(obj)
	if err != nil {
		return make([]byte, 0), err
	}

	err = h.Config.AddDataStore(ds)
	if err != nil {
		return make([]byte, 0), err
	}

	err = h.Config.Save()
	if err != nil {
		return make([]byte, 0), err
	}

	return map[string]interface{}{"success": true, "object": ds.Map()}, nil
}
