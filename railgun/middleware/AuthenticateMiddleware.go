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

			raw := ""
			if str := r.Header.Get("Authorization"); len(str) > 0 {
				if parts := strings.Split(str, " "); len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					raw = parts[1]
				}
			} else if cookie, err := r.Cookie("session"); err == nil {
				raw = cookie.Value
			}

			if len(raw) > 0 {
				token, err := parser.ParseWithClaims(raw, &jwt.StandardClaims{}, keyFunc)
				if err == nil {
					claims := token.Claims.(*jwt.StandardClaims)
					if claims.Valid() == nil {
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
