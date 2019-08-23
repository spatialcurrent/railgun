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
	"sync"
)

import (
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
)

var LogMiddleware = func(logger *gsl.Logger) func(h http.Handler) http.Handler {
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
							logger.Info(x.Map())
						}
					}
				})
			}()
			h.ServeHTTP(w, r)
		})
	}
}
