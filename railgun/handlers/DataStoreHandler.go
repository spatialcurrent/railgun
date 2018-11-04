package handlers

import (
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"net/http"
)

type DataStoreHandler struct {
	*BaseHandler
}

func (h *DataStoreHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	_, format, _ := railgun.SplitNameFormatCompression(r.URL.Path)

	switch r.Method {
	case "GET":
		b, err := h.Get(w, r, format)
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		}
		w.Write(b)
	case "DELETE":
		b, err := h.Delete(w, r, format)
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		}
		w.Write(b)
	default:
		err := h.RespondWithNotImplemented(w, format)
		if err != nil {
			panic(err)
		}
	}

}

func (h *DataStoreHandler) Get(w http.ResponseWriter, r *http.Request, format string) ([]byte, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return make([]byte, 0), errors.New("missing data store name")
	}
	ds, ok := h.Config.GetDataStore(name)
	if !ok {
		return make([]byte, 0), errors.New("no data store with name " + name)
	}

	data := map[string]interface{}{}
	data["datastore"] = ds.Map()

	b, err := gss.SerializeBytes(data, format, []string{}, gss.NoLimit)
	if err != nil {
		return b, errors.Wrap(err, "error serializing response body")
	}
	return b, nil
}

func (h *DataStoreHandler) Delete(w http.ResponseWriter, r *http.Request, format string) ([]byte, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return make([]byte, 0), errors.New("missing data store name")
	}
	ds, ok := h.Config.GetDataStore(name)
	if !ok {
		return make([]byte, 0), errors.New("no data store with name " + name)
	}

	err := h.Config.DeleteDataStore(name)
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error deleting data store")
	}

	err = h.Config.Save()
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error saving config")
	}

	data := map[string]interface{}{}
	data["success"] = true
	data["message"] = "data store with name " + ds.Name + " deleted."

	b, err := gss.SerializeBytes(data, format, []string{}, gss.NoLimit)
	if err != nil {
		return b, errors.Wrap(err, "error serializing response body")
	}
	return b, nil
}
