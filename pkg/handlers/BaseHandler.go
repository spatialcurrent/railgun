// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
)

import (
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	jwt "github.com/dgrijalva/jwt-go"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/go-try-get/pkg/gtg"
	"github.com/spatialcurrent/railgun/pkg/catalog"
	rerrors "github.com/spatialcurrent/railgun/pkg/errors"
	"github.com/spatialcurrent/railgun/pkg/parser"
	"github.com/spatialcurrent/railgun/pkg/request"
	"github.com/spatialcurrent/railgun/pkg/util"
	"github.com/spatialcurrent/viper"
)

type BaseHandler struct {
	Viper           *viper.Viper
	Catalog         *catalog.RailgunCatalog
	Requests        chan request.Request
	Messages        chan interface{}
	Errors          chan interface{}
	AwsSessionCache *gocache.Cache
	PublicKey       *rsa.PublicKey
	PrivateKey      *rsa.PrivateKey
	SessionDuration time.Duration
	ValidMethods    []string
	Debug           bool
	GitBranch       string
	GitCommit       string
}

func (h *BaseHandler) SendDebug(message interface{}) {
	if h.Debug {
		h.Messages <- message
	}
}

func (h *BaseHandler) SendInfo(message interface{}) {
	h.Messages <- message
}

func (h *BaseHandler) SendWarn(message interface{}) {
	h.Errors <- message
}

func (h *BaseHandler) SendError(message interface{}) {
	h.Errors <- message
}

func (h *BaseHandler) BuildCacheKeyDataStore(datastore string, uri string, lastModified time.Time) string {
	return fmt.Sprintf("datastore=%s\nuri=%s\nlastmodified=%d", datastore, uri, lastModified.UnixNano())
}

func (h *BaseHandler) BuildCacheKeyServiceVariables(service string) string {
	return fmt.Sprintf("service=%s\nvariables", service)
}

func (h *BaseHandler) BuildCacheBucketObjects(bucket string) string {
	return fmt.Sprintf("bucket=%s\nobjects", bucket)
}

func (h *BaseHandler) GetServiceVariables(cache *gocache.Cache, service string) map[string]interface{} {
	if cacheVariables, found := cache.Get(h.BuildCacheKeyServiceVariables(service)); found {
		if m, ok := cacheVariables.(*map[string]interface{}); ok {
			return *m
		}
	}
	return map[string]interface{}{}
}

func (h *BaseHandler) SetServiceVariables(cache *gocache.Cache, service string, variables map[string]interface{}) {
	cache.Set(h.BuildCacheKeyServiceVariables(service), &variables, gocache.DefaultExpiration)
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

	inputObject, err := h.DeserializeBytes(inputBytes, format)
	if err != nil {
		return nil, errors.Wrap(err, "error deserializing body")
	}

	return inputObject, nil
}

/* #nosec */
func (h *BaseHandler) RespondWithObject(resp *Response) error {

	if resp.Format == "html" {
		code, err := json.MarshalIndent(gtg.TryGet(resp.Object, "item", resp.Object), "", "    ")
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
		head.WriteString("pre { border:2px solid black; padding: 20px; }")
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
		requestUrlPath := ""
		if resp.Url != nil {
			requestUrlPath = resp.Url.Path
		}
		html := `
    <html>
      <head>` + head.String() + `</head>
      <body>
        <div class="container">
          <div class="row"><div class="col-md-12 h2">Items</div></div>
          <hr>
          <div class="row">
            <div class="col-sm-2">
              <!--<h4>Actions</h4>
              <button type="submit" class="btn btn-block btn-danger">Delete</button>-->
            </div>
            <div class="col-sm-8">
						  <nav>
								<div class="nav nav-tabs" id="nav-tab" role="tablist" style="margin-bottom: 8px;">
								  <a class="nav-item nav-link active" id="nav-preview-tab" data-toggle="tab" href="#nav-preview" role="tab" aria-controls="nav-preview" aria-selected="true">Preview</a>
									<a class="nav-item nav-link" id="nav-edit-tab" data-toggle="tab" href="#nav-edit" role="tab" aria-controls="nav-edit" aria-selected="false">Edit</a>
								</div>
							</nav>
							<div class="tab-content" id="nav-tabContent">
							  <div class="tab-pane fade show active" id="nav-preview" role="tabpanel" aria-labelledby="nav-preview-tab">
									` + preview.String() + `
								</div>
							  <div class="tab-pane fade" id="nav-edit" role="tabpanel" aria-labelledby="nav-edit-tab">
							    <form action="` + requestUrlPath + `" method="post">
								    <pre id="code" contenteditable="true">` + string(code) + `</pre>
										<input id="item" type="hidden" name="item" value="">
									  <button type="submit" class="btn btn-block btn-primary" onclick="document.getElementById('item').value = document.getElementById('code').textContent">Update</button>
									</form>
								</div>
							</div>
            </div>
						<div class="col-sm-2">
            </div>
          </div>
        </div>
      </body>
    </html>
   `
		resp.Writer.Header().Set("Content-Type", "text/html")
		resp.Writer.WriteHeader(resp.StatusCode)
		resp.Writer.Write([]byte(html))
		return nil
	}

	b, err := gss.SerializeBytes(&gss.SerializeBytesInput{
		Object:            resp.Object,
		Format:            resp.Format,
		Header:            gss.NoHeader,
		Limit:             gss.NoLimit,
		Pretty:            resp.Pretty,
		LineSeparator:     "\n",
		KeyValueSeparator: "=",
	})
	if err != nil {
		return errors.Wrap(err, "error serializing response body")
	}

	contentType := ""
	switch resp.Format {
	case "bson":
		contentType = "application/ubjson"
	case "json":
		contentType = "application/json"
	case "toml":
		contentType = "application/toml"
	case "yaml", "yml":
		contentType = "text/yaml"
	default:
		contentType = "text/plain; charset=utf-8"
	}

	if len(resp.Filename) > 0 {
		resp.Writer.Header().Set("Content-Disposition", "attachment; filename="+resp.Filename)
	}

	resp.Writer.Header().Set("Content-Type", contentType)
	if resp.StatusCode != http.StatusOK {
		resp.Writer.WriteHeader(resp.StatusCode)
	}
	resp.Writer.Write(b)
	return nil
}

