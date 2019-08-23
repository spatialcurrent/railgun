// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package middleware

import (
	//"fmt"
	"net/http"
)

import (
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
)

var RecoverMiddleware = func(logger *gsl.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					//fmt.Println("r:", r)
					logger.Error(r)
					logger.Flush()
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			h.ServeHTTP(w, r)
		})
	}
}
