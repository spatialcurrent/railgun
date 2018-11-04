package handlers

import (
	gocache "github.com/patrickmn/go-cache"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
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

func (h *BaseHandler) RespondWithError(w http.ResponseWriter, err error, format string) error {
	b, err := gss.SerializeBytes(map[string]interface{}{"success": false, "error": err.Error()}, format, []string{}, gss.NoLimit)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusInternalServerError)
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
