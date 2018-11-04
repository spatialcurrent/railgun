package handlers

import (
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"net/http"
)

type WorkspaceHandler struct {
	*BaseHandler
}

func (h *WorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

func (h *WorkspaceHandler) Get(w http.ResponseWriter, r *http.Request, format string) ([]byte, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return make([]byte, 0), errors.New("missing workspace name")
	}
	ws, ok := h.Config.GetWorkspace(name)
	if !ok {
		return make([]byte, 0), errors.New("no workspace with name " + name)
	}

	data := map[string]interface{}{}
	data["workspace"] = ws.Map()

	b, err := gss.SerializeBytes(data, format, []string{}, gss.NoLimit)
	if err != nil {
		return b, errors.Wrap(err, "error serializing response body")
	}
	return b, nil
}

func (h *WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request, format string) ([]byte, error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return make([]byte, 0), errors.New("missing workspace name")
	}
	ws, ok := h.Config.GetWorkspace(name)
	if !ok {
		return make([]byte, 0), errors.New("no workspace with name " + name)
	}

	err := h.Config.DeleteWorkspace(name)
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error deleting workspace")
	}

	err = h.Config.Save()
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error saving config")
	}

	data := map[string]interface{}{}
	data["success"] = true
	data["message"] = "workspace with name " + ws.Name + " deleted."

	b, err := gss.SerializeBytes(data, format, []string{}, gss.NoLimit)
	if err != nil {
		return b, errors.Wrap(err, "error serializing response body")
	}
	return b, nil
}
