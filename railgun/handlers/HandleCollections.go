package handlers

import (
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spf13/viper"
	"net/http"
)

func HandleCollections(v *viper.Viper, w http.ResponseWriter, r *http.Request, vars map[string]string, qs railgun.QueryString, requests chan railgun.Request, messages chan interface{}, errors chan error, collectionsList []railgun.Collection, collectionsByName map[string]railgun.Collection) {

	data := map[string]interface{}{}
	data["collections"] = collectionsList

	_, format, _ := railgun.SplitNameFormatCompression(r.URL.Path)
	b, err := gss.SerializeBytes(data, format, []string{}, -1)
	if err != nil {
		messages <- err
		return
	}
	w.Write(b)

}
