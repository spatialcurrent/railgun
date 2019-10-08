// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package router

import (
	"compress/gzip"
	"crypto/rsa"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"

	gocache "github.com/patrickmn/go-cache"

	"github.com/spatialcurrent/go-adaptive-functions/pkg/af"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
	"github.com/spatialcurrent/railgun/pkg/catalog"
	"github.com/spatialcurrent/railgun/pkg/core"
	"github.com/spatialcurrent/railgun/pkg/handlers"
	"github.com/spatialcurrent/railgun/pkg/middleware"
	"github.com/spatialcurrent/railgun/pkg/request"
	"github.com/spatialcurrent/viper"
)

type RailgunRouter struct {
	*Router
	Viper           *viper.Viper
	Catalog         *catalog.RailgunCatalog
	PublicKey       *rsa.PublicKey
	PrivateKey      *rsa.PrivateKey
	ValidMethods    []string
	SessionDuration time.Duration
	Debug           bool
	GitBranch       string
	GitCommit       string
}

type NewRailgunRouterInput struct {
	Viper           *viper.Viper
	RailgunCatalog  *catalog.RailgunCatalog
	Requests        chan request.Request
	Messages        chan interface{}
	ErrorsChannel   chan interface{}
	AwsSessionCache *gocache.Cache
	PublicKey       *rsa.PublicKey
	PrivateKey      *rsa.PrivateKey
	ValidMethods    []string
	GitBranch       string
	GitCommit       string
	Logger          *gsl.Logger
}

