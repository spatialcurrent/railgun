// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"github.com/spatialcurrent/railgun/railgun/request"
	"github.com/spatialcurrent/railgun/railgun/util"
	"net/http"
)

type ObjectHandler struct {
	*BaseHandler
	Object interface{}
}

func (h *ObjectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	qs := request.NewQueryString(r)
	pretty, _ := qs.FirstBool("pretty")

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)

	switch r.Method {
	case "GET":
		err := h.RespondWithObject(&Response{
			Writer:     w,
			StatusCode: http.StatusOK,
			Format:     format,
			Filename:   "",
			Object:     h.Object,
			Pretty:     pretty,
		})
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		}
	default:
		err := h.RespondWithNotImplemented(w, format)
		if err != nil {
			panic(err)
		}
	}

}
