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
)

import (
	jwt "github.com/dgrijalva/jwt-go"
)

type LogoutHandler struct {
	*BaseHandler
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		h.Get(w, r)
	case "POST":
		h.Post(w, r)
	case "OPTIONS":
	default:
		err := h.RespondWithNotImplemented(w, "html")
		if err != nil {
			panic(err)
		}
	}

}

func (h *LogoutHandler) Get(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var claims *jwt.StandardClaims
	if v := ctx.Value("claims"); v != nil {
		if c, ok := v.(*jwt.StandardClaims); ok {
			claims = c
		}
	}

	var body strings.Builder
	if claims != nil && len(claims.Subject) > 0 {
		body.WriteString(fmt.Sprintf("Logged in as <b>%s</b>.<br>", claims.Subject))
		body.WriteString(fmt.Sprintf("Session valid for <b>%s</b><br>", (time.Unix(claims.ExpiresAt, 0).Sub(time.Now()))))
		body.WriteString("<form action=\"" + r.URL.Path + "\" method=\"post\">")
		body.WriteString("<button type=\"submit\">Log Out</button>")
		body.WriteString("</form>")
	} else {
		body.WriteString(fmt.Sprintf("You are not logged in."))
		body.WriteString(fmt.Sprintf("You can login at <a href=\"/login.html\">login.html</a>"))
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<html><head><title>Railgun</title><style>input {display:block;margin-bottom:4px;}</style></head><body>" + body.String() + "</body></html>"))
}

func (h *LogoutHandler) Post(w http.ResponseWriter, r *http.Request) {

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		MaxAge: -1,
	})
	w.Write([]byte("Success!"))

}
