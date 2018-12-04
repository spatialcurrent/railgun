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
			ctx := context.WithValue(r.Context(), "request", Request{
				Host:    r.Host,
				Url:     r.URL.String(),
				Method:  r.Method,
				Start:    &start,
				End: nil,
				Subject: "",
				Handler: "",
				Error: nil,
			})
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
