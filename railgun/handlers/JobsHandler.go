package handlers

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"net/http"
)

type JobsHandler struct {
	*BaseHandler
}

func (h *JobsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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
func (h *JobsHandler) Get(w http.ResponseWriter, r *http.Request, format string) ([]byte, error) {

	jobsList := h.Config.ListJobs()
	jobs := make([]map[string]interface{}, 0, len(jobsList))
	for i := 0; i < len(jobsList); i++ {
		jobs = append(jobs, jobsList[i].Map())
	}

	data := map[string]interface{}{}
	data["jobs"] = jobs

	b, err := gss.SerializeBytes(data, format, []string{}, gss.NoLimit)
	if err != nil {
		return b, errors.Wrap(err, "error serializing response body")
	}
	return b, nil

}

func (h *JobsHandler) Post(w http.ResponseWriter, r *http.Request, format string) ([]byte, error) {

	obj, err := h.ParseBody(r, format)
	if err != nil {
		return make([]byte, 0), err
	}

	j, err := h.Config.ParseJob(obj)
	if err != nil {
		return make([]byte, 0), err
	}

	err = h.Config.AddJob(j)
	if err != nil {
		return make([]byte, 0), err
	}

	err = h.Config.Save()
	if err != nil {
		return make([]byte, 0), err
	}

	b, err := gss.SerializeBytes(map[string]interface{}{"success":true, "object": j.Map()}, format, []string{}, gss.NoLimit)
	if err != nil {
		return b, errors.Wrap(err, "error serializing response body")
	}
	return b, nil

}
