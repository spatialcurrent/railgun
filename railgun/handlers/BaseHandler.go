// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun/catalog"
	"github.com/spatialcurrent/railgun/railgun/request"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strings"
)

type BaseHandler struct {
	Viper           *viper.Viper
	Catalog         *catalog.RailgunCatalog
	Requests        chan request.Request
	Messages        chan interface{}
	Errors          chan error
	AwsSessionCache *gocache.Cache
}

func (h *BaseHandler) ParseBody(r *http.Request, format string) (interface{}, error) {
	inputBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	inputType, err := gss.GetType(inputBytes, format)
	if err != nil {
		return nil, err
	}

	inputObject, err := gss.DeserializeBytes(inputBytes, format, []string{}, "", false, gss.NoLimit, inputType, false)
	if err != nil {
		return nil, err
	}

	return inputObject, nil
}

func (h *BaseHandler) RespondWithObject(w http.ResponseWriter, obj interface{}, format string) error {
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
		w.Write([]byte(html))
		return nil
	}

	b, err := gss.SerializeBytes(obj, format, []string{}, gss.NoLimit)
	if err != nil {
		return errors.Wrap(err, "error serializing response body")
	}
	switch format {
	case "bson":
		w.Header().Set("Content-Type", "application/ubjson")
	case "json":
		w.Header().Set("Content-Type", "application/json")
	case "toml":
		w.Header().Set("Content-Type", "application/toml")
	case "yaml":
		w.Header().Set("Content-Type", "text/yaml")
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

	fmt.Println("Response Bytes:", string(b))

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
