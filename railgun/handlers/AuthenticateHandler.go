// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-try-get/gtg"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/util"
	"net/http"
	"reflect"
)

type AuthenticateHandler struct {
	*BaseHandler
	Type reflect.Type
}

func (h *AuthenticateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)

	switch r.Method {
	case "POST":
		statusCode, obj, err := h.Post(w, r, format)
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		} else {
			err = h.RespondWithObject(w, statusCode, obj, format)
			if err != nil {
				h.Messages <- err
				err = h.RespondWithError(w, err, format)
				if err != nil {
					panic(err)
				}
			}
		}
	case "OPTIONS":
	default:
		err := h.RespondWithNotImplemented(w, format)
		if err != nil {
			panic(err)
		}
	}

}

func (h *AuthenticateHandler) Post(w http.ResponseWriter, r *http.Request, format string) (int, interface{}, error) {

	body, err := h.ParseBody(r, format)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}

	username := gtg.TryGetString(body, "username", "")
	if len(username) == 0 {
		return http.StatusBadRequest, nil, &rerrors.ErrMissingRequiredParameter{Name: "username"}
	}

	password := gtg.TryGetString(body, "password", "")
	if len(password) == 0 {
		return http.StatusBadRequest, nil, &rerrors.ErrMissingRequiredParameter{Name: "password"}
	}

	if rootPassword := h.Viper.GetString("root-password"); len(rootPassword) > 0 {
		if username == "root" {
			if password != rootPassword {
				obj := map[string]interface{}{
					"success":  false,
					"username": username,
					"message":  "error authenticating as " + username,
				}
				return http.StatusUnauthorized, obj, nil
			}
			token, err := h.NewAuthorization(r, username)
			if err != nil {
				obj := map[string]interface{}{
					"success":  false,
					"username": username,
					"message":  "error authenticating as " + username,
				}
				return http.StatusInternalServerError, obj, nil
			} else {
				obj := map[string]interface{}{
					"success":  true,
					"username": username,
					"message":  "authenticated as " + username,
					"token":    token,
				}
				return http.StatusOK, obj, nil
			}
		}
	}

	return http.StatusBadRequest, nil, errors.New("could not authenticate")
}
