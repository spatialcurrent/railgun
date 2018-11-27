package router

import (
	"context"
	"net/http"
	"time"
)

var DebugMiddleware = func(messages chan interface{}) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()
			ctx := context.WithValue(r.Context(), "start", start)
			ctx = context.WithValue(ctx, "request", map[string]interface{}{
				"host":   r.Host,
				"url":    r.URL.String(),
				"method": r.Method,
				"time":   start.Format(time.RFC3339),
			})

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