func NewRailgunRouter(input *NewRailgunRouterInput) *RailgunRouter {

	v := input.Viper
	messages := input.Messages

	r := &RailgunRouter{
		Viper:           input.Viper,
		Catalog:         input.RailgunCatalog,
		Router:          NewRouter(input.Requests, input.Messages, input.ErrorsChannel, input.AwsSessionCache),
		PublicKey:       input.PublicKey,
		PrivateKey:      input.PrivateKey,
		ValidMethods:    input.ValidMethods,
		SessionDuration: v.GetDuration("jwt-session-duration"),
		Debug:           v.GetBool("verbose"),
		GitBranch:       input.GitBranch,
		GitCommit:       input.GitCommit,
	}

	if v.GetBool("http-middleware-recover") {
		messages <- map[string]interface{}{"middleware": "recover", "loaded": true}
		r.Use(middleware.RecoverMiddleware(input.Logger))
	}

	messages <- map[string]interface{}{"middleware": "request", "loaded": true}
	r.Use(middleware.RequestMiddleware())

	messages <- map[string]interface{}{"middleware": "authenticate", "loaded": true}
	r.Use(middleware.AuthenticateMiddleware(input.ValidMethods, input.PublicKey))

	r.Use(middleware.LogMiddleware(input.Logger))

	if v.GetBool("http-middleware-gzip") {
		messages <- map[string]interface{}{"middleware": "gzip", "loaded": true}
		r.Use(gziphandler.MustNewGzipLevelHandler(gzip.DefaultCompression))
	}

	if v.GetBool("http-middleware-cors") {
		messages <- map[string]interface{}{"middleware": "cors", "loaded": true}
		r.Use(middleware.CorsMiddleware(v.GetString("cors-origin"), v.GetString("cors-credentials")))
	}

	homeFormat := `<html>
  <head>
    <meta name="description" content="this is an app for mapping things">
    <meta name="author" content="Spatial Current">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <meta charset="utf-8">
    <link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto:300,400,500">
    <link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons">
    <style>
      * {
				box-sizing: border-box;
			}
      html, body, #app {
        height: 100%%;
        margin: 0;
        overflow: hidden;
        width: 100%%;
      }
    </style>
    <script>
      window.API_URL = "%s";
      window.BASELAYER_URL = "%s";
    </script>
  </head>
  <body>
    <div id='app'></div>
  <script type="text/javascript" src="%s"></script></body>
</html>
	`

	r.AddHandlerFunc(
		"home",
		[]string{"GET"},
		[]string{"/", "/queries", "/queries/{query:.+}"},
		handlers.FormatHandlerFunc(
			homeFormat,
			r.Viper.GetString("coconut-api-url"),
			r.Viper.GetString("coconut-baselayer-url"),
			r.Viper.GetString("coconut-bundle-url"),
		),
	)

	r.AddHandler("swagger", []string{"GET"}, []string{"/swagger.{ext}"}, &handlers.SwaggerHandler{
		BaseHandler: r.NewBaseHandler(),
	})

	r.AddHandler("health", []string{"GET"}, []string{"/health.{ext}"}, &handlers.HealthHandler{
		BaseHandler: r.NewBaseHandler(),
	})

	r.AddHandler("login", []string{"GET", "POST"}, []string{"/login.html"}, &handlers.LoginHandler{
		BaseHandler: r.NewBaseHandler(),
	})

	r.AddHandler("logout", []string{"GET", "POST"}, []string{"/logout.html"}, &handlers.LogoutHandler{
		BaseHandler: r.NewBaseHandler(),
	})

	r.AddHandler("authenticate", []string{"POST"}, []string{"/authenticate.{ext}"}, &handlers.AuthenticateHandler{
		BaseHandler: r.NewBaseHandler(),
	})

	r.AddHandler("formats", []string{"GET"}, []string{"/gss/formats.{ext}"}, &handlers.ObjectHandler{
		Object:      map[string]interface{}{"formats": gss.Formats},
		BaseHandler: r.NewBaseHandler(),
	})

	functions := make([]map[string]interface{}, 0, len(af.Functions))
	for i := 0; i < len(af.Functions); i++ {
		functions = append(functions, af.Functions[i].Map())
	}

	r.AddHandler("functions", []string{"GET"}, []string{"/dfl/functions.{ext}"}, &handlers.ObjectHandler{
		Object:      map[string]interface{}{"functions": functions},
		BaseHandler: r.NewBaseHandler(),
	})

	routes := []struct {
		Singular string
		Plural   string
		Type     reflect.Type
	}{
		struct {
			Singular string
			Plural   string
			Type     reflect.Type
		}{Singular: "workspace", Plural: "workspaces", Type: core.WorkspaceType},
		struct {
			Singular string
			Plural   string
			Type     reflect.Type
		}{Singular: "data store", Plural: "data stores", Type: core.DataStoreType},
		struct {
			Singular string
			Plural   string
			Type     reflect.Type
		}{Singular: "layer", Plural: "layers", Type: core.LayerType},
		struct {
			Singular string
			Plural   string
			Type     reflect.Type
		}{Singular: "functions", Plural: "functions", Type: core.FunctionType},
		struct {
			Singular string
			Plural   string
			Type     reflect.Type
		}{Singular: "process", Plural: "processes", Type: core.ProcessType},
		struct {
			Singular string
			Plural   string
			Type     reflect.Type
		}{Singular: "service", Plural: "services", Type: core.ServiceType},
		struct {
			Singular string
			Plural   string
			Type     reflect.Type
		}{Singular: "job", Plural: "jobs", Type: core.JobType},
		struct {
			Singular string
			Plural   string
			Type     reflect.Type
		}{Singular: "workflow", Plural: "workflows", Type: core.WorkflowType},
	}

	for _, route := range routes {

		r.AddHandler(
			strings.ToLower(strings.Replace(route.Plural, " ", "", -1)),
			[]string{"GET", "POST", "PUT", "OPTIONS"},
			[]string{fmt.Sprintf("/%s.{ext}", strings.ToLower(strings.Replace(route.Plural, " ", "", -1)))},
			&handlers.GroupHandler{
				Type:        route.Type,
				BaseHandler: r.NewBaseHandler(),
			},
		)

		r.AddHandler(
			strings.ToLower(strings.Replace(route.Singular, " ", "", -1)),
			[]string{"GET", "POST", "OPTIONS", "DELETE"},
			[]string{fmt.Sprintf("/%s/{name}.{ext}", strings.ToLower(strings.Replace(route.Plural, " ", "", -1)))},
			&handlers.ItemHandler{
				Singular:    route.Singular,
				Plural:      route.Plural,
				Type:        route.Type,
				BaseHandler: r.NewBaseHandler(),
			},
		)

	}

	cache := gocache.New(5*time.Minute, 10*time.Minute)

	r.AddHandler("service_exec", []string{"POST", "OPTIONS"}, []string{"/services/{name}/exec.{ext}"}, &handlers.ServiceExecHandler{
		BaseHandler: r.NewBaseHandler(),
		Cache:       cache,
	})

	r.AddHandler("service_download", []string{"GET"}, []string{"/services/{name}/download.{ext}"}, &handlers.ServiceDownloadHandler{
		BaseHandler: r.NewBaseHandler(),
		Cache:       cache,
	})

	r.AddHandler("service_tile", []string{"GET"}, []string{"/services/{name}/tiles/{z}/{x}/{y}.{ext}"}, &handlers.ServiceTileHandler{
		BaseHandler: r.NewBaseHandler(),
		Cache:       cache,
	})

	r.AddHandler("job_exec", []string{"POST", "OPTIONS"}, []string{"/jobs/{name}/exec.{ext}"}, &handlers.JobExecHandler{
		BaseHandler: r.NewBaseHandler(),
	})

	r.AddHandler("workflow_exec", []string{"POST", "OPTIONS"}, []string{"/workflows/{name}/exec.{ext}"}, &handlers.WorkflowExecHandler{
		BaseHandler: r.NewBaseHandler(),
	})

	r.AddHandler("layer_tile", []string{"GET"}, []string{"/layers/{name}/tiles/data/{z}/{x}/{y}.{ext}"}, &handlers.LayerTileHandler{
		BaseHandler: r.NewBaseHandler(),
	})

	r.AddHandler("layer_mask", []string{"GET"}, []string{"/layers/{name}/tiles/mask/{z}/{x}/{y}.{ext}"}, &handlers.LayerMaskHandler{
		BaseHandler: r.NewBaseHandler(),
	})

	return r
}

func (r *RailgunRouter) NewBaseHandler() *handlers.BaseHandler {
	return &handlers.BaseHandler{
		Viper:           r.Viper,
		Catalog:         r.Catalog,
		Requests:        r.Requests,
		Messages:        r.Messages,
		Errors:          r.Errors,
		AwsSessionCache: r.AwsSessionCache,
		PublicKey:       r.PublicKey,
		PrivateKey:      r.PrivateKey,
		ValidMethods:    r.ValidMethods,
		SessionDuration: r.SessionDuration,
		Debug:           r.Debug,
		GitBranch:       r.GitBranch,
		GitCommit:       r.GitCommit,
	}
}

func (r *RailgunRouter) AddHandler(name string, methods []string, paths []string, handler http.Handler) {
	for _, path := range paths {
		r.Methods(methods...).Name(name).Path(path).Handler(handler)
	}
}

func (r *RailgunRouter) AddHandlerFunc(name string, methods []string, paths []string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	for _, path := range paths {
		r.Methods(methods...).Name(name).Path(path).HandlerFunc(handlerFunc)
	}
}
