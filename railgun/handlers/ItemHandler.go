// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

import (
	"github.com/aws/aws-sdk-go/service/s3"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/request"
	"github.com/spatialcurrent/railgun/railgun/util"
)

type ItemHandler struct {
	*BaseHandler
	Singular string
	Plural   string
	Type     reflect.Type
}

func (h *ItemHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	qs := request.NewQueryString(r)
	pretty, _ := qs.FirstBool("pretty")

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)

	switch r.Method {
	case "GET":
		once := &sync.Once{}
		h.Catalog.RLock()
		defer once.Do(func() { h.Catalog.RUnlock() })
		obj, err := h.Get(w, r, format)
		once.Do(func() { h.Catalog.RUnlock() })
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		} else {
			err = h.RespondWithObject(&Response{
				Url:        r.URL,
				Writer:     w,
				StatusCode: http.StatusOK,
				Format:     format,
				Filename:   "",
				Object:     obj,
				Pretty:     pretty,
			})
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
		h.Catalog.Lock()
		defer once.Do(func() { h.Catalog.Unlock() })
		obj, err := h.Post(w, r, format)
		once.Do(func() { h.Catalog.Unlock() })
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		} else {
			if format == "html" {
				// StatusSeeOther always uses method GET
				http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
			} else {
				err = h.RespondWithObject(&Response{
					Url:        r.URL,
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
		}
	case "DELETE":
		once := &sync.Once{}
		h.Catalog.Lock()
		defer once.Do(func() { h.Catalog.Unlock() })
		obj, err := h.Delete(w, r, format)
		once.Do(func() { h.Catalog.Unlock() })
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		} else {
			if format == "html" {
				// StatusSeeOther always uses method GET
				http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
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
	ctx := r.Context()
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return make([]byte, 0), &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	item, ok := h.Catalog.GetItem(name, h.Type)
	if !ok {
		return make([]byte, 0), &rerrors.ErrMissingObject{Type: h.Singular, Name: name}
	}
	obj := map[string]interface{}{
		"success": true,
		"item":    item.Map(ctx),
	}
	return obj, nil
}

func (h *ItemHandler) Post(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {

	ctx := r.Context()

	var claims *jwt.StandardClaims
	if v := ctx.Value("claims"); v != nil {
		if c, ok := v.(*jwt.StandardClaims); ok {
			claims = c
		}
	}

	if claims == nil {
		return nil, errors.New("not authorized")
	}

	if claims.Subject != "root" {
		return nil, errors.New("not authorized")
	}

	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return nil, errors.Wrap(&rerrors.ErrMissingRequiredParameter{Name: "name"}, "error updating "+h.Singular)
	}

	_, ok = h.Catalog.GetItem(name, h.Type)
	if !ok {
		return nil, errors.Wrap(&rerrors.ErrMissingObject{Type: h.Singular, Name: name}, "error updating "+h.Singular)
	}

	raw := make([]byte, 0)
	itemFormat := ""
	if format == "html" {
		r.ParseForm()
		raw = []byte(r.FormValue("item"))
		itemFormat = "json"
	} else {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, errors.Wrap(err, "error reading from request body")
		}
		raw = body
		itemFormat = format
	}

	obj, err := h.ParseBody(raw, itemFormat)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing body")
	}

	item, err := h.Catalog.ParseItem(obj, h.Type)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing item")
	}

	if item.GetName() != name {
		return nil, errors.New(fmt.Sprintf("the old name %s does not match the new name %s", name, item.GetName()))
	}

	err = h.Catalog.Update(item)
	if err != nil {
		return nil, errors.Wrap(err, "error updating "+h.Singular)
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
	data["message"] = h.Singular + " with name " + name + " updated."
	return data, nil
}

func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {

	ctx := r.Context()

	var claims *jwt.StandardClaims
	if v := ctx.Value("claims"); v != nil {
		if c, ok := v.(*jwt.StandardClaims); ok {
			claims = c
		}
	}

	if claims == nil {
		return nil, errors.New("not authorized")
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

	err := h.Catalog.DeleteItem(name, h.Type)
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
