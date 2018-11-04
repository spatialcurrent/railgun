package handlers

import (
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"net/http"
)

type LayersHandler struct {
	*BaseHandler
}

func (h *LayersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	layersList := h.Config.ListWorkspaces()
	layers := make([]map[string]interface{}, 0, len(layersList))
	for i := 0; i < len(layersList); i++ {
		layers = append(layers, layersList[i].Map())
	}

	data := map[string]interface{}{}
	data["layers"] = layers

	_, format, _ := railgun.SplitNameFormatCompression(r.URL.Path)
	b, err := gss.SerializeBytes(data, format, []string{}, -1)
	if err != nil {
		h.Messages <- err
		return
	}
	w.Write(b)

}
