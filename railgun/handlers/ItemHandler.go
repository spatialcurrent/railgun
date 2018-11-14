// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/util"
	"net/http"
	"reflect"
	"strings"
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
			err = h.RespondWithObject(w, http.StatusOK, obj, format)
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
	obj := map[string]interface{}{
		"success": true,
		"item":    ws.Map(),
	}
	return obj, nil
}

func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {

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

	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return nil, errors.Wrap(&rerrors.ErrMissingRequiredParameter{Name: "name"}, "error deleting "+h.Singular)
	}

	obj, ok := h.Catalog.GetItem(name, h.Type)
	if !ok {
		return nil, errors.Wrap(&rerrors.ErrMissingObject{Type: h.Singular, Name: name}, "error deleting "+h.Singular)
	}

	err = h.Catalog.DeleteItem(name, h.Type)
	if err != nil {
		return nil, errors.Wrap(err, "error deleting "+h.Singular)
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
			return nil, errors.Wrap(err, "error saving config")
		}
	}

	data := map[string]interface{}{}
	data["success"] = true
	data["message"] = h.Singular + " with name " + obj.GetName() + " deleted."
	return data, nil
}
