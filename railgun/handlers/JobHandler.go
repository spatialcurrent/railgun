package handlers

import (
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"net/http"
)

type JobHandler struct {
	*BaseHandler
}

func (h *JobHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

func (h *JobHandler) Get(w http.ResponseWriter, r *http.Request, format string) ([]byte, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return make([]byte, 0), errors.New("missing job name")
	}
	j, ok := h.Config.GetJob(name)
	if !ok {
		return make([]byte, 0), errors.New("no job with name " + name)
	}

	data := map[string]interface{}{}
	data["job"] = j.Map()

	b, err := gss.SerializeBytes(data, format, []string{}, gss.NoLimit)
	if err != nil {
		return b, errors.Wrap(err, "error serializing response body")
	}
	return b, nil
}

func (h *JobHandler) Delete(w http.ResponseWriter, r *http.Request, format string) ([]byte, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return make([]byte, 0), errors.New("missing job name")
	}
	j, ok := h.Config.GetJob(name)
	if !ok {
		return make([]byte, 0), errors.New("no job with name " + name)
	}

	err := h.Config.DeleteJob(j.Name)
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error deleting job")
	}

	err = h.Config.Save()
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error saving config")
	}

	data := map[string]interface{}{}
	data["success"] = true
	data["message"] = "job with name " + j.Name + " deleted."

	b, err := gss.SerializeBytes(data, format, []string{}, gss.NoLimit)
	if err != nil {
		return b, errors.Wrap(err, "error serializing response body")
	}
	return b, nil
}
