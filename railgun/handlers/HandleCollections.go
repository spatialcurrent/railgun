package handlers

import (
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spf13/viper"
	"net/http"
)

func HandleCollections(v *viper.Viper, w http.ResponseWriter, r *http.Request, vars map[string]string, qs railgun.QueryString, messages chan interface{}, collectionsList []railgun.Collection, collectionsByName map[string]railgun.Collection) {

	data := map[string]interface{}{}
	data["collections"] = collectionsList

	_, format, _ := railgun.SplitNameFormatCompression(r.URL.Path)
	str, err := gss.Serialize(data, format)
	if err != nil {
		messages <- err
		return
	}
	w.Write([]byte(str))

}
