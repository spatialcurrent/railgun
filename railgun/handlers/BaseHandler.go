// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spatialcurrent/railgun/railgun/railgunerrors"
	"io/ioutil"
	"net/http"
)

type BaseHandler struct {
	Config          *railgun.Config
	Requests        chan railgun.Request
	Messages        chan interface{}
	Errors          chan error
	AwsSessionCache *gocache.Cache
	DflFuncs        dfl.FunctionMap
}

func (h *BaseHandler) ParseBody(r *http.Request, format string) (interface{}, error) {
	inputBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	inputType, err := gss.GetType(inputBytes, format)
	if err != nil {
		return nil, err
	}

	inputObject, err := gss.DeserializeBytes(inputBytes, format, []string{}, "", false, gss.NoLimit, inputType, false)
	if err != nil {
		return nil, err
	}

	return inputObject, nil
}

func (h *BaseHandler) RespondWithObject(w http.ResponseWriter, obj interface{}, format string) error {
	b, err := gss.SerializeBytes(obj, format, []string{}, gss.NoLimit)
	if err != nil {
		return errors.Wrap(err, "error serializing response body")
	}
	switch format {
	case "bson":
		w.Header().Set("Content-Type", "application/ubjson")
	case "json":
		w.Header().Set("Content-Type", "application/json")
	case "toml":
		w.Header().Set("Content-Type", "application/toml")
	case "yaml":
		w.Header().Set("Content-Type", "text/yaml")
	}
	w.Write(b)
	return nil
}

func (h *BaseHandler) RespondWithError(w http.ResponseWriter, err error, format string) error {

	b, serr := gss.SerializeBytes(map[string]interface{}{"success": false, "error": err.Error()}, format, []string{}, gss.NoLimit)
	if serr != nil {
		return serr
	}

	switch errors.Cause(err).(type) {
	case *railgunerrors.ErrMissingRequiredParameter:
		w.WriteHeader(http.StatusBadRequest)
	case *railgunerrors.ErrMissingObject:
		w.WriteHeader(http.StatusNotFound)
	case *railgunerrors.ErrDependent:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Write(b)
	return nil
}

func (h *BaseHandler) RespondWithNotImplemented(w http.ResponseWriter, format string) error {
	b, err := gss.SerializeBytes(map[string]interface{}{"success": false, "error": "not implemented"}, format, []string{}, gss.NoLimit)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNotImplemented)
	w.Write(b)
	return nil
}
