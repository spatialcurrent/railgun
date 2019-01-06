package router

import (
	"compress/gzip"
	"crypto/rsa"
	"fmt"
	"reflect"
	"strings"
	"time"
)

import (
	"github.com/NYTimes/gziphandler"
	gocache "github.com/patrickmn/go-cache"
	"github.com/spatialcurrent/go-adaptive-functions/af"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun/catalog"
	"github.com/spatialcurrent/railgun/railgun/core"
	"github.com/spatialcurrent/railgun/railgun/handlers"
	"github.com/spatialcurrent/railgun/railgun/middleware"
	"github.com/spatialcurrent/railgun/railgun/request"
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
	Version string
	GitBranch string
	GitCommit string
}

func NewRailgunRouter(v *viper.Viper, railgunCatalog *catalog.RailgunCatalog, requests chan request.Request, messages chan interface{}, errors chan interface{}, awsSessionCache *gocache.Cache, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey, validMethods []string, version string, gitBranch string, gitCommit string) *RailgunRouter {

	r := &RailgunRouter{
		Viper:           v,
		Catalog:         railgunCatalog,
		Router:          NewRouter(requests, messages, errors, awsSessionCache),
		PublicKey:       publicKey,
		PrivateKey:      privateKey,
		ValidMethods:    validMethods,
		SessionDuration: v.GetDuration("jwt-session-duration"),
		Debug:           v.GetBool("verbose"),
		Version: version,
		GitBranch: gitBranch,
		GitCommit: gitCommit,
	}

	if v.GetBool("http-middleware-recover") {
		messages <- map[string]interface{}{"middleware": "recover", "loaded": true}
		r.Use(middleware.RecoverMiddleware(errors))
	}

	messages <- map[string]interface{}{"middleware": "request", "loaded": true}
	r.Use(middleware.RequestMiddleware())

	messages <- map[string]interface{}{"middleware": "authenticate", "loaded": true}
	r.Use(middleware.AuthenticateMiddleware(validMethods, publicKey))

	r.Use(middleware.LogMiddleware(messages))

	if v.GetBool("http-middleware-gzip") {
		messages <- map[string]interface{}{"middleware": "gzip", "loaded": true}
		r.Use(gziphandler.MustNewGzipLevelHandler(gzip.DefaultCompression))
	}

	if v.GetBool("http-middleware-cors") {
		messages <- map[string]interface{}{"middleware": "cors", "loaded": true}
		r.Use(middleware.CorsMiddleware(v.GetString("cors-origin"), v.GetString("cors-credentials")))
	}

	r.AddHomeHandler("home", "/")

	r.AddSwaggerHandler("swagger", "/swagger.{ext}")

	r.AddHealthHandler("health", "/health.{ext}")

	r.AddAuthenticateHandler("authenticate", "/authenticate.{ext}")

	r.AddObjectHandler("formats", "/gss/formats.{ext}", map[string]interface{}{"formats": gss.Formats})

	functions := make([]map[string]interface{}, 0, len(af.Functions))
	for i := 0; i < len(af.Functions); i++ {
		functions = append(functions, af.Functions[i].Map())
	}

	r.AddObjectHandler("functions", "/dfl/functions.{ext}", map[string]interface{}{"functions": functions})

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

		r.AddGroupHandler(
			strings.ToLower(strings.Replace(route.Plural, " ", "", -1)),
			fmt.Sprintf("/%s.{ext}", strings.ToLower(strings.Replace(route.Plural, " ", "", -1))),
			route.Type,
		)

		r.AddItemHandler(
			strings.ToLower(strings.Replace(route.Singular, " ", "", -1)),
			fmt.Sprintf("/%s/{name}.{ext}", strings.ToLower(strings.Replace(route.Plural, " ", "", -1))),
			route.Type,
			route.Singular,
			route.Plural,
		)

	}

	r.AddServiceExecHandler("service_exec", "/services/{name}/exec.{ext}")
	
	r.AddServiceDownloadHandler("service_download", "/services/{name}/download.{ext}")

	r.AddServiceTileHandler("service_tile", "/services/{name}/tiles/{z}/{x}/{y}.{ext}")

	r.AddJobExecHandler("job_exec", "/jobs/{name}/exec.{ext}")

	r.AddWorkflowExecHandler("workflow_exec", "/workflows/{name}/exec.{ext}")

	r.AddLayerTileHandler("tile", "/layers/{name}/tiles/data/{z}/{x}/{y}.{ext}")

	r.AddLayerMaskHandler("mask", "/layers/{name}/tiles/mask/{z}/{x}/{y}.{ext}")

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
		Version: r.Version,
		GitBranch: r.GitBranch,
		GitCommit: r.GitCommit,
	}
}

