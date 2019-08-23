// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package middleware

import (
	"context"
	"net/http"
	"time"
)

var RequestMiddleware = func() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			client := r.Header.Get("X-Forwarded-For")
			ctx := context.WithValue(r.Context(), "request", Request{
				Client:  client,
				Host:    r.Host,
				Url:     r.URL.String(),
				Method:  r.Method,
				Start:   &start,
				End:     nil,
				Subject: "",
				Handler: "",
				Error:   nil,
			})
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
