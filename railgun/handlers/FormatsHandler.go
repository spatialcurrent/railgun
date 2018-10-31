package handlers

import (
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"net/http"
)

type FormatsHandler struct {
	*BaseHandler
}

func (h *FormatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	data := map[string]interface{}{}
	data["formats"] = gss.Formats

	_, format, _ := railgun.SplitNameFormatCompression(r.URL.Path)
	b, err := gss.SerializeBytes(data, format, []string{}, -1)
	if err != nil {
		h.Messages <- err
		return
	}
	w.Write(b)

}
