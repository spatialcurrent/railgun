// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"github.com/spatialcurrent/go-adaptive-functions/af"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"net/http"
)

type FunctionsHandler struct {
	*BaseHandler
}

func (h *FunctionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	functions := make([]map[string]interface{}, 0, len(af.Functions))
	for i := 0; i < len(af.Functions); i++ {
		functions = append(functions, af.Functions[i].Map())
	}

	data := map[string]interface{}{}
	data["functions"] = functions

	_, format, _ := railgun.SplitNameFormatCompression(r.URL.Path)
	b, err := gss.SerializeBytes(data, format, []string{}, -1)
	if err != nil {
		h.Messages <- err
		return
	}
	w.Write(b)

}