func (r *RailgunRouter) AddObjectHandler(name string, path string, object interface{}) {
	r.Methods("Get").Name(name).Path(path).Handler(&handlers.ObjectHandler{
		Object:      object,
		BaseHandler: r.NewBaseHandler(),
	})
}

func (r *RailgunRouter) AddGroupHandler(name string, path string, t reflect.Type) {

	r.Messages <- map[string]interface{}{
		"init": map[string]interface{}{
			"handler": map[string]interface{}{"name": name, "path": path},
		},
	}

	r.Methods("GET", "POST", "PUT", "OPTIONS").Name(name).Path(path).Handler(&handlers.GroupHandler{
		Type:        t,
		BaseHandler: r.NewBaseHandler(),
	})
}

func (r *RailgunRouter) AddItemHandler(name string, path string, t reflect.Type, singular string, plural string) {

	r.Messages <- map[string]interface{}{
		"init": map[string]interface{}{
			"handler": map[string]interface{}{"name": name, "path": path},
		},
	}

	r.Methods("GET", "POST", "OPTIONS", "DELETE").Name(name).Path(path).Handler(&handlers.ItemHandler{
		Singular:    singular,
		Plural:      plural,
		Type:        t,
		BaseHandler: r.NewBaseHandler(),
	})
}

func (r *RailgunRouter) AddSwaggerHandler(name string, path string) {
	r.Methods("GET").Name(name).Path(path).Handler(&handlers.SwaggerHandler{
		BaseHandler: r.NewBaseHandler(),
	})
}

func (r *RailgunRouter) AddHealthHandler(name string, path string) {
	r.Methods("GET").Name(name).Path(path).Handler(&handlers.HealthHandler{
		BaseHandler: r.NewBaseHandler(),
	})
}

func (r *RailgunRouter) AddAuthenticateHandler(name string, path string) {
	r.Methods("POST").Name(name).Path(path).Handler(&handlers.AuthenticateHandler{
		BaseHandler: r.NewBaseHandler(),
	})
}

func (r *RailgunRouter) AddHomeHandler(name string, path string) {
	r.Methods("GET").Name(name).Path(path).Handler(&handlers.HomeHandler{
		BaseHandler: r.NewBaseHandler(),
	})
}

func (r *RailgunRouter) AddServiceExecHandler(name string, path string) {
	r.Methods("POST", "OPTIONS").Name(name).Path(path).Handler(&handlers.ServiceExecHandler{
		BaseHandler: r.NewBaseHandler(),
		Cache:       gocache.New(5*time.Minute, 10*time.Minute),
	})
}

func (r *RailgunRouter) AddServiceDownloadHandler(name string, path string) {
	r.Methods("GET").Name(name).Path(path).Handler(&handlers.ServiceDownloadHandler{
		BaseHandler: r.NewBaseHandler(),
		Cache:       gocache.New(5*time.Minute, 10*time.Minute),
	})
}

func (r *RailgunRouter) AddServiceTileHandler(name string, path string) {
	r.Methods("GET").Name(name).Path(path).Handler(&handlers.ServiceTileHandler{
		BaseHandler: r.NewBaseHandler(),
		Cache:       gocache.New(5*time.Minute, 10*time.Minute),
	})
}

func (r *RailgunRouter) AddJobExecHandler(name string, path string) {
	r.Methods("POST", "OPTIONS").Name(name).Path(path).Handler(&handlers.JobExecHandler{
		BaseHandler: r.NewBaseHandler(),
	})
}

func (r *RailgunRouter) AddWorkflowExecHandler(name string, path string) {
	r.Methods("POST", "OPTIONS").Name(name).Path(path).Handler(&handlers.WorkflowExecHandler{
		BaseHandler: r.NewBaseHandler(),
	})
}

func (r *RailgunRouter) AddLayerTileHandler(name string, path string) {
	r.Methods("GET").Name(name).Path(path).Handler(&handlers.LayerTileHandler{
		BaseHandler: r.NewBaseHandler(),
	})
}

func (r *RailgunRouter) AddLayerMaskHandler(name string, path string) {
	r.Methods("GET").Name(name).Path(path).Handler(&handlers.LayerMaskHandler{
		BaseHandler: r.NewBaseHandler(),
	})
}
