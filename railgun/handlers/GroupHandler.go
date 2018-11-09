// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"github.com/spatialcurrent/railgun/railgun/core"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/util"
	"net/http"
	"reflect"
)

type GroupHandler struct {
	*BaseHandler
	Type reflect.Type
}

func (h *GroupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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
	case "POST":
		h.Catalog.Lock()
		obj, err := h.Post(w, r, format)
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

func (h *GroupHandler) Get(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {
	list := reflect.ValueOf(h.Catalog.List(h.Type))
	size := list.Len()
	items := make([]map[string]interface{}, 0, size)
	for i := 0; i < size; i++ {
		obj := list.Index(i).Interface()
		if m, ok := obj.(core.Mapper); ok {
			items = append(items, m.Map())
		} else {
			return nil, &rerrors.ErrInvalidType{Value: reflect.TypeOf(obj), Type: reflect.TypeOf((*core.Mapper)(nil))}
		}
	}
	return map[string]interface{}{"items": items}, nil
}

func (h *GroupHandler) Post(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {

	body, err := h.ParseBody(r, format)
	if err != nil {
		return nil, err
	}

	obj, err := h.Catalog.ParseItem(body, h.Type)
	if err != nil {
		return nil, err
	}

	err = h.Catalog.Add(obj)
	if err != nil {
		return nil, err
	}

	catalogUri := h.Viper.GetString("catalog-uri")
	if len(catalogUri) > 0 {
		err = h.Catalog.SaveToFile(catalogUri)
		if err != nil {
			return nil, err
		}
	}

	if m, ok := obj.(core.Mapper); ok {
		return map[string]interface{}{"success": true, "object": m.Map()}, nil
	}

	return nil, &rerrors.ErrInvalidType{Value: reflect.TypeOf(obj), Type: reflect.TypeOf((*core.Mapper)(nil))}
}
