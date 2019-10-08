// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type LoginHandler struct {
	*BaseHandler
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		h.Get(w, r)
	case "POST":
		statusCode, token := h.Post(w, r)
		switch statusCode {
		case http.StatusOK:
			http.SetCookie(w, &http.Cookie{
				Name:    "session",
				Value:   token,
				Expires: time.Now().Add(time.Minute * 60),
			})
			w.Write([]byte("Success!"))
		case http.StatusUnauthorized:
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		case http.StatusBadRequest:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request"))
		case http.StatusInternalServerError:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Internal server error"))
		}
	case "OPTIONS":
	default:
		err := h.RespondWithNotImplemented(w, "html")
		if err != nil {
			panic(err)
		}
	}

}

func (h *LoginHandler) Get(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var claims *jwt.StandardClaims
	if v := ctx.Value("claims"); v != nil {
		if c, ok := v.(*jwt.StandardClaims); ok {
			claims = c
		}
	}

	var body strings.Builder
	if claims != nil && len(claims.Subject) > 0 {
		body.WriteString(fmt.Sprintf("Already logged in as <b>%s</b>.<br>", claims.Subject))
		body.WriteString(fmt.Sprintf("Session valid for <b>%s</b><br>", (time.Unix(claims.ExpiresAt, 0).Sub(time.Now()))))
	}
	body.WriteString("<form action=\"" + r.URL.Path + "\" method=\"post\">")
	body.WriteString("<label for=\"username\"><b>Username</b></label>")
	body.WriteString("<input type=\"text\" name=\"username\" placeholder=\"Enter username here\">")
	body.WriteString("<label for=\"password\"><b>Password</b></label>")
	body.WriteString("<input type=\"password\" name=\"password\" placeholder=\"Enter password here\">")
	if claims != nil && len(claims.Subject) > 0 {
		body.WriteString("<button type=\"submit\">Renew Session</button><br>")
		body.WriteString(fmt.Sprintf("You can logout at <a href=\"/logout.html\">logout.html</a>"))
	} else {
		body.WriteString("<button type=\"submit\">Log In</button>")
	}
	body.WriteString("</form>")
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<html><head><title>Railgun</title><style>input {display:block;margin-bottom:4px;}</style></head><body>" + body.String() + "</body></html>"))
}

func (h *LoginHandler) Post(w http.ResponseWriter, r *http.Request) (int, string) {

	r.ParseForm()

	username := r.FormValue("username")
	if len(username) == 0 {
		return http.StatusBadRequest, ""
	}

	password := r.FormValue("password")
	if len(password) == 0 {
		return http.StatusBadRequest, ""
	}

	if rootPassword := h.Viper.GetString("root-password"); len(rootPassword) > 0 {
		if username == "root" {
			if password != rootPassword {
				return http.StatusUnauthorized, ""
			}
			token, err := h.NewAuthorization(r, username)
			if err != nil {
				return http.StatusInternalServerError, ""
			} else {
				return http.StatusOK, token
			}
		}
	}

	return http.StatusInternalServerError, ""
}
