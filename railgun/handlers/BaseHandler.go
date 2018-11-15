// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	jwt "github.com/dgrijalva/jwt-go"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun/catalog"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/request"
	"github.com/spatialcurrent/railgun/railgun/util"
	"github.com/spatialcurrent/viper"
	"net/http"
	"strings"
	"time"
)

type BaseHandler struct {
	Viper           *viper.Viper
	Catalog         *catalog.RailgunCatalog
	Requests        chan request.Request
	Messages        chan interface{}
	Errors          chan error
	AwsSessionCache *gocache.Cache
	PublicKey       *rsa.PublicKey
	PrivateKey      *rsa.PrivateKey
	SessionDuration time.Duration
	ValidMethods    []string
}

func (h *BaseHandler) GetAuthorization(r *http.Request) (string, error) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		return "", &rerrors.ErrMissingRequiredParameter{Name: "Authorization"}
	}

	parts := strings.Split(authorization, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", &rerrors.ErrInvalidParameter{Name: "Authorization", Value: authorization}
	}

	return parts[1], nil
}

func (h *BaseHandler) NewAuthorization(r *http.Request, user string) (string, error) {
	fmt.Println("SessionDuration:", h.SessionDuration)
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, &jwt.StandardClaims{
		Subject:   user,
		ExpiresAt: time.Now().Add(h.SessionDuration).Unix(),
	})
	str, err := token.SignedString(h.PrivateKey)
	if err != nil {
		return "", errors.Wrap(err, "error signing JWT token")
	}
	return str, nil
	//r.Header.Set("Authorization", "bearer "+str)
	//return nil
}

/*func (h *BaseHandler) VerifyAuthorization(authorization string) {
  parts := strings.Split(authorization, ".")
  return jwt.SigningMethodRS512.Verify(strings.Join(parts[0:2], "."), parts[2], h.PublicKey)
}*/

func (h *BaseHandler) ParseAuthorization(str string) (*jwt.StandardClaims, error) {
	parser := &jwt.Parser{
		ValidMethods: h.ValidMethods,
	}
	fmt.Println("Parser:", parser)
	token, err := parser.ParseWithClaims(str, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return h.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims.(*jwt.StandardClaims), nil
}

func (h *BaseHandler) GetAWSSessionId(awsAccessKeyId string, awsSessionToken string) string {

	if len(awsAccessKeyId) > 0 {
		if len(awsSessionToken) > 0 {
			return awsAccessKeyId + "\n" + awsSessionToken
		}
		return awsAccessKeyId
	}
	return "AWS"
}

func (h *BaseHandler) GetAWSS3Client() (*s3.S3, error) {

	awsAccessKeyId := h.Viper.GetString("aws-access-key-id")
	awsSessionToken := h.Viper.GetString("aws-session-token")

	awsSessionId := h.GetAWSSessionId(awsAccessKeyId, awsSessionToken)

	obj, found := h.AwsSessionCache.Get(awsSessionId)
	if found {
		return s3.New(obj.(*session.Session)), nil
	}

	awsDefaultRegion := h.Viper.GetString("aws-default-region")
	awsSecretAccessKey := h.Viper.GetString("aws-secret-access-key")

	awsSession, err := util.ConnectToAWS(awsAccessKeyId, awsSecretAccessKey, awsSessionToken, awsDefaultRegion)
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to AWS")
	}
	h.AwsSessionCache.Set(awsSessionId, awsSession, gocache.DefaultExpiration)
	return s3.New(awsSession), nil
}

func (h *BaseHandler) ParseBody(inputBytes []byte, format string) (interface{}, error) {

	inputType, err := gss.GetType(inputBytes, format)
	if err != nil {
		return nil, err
	}

	fmt.Println("Bytes:", string(inputBytes))

	inputObject, err := gss.DeserializeBytes(inputBytes, format, []string{}, "", false, gss.NoLimit, inputType, false)
	if err != nil {
		return nil, errors.Wrap(err, "error deserializing body")
	}

	return inputObject, nil
}

