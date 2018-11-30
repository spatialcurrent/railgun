// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	//"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/railgun/railgun/core"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/util"
)

type GroupHandler struct {
	*BaseHandler
	Type reflect.Type
}

func (h *GroupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)

	switch r.Method {
	case "GET":
		once := &sync.Once{}
		once.Do(func() { h.Catalog.ReadLock() })
		defer once.Do(func() { h.Catalog.ReadUnlock() })
		obj, err := h.Get(w, r, format)
		once.Do(func() { h.Catalog.ReadUnlock() })
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
	case "POST":
		once := &sync.Once{}
		once.Do(func() { h.Catalog.WriteLock() })
		defer once.Do(func() { h.Catalog.WriteUnlock() })
		obj, err := h.Post(w, r, format)
		once.Do(func() { h.Catalog.WriteUnlock() })
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

	authorization, err := h.GetAuthorization(r)
	if err != nil {
		return nil, err
	}

	claims, err := h.ParseAuthorization(authorization)
	if err != nil {
		return nil, errors.Wrap(err, "could not verify authorization")
	}

	if claims.Subject != "root" {
		return nil, errors.New("not authorized")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading from request body")
	}

	obj, err := h.ParseBody(body, format)
	if err != nil {
		return nil, err
	}

	item, err := h.Catalog.ParseItem(obj, h.Type)
	if err != nil {
		return nil, err
	}

	err = h.Catalog.Add(item)
	if err != nil {
		return nil, err
	}

	catalogUri := h.Viper.GetString("catalog-uri")
	if len(catalogUri) > 0 {

		var s3_client *s3.S3
		if strings.HasPrefix(catalogUri, "s3://") {
			client, err := h.GetAWSS3Client()
			if err != nil {
				return nil, errors.Wrap(err, "error connecting to AWS")
			}
			s3_client = client
		}

		err = h.Catalog.SaveToUri(catalogUri, s3_client)
		if err != nil {
			return nil, err
		}
	}

	if m, ok := item.(core.Mapper); ok {
		return map[string]interface{}{"success": true, "object": m.Map()}, nil
	}

	return nil, &rerrors.ErrInvalidType{Value: reflect.TypeOf(item), Type: reflect.TypeOf((*core.Mapper)(nil))}
}
