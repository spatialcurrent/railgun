package handlers

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"net/http"
)

type ProcessesHandler struct {
	*BaseHandler
}

func (h *ProcessesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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
func (h *ProcessesHandler) Get(w http.ResponseWriter, r *http.Request, format string) ([]byte, error) {

	processesList := h.Config.ListProcesses()
	processes := make([]map[string]interface{}, 0, len(processesList))
	for i := 0; i < len(processesList); i++ {
		processes = append(processes, processesList[i].Map())
	}

	data := map[string]interface{}{}
	data["processes"] = processes

	b, err := gss.SerializeBytes(data, format, []string{}, gss.NoLimit)
	if err != nil {
		return b, errors.Wrap(err, "error serializing response body")
	}
	return b, nil

}

func (h *ProcessesHandler) Post(w http.ResponseWriter, r *http.Request, format string) ([]byte, error) {

	obj, err := h.ParseBody(r, format)
	if err != nil {
		return make([]byte, 0), err
	}

	p, err := h.Config.ParseProcess(obj)
	if err != nil {
		return make([]byte, 0), err
	}

	err = h.Config.AddProcess(p)
	if err != nil {
		return make([]byte, 0), err
	}

	err = h.Config.Save()
	if err != nil {
		return make([]byte, 0), err
	}

	b, err := gss.SerializeBytes(map[string]interface{}{"success":true, "object": p.Map()}, format, []string{}, gss.NoLimit)
	if err != nil {
		return b, errors.Wrap(err, "error serializing response body")
	}
	return b, nil

}
