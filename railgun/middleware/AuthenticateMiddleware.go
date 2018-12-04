package middleware

import (
	"context"
	"crypto/rsa"
	"net/http"
	"strings"
)

import (
	jwt "github.com/dgrijalva/jwt-go"
)

var AuthenticateMiddleware = func(validMethods []string, publicKey *rsa.PublicKey) func(h http.Handler) http.Handler {
	parser := &jwt.Parser{ValidMethods: validMethods}
	keyFunc := func(t *jwt.Token) (interface{}, error) { return publicKey, nil }
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if str := r.Header.Get("Authorization"); len(str) > 0 {
				parts := strings.Split(str, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					token, err := parser.ParseWithClaims(parts[1], &jwt.StandardClaims{}, keyFunc)
					if err == nil {
						claims := token.Claims.(*jwt.StandardClaims)
						ctx = context.WithValue(ctx, "claims", claims)
						if v := ctx.Value("request"); v != nil {
							if x, ok := v.(Request); ok {
								x.Subject = claims.Subject
								ctx = context.WithValue(ctx, "request", x)
							}
						}
					}
				}
			}
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
