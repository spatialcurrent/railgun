package handlers

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"net/http"
)

type WorkspacesHandler struct {
	*BaseHandler
}

func (h *WorkspacesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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
	case "POST":
		b, err := h.Post(w, r, format)
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

func (h *WorkspacesHandler) Get(w http.ResponseWriter, r *http.Request, format string) ([]byte, error) {

	workspacesList := h.Config.ListWorkspaces()
	workspaces := make([]map[string]interface{}, 0, len(workspacesList))
	for i := 0; i < len(workspacesList); i++ {
		workspaces = append(workspaces, workspacesList[i].Map())
	}

	data := map[string]interface{}{}
	data["workspaces"] = workspaces

	b, err := gss.SerializeBytes(data, format, []string{}, gss.NoLimit)
	if err != nil {
		return b, errors.Wrap(err, "error serializing response body")
	}
	return b, nil
}

func (h *WorkspacesHandler) Post(w http.ResponseWriter, r *http.Request, format string) ([]byte, error) {

	obj, err := h.ParseBody(r, format)
	if err != nil {
		return make([]byte, 0), err
	}

	ws, err := h.Config.ParseWorkspace(obj)
	if err != nil {
		return make([]byte, 0), err
	}

	err = h.Config.AddWorkspace(ws)
	if err != nil {
		return make([]byte, 0), err
	}

	err = h.Config.Save()
	if err != nil {
		return make([]byte, 0), err
	}

	b, err := gss.SerializeBytes(map[string]interface{}{"success":true, "object": ws.Map()}, format, []string{}, gss.NoLimit)
	if err != nil {
		return b, errors.Wrap(err, "error serializing response body")
	}

	return b, nil

}
