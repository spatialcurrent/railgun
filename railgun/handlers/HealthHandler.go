// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"context"
	"net/http"
	"reflect"
	"sync"
	"time"
)

import (
	"github.com/gorilla/mux"
	//"github.com/spatialcurrent/railgun/railgun"
	"github.com/spatialcurrent/railgun/railgun/middleware"
	"github.com/spatialcurrent/railgun/railgun/util"
)

type HealthHandler struct {
	*BaseHandler
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	vars := mux.Vars(r)

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)

	if v := ctx.Value("request"); v != nil {
		if req, ok := v.(middleware.Request); ok {
			req.Vars = vars
			req.Handler = reflect.TypeOf(h).Elem().Name()
			ctx = context.WithValue(ctx, "request", req)
		}
	}

	defer func() {
		if v := ctx.Value("log"); v != nil {
			if log, ok := v.(*sync.Once); ok {
				log.Do(func() {
					if v := ctx.Value("request"); v != nil {
						if req, ok := v.(middleware.Request); ok {
							end := time.Now()
							req.End = &end
							h.SendInfo(req.Map())
						}
					}
				})
			}
		}
	}()

	switch r.Method {
	case "GET":
		obj := map[string]interface{}{
			"status":    "ok",
			"version":   h.Version,
			"gitBranch": h.GitBranch,
			"gitCommit": h.GitCommit,
		}
		err := h.RespondWithObject(&Response{
			Writer:     w,
			StatusCode: http.StatusOK,
			Format:     format,
			Filename:   "",
			Object:     obj,
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
