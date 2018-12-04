package middleware

import (
	"context"
	"net/http"
	"sync"
)

var LogMiddleware = func(messages chan interface{}) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := &sync.Once{}
			ctx := r.Context()
			ctx = context.WithValue(ctx, "log", log)
			r = r.WithContext(ctx)
			defer func() {
				log.Do(func() {
					if v := ctx.Value("request"); v != nil {
						if x, ok := v.(Request); ok {
							messages <- x.Map()
						}
					}
				})
			}()
			h.ServeHTTP(w, r)
		})
	}
}