func (h *BaseHandler) RespondWithError(w http.ResponseWriter, err error, format string) error {

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

	if format == "html" {
		w.Write([]byte(err.Error()))
		return nil
	}

	b, serr := gss.SerializeBytes(&gss.SerializeBytesInput{
		Object:            map[string]interface{}{"success": false, "error": err.Error()},
		Format:            format,
		Header:            gss.NoHeader,
		Limit:             gss.NoLimit,
		Pretty:            false,
		LineSeparator:     "\n",
		KeyValueSeparator: "=",
	})
	if serr != nil {
		return serr
	}

	w.Write(b) // #nosec
	return nil
}

func (h *BaseHandler) RespondWithNotImplemented(w http.ResponseWriter, format string) error {
	if format == "html" {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Not implemented"))
		return nil
	}
	b, err := gss.SerializeBytes(&gss.SerializeBytesInput{
		Object:            map[string]interface{}{"success": false, "error": "not implemented"},
		Format:            format,
		Header:            gss.NoHeader,
		Limit:             gss.NoLimit,
		Pretty:            false,
		LineSeparator:     "\n",
		KeyValueSeparator: "=",
	})
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNotImplemented)
	w.Write(b) // #nosec
	return nil
}

func (h *BaseHandler) RespondWithBadRequest(w http.ResponseWriter, format string) error {
	if format == "html" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
		return nil
	}
	b, err := gss.SerializeBytes(&gss.SerializeBytesInput{
		Object:            map[string]interface{}{"success": false, "error": "Bad request"},
		Format:            format,
		Header:            gss.NoHeader,
		Limit:             gss.NoLimit,
		Pretty:            false,
		LineSeparator:     "\n",
		KeyValueSeparator: "=",
	})
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write(b) // #nosec
	return nil
}

func (h *BaseHandler) AggregateSlices(inputObjects []interface{}) []interface{} {
	outputSlice := reflect.ValueOf(make([]interface{}, 0))
	for _, inputObject := range inputObjects {
		if kind := reflect.TypeOf(inputObject).Kind(); !(kind == reflect.Array || kind == reflect.Slice) {
			continue
		}
		inputObjectValue := reflect.ValueOf(inputObject)
		inputObjectLength := inputObjectValue.Len()
		for i := 0; i < inputObjectLength; i++ {
			outputSlice = reflect.Append(outputSlice, inputObjectValue.Index(i))
		}
	}
	return outputSlice.Interface().([]interface{})
}

func (h *BaseHandler) AggregateMaps(inputMaps ...map[string]interface{}) map[string]interface{} {
	outputMap := map[string]interface{}{}
	for _, m := range inputMaps {
		for k, v := range m {
			outputMap[k] = v
		}
	}
	return outputMap
}

func (h *BaseHandler) ParseVariables(body []byte, format string) (map[string]interface{}, error) {
	if len(body) > 0 {
		obj, err := h.ParseBody(body, format)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing body")
		}

		variables, err := parser.ParseMap(obj, "variables")
		if err != nil {
			return nil, &rerrors.ErrInvalidParameter{Name: "variables", Value: gtg.TryGetString(obj, "variables", "")}
		}

		return variables, nil
	}
	return map[string]interface{}{}, nil
}

func (h *BaseHandler) DeserializeBytes(inputBytes []byte, inputFormat string) (interface{}, error) {

	object, err := gss.DeserializeBytes(&gss.DeserializeBytesInput{
		Bytes:         inputBytes,
		Format:        inputFormat,
		Header:        gss.NoHeader,
		Comment:       gss.NoComment,
		LazyQuotes:    false,
		SkipLines:     gss.NoSkip,
		Limit:         gss.NoLimit,
		LineSeparator: "\n",
	})
	if err != nil {
		return nil, errors.Wrap(err, "error deserializing input using format "+inputFormat)
	}

	return object, nil
}