func (h *BaseHandler) RespondWithObject(w http.ResponseWriter, statusCode int, obj interface{}, format string) error {

	if format == "html" {
		code, err := json.MarshalIndent(obj, "", "    ")
		if err != nil {
			return errors.Wrap(err, "error serializing response body")
		}
		var head strings.Builder
		head.WriteString("<title>Railgun</title>")
		head.WriteString(`<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/css/bootstrap.min.css" integrity="sha384-MCw98/SFnGE8fJT3GXwEOngsV7Zt27NXFoaoApmYm81iuXoPkFOJwJ8ERdknLPMO" crossorigin="anonymous">`)
		head.WriteString(`<script src="https://code.jquery.com/jquery-3.3.1.slim.min.js" integrity="sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo" crossorigin="anonymous"></script>`)
		head.WriteString(`<script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.3/umd/popper.min.js" integrity="sha384-ZMP7rVo3mIykV+2+9J3UJ46jBk0WLaUAdn689aCwoqbBJiSnjAK/l8WvCWPIPm49" crossorigin="anonymous"></script>`)
		head.WriteString(`<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.1.3/js/bootstrap.min.js" integrity="sha384-ChfqqxuZUCnJSK3+MXmPNIyE6ZbWh2IMqE241rYiqJxyMiZ6OW/JmZQ5stwEULTy" crossorigin="anonymous"></script>`)
		head.WriteString("<style>")
		formatter := html.New(html.WithClasses())
		style := styles.Get("xcode")
		err = formatter.WriteCSS(&head, style)
		if err != nil {
			return errors.Wrap(err, "error writing chroma styles")
		}
		head.WriteString("pre.chroma { border:2px solid black; padding: 20px; }")
		head.WriteString("</style>")
		lexer := chroma.Coalesce(lexers.Get("json"))
		iterator, err := lexer.Tokenise(nil, string(code))
		if err != nil {
			return errors.Wrap(err, "error tokenizing source code")
		}
		var preview strings.Builder
		err = formatter.Format(&preview, style, iterator)
		if err != nil {
			return errors.Wrap(err, "error formatting preview")
		}
		html := `
    <html>
      <head>` + head.String() + `</head>
      <body>
        <div class="container">
          <div class="row"><div class="col-md-12 h2">Items</div></div>
          <hr>
          <div class="row">
            <div class="col-md-2">
              <h4>Actions</h4>
              <button type="submit" class="btn btn-block btn-primary">Update</button>
              <button type="submit" class="btn btn-block btn-danger">Delete</button>
            </div>
            <div class="col-sm-10">
              <h4>Preview</h4>
              ` + preview.String() + `
            </div>
          </div>
        </div>
      </body>
    </html>
   `
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(statusCode)
		w.Write([]byte(html))
		return nil
	}

	b, err := gss.SerializeBytes(obj, format, []string{}, gss.NoLimit)
	if err != nil {
		return errors.Wrap(err, "error serializing response body")
	}

	contentType := "text/plain; charset=utf-8"
	switch format {
	case "bson":
		contentType = "application/ubjson"
	case "json":
		contentType = "application/json"
	case "toml":
		contentType = "application/toml"
	case "yaml", "yml":
		contentType = "text/yaml"
	}

	w.Header().Set("Content-Type", contentType)
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}
	w.Write(b)
	return nil
}

func (h *BaseHandler) RespondWithError(w http.ResponseWriter, err error, format string) error {

	b, serr := gss.SerializeBytes(map[string]interface{}{"success": false, "error": err.Error()}, format, []string{}, gss.NoLimit)
	if serr != nil {
		return serr
	}

	switch errors.Cause(err).(type) {
	case *rerrors.ErrMissingRequiredParameter:
		w.WriteHeader(http.StatusBadRequest)
	case *rerrors.ErrMissingObject:
		w.WriteHeader(http.StatusNotFound)
	case *rerrors.ErrDependent:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Write(b)
	return nil
}

func (h *BaseHandler) RespondWithNotImplemented(w http.ResponseWriter, format string) error {
	b, err := gss.SerializeBytes(map[string]interface{}{"success": false, "error": "not implemented"}, format, []string{}, gss.NoLimit)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNotImplemented)
	w.Write(b)
	return nil
}
