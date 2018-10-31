package handlers

import (
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"net/http"
)

type CollectionsHandler struct {
	*BaseHandler
	CollectionsList   []railgun.Collection
	CollectionsByName map[string]railgun.Collection
}

func (h *CollectionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	data := map[string]interface{}{}
	data["collections"] = h.CollectionsList

	_, format, _ := railgun.SplitNameFormatCompression(r.URL.Path)
	b, err := gss.SerializeBytes(data, format, []string{}, -1)
	if err != nil {
		h.Messages <- err
		return
	}
	w.Write(b)

}
