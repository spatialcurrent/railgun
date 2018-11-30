package router

import (
	"context"
	//"fmt"
	"net/http"
	"sync"
	"time"
)

var DebugMiddleware = func(messages chan interface{}) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log := &sync.Once{}

			start := time.Now()
			ctx := context.WithValue(r.Context(), "start", start)
			ctx = context.WithValue(ctx, "log", log)
			ctx = context.WithValue(ctx, "request", map[string]interface{}{
				"host":   r.Host,
				"url":    r.URL.String(),
				"method": r.Method,
				"time":   start.Format(time.RFC3339),
			})
			r = r.WithContext(ctx)

			defer func() {
				log.Do(func() {
					m := map[string]interface{}{
						"request": ctx.Value("request"),
					}
					messages <- m
				})
			}()

			h.ServeHTTP(w, r)
		})
	}
}
