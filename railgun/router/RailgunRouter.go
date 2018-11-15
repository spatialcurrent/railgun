package router

import (
	"compress/gzip"
	"crypto/rsa"
	"fmt"
	"github.com/NYTimes/gziphandler"
	gocache "github.com/patrickmn/go-cache"
	"github.com/spatialcurrent/go-adaptive-functions/af"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun/catalog"
	"github.com/spatialcurrent/railgun/railgun/core"
	"github.com/spatialcurrent/railgun/railgun/handlers"
	"github.com/spatialcurrent/railgun/railgun/request"
	"github.com/spatialcurrent/viper"
	"reflect"
	"strings"
	"time"
)

type RailgunRouter struct {
	*Router
	Viper           *viper.Viper
	Catalog         *catalog.RailgunCatalog
	PublicKey       *rsa.PublicKey
	PrivateKey      *rsa.PrivateKey
	ValidMethods    []string
	SessionDuration time.Duration
}

func NewRailgunRouter(v *viper.Viper, railgunCatalog *catalog.RailgunCatalog, requests chan request.Request, messages chan interface{}, errors chan error, awsSessionCache *gocache.Cache, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey, validMethods []string) *RailgunRouter {

	r := &RailgunRouter{
		Viper:           v,
		Catalog:         railgunCatalog,
		Router:          NewRouter(requests, messages, errors, awsSessionCache),
		PublicKey:       publicKey,
		PrivateKey:      privateKey,
		ValidMethods:    validMethods,
		SessionDuration: v.GetDuration("jwt-session-duration"),
	}

	//r.Use(GzipMiddleware)
	if v.GetBool("http-middleware-gzip") {
		r.Use(gziphandler.MustNewGzipLevelHandler(gzip.DefaultCompression))
	}
	r.Use(DebugMiddleware)
	r.Use(CorsMiddleware(v.GetString("cors-origin"), v.GetString("cors-credentials")))

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

	r.AddJobExecHandler("job_exec", "/jobs/{name}/exec.{ext}")

	r.AddWorkflowExecHandler("workflow_exec", "/workflows/{name}/exec.{ext}")

	r.AddTileHandler("tile", "/layers/{name}/data/tiles/{z}/{x}/{y}.{ext}")

	r.AddMaskHandler("mask", "/layers/{name}/mask/tiles/{z}/{x}/{y}.{ext}")

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
	}
}

func (r *RailgunRouter) AddObjectHandler(name string, path string, object interface{}) {
	r.Methods("Get").Name(name).Path(path).Handler(&handlers.ObjectHandler{
		Object:      object,
		BaseHandler: r.NewBaseHandler(),
	})
}

func (r *RailgunRouter) AddGroupHandler(name string, path string, t reflect.Type) {

	fmt.Println("* adding group handler " + name + " at path " + path)

	r.Methods("GET", "POST", "PUT", "OPTIONS").Name(name).Path(path).Handler(&handlers.GroupHandler{
		Type:        t,
		BaseHandler: r.NewBaseHandler(),
	})
}

func (r *RailgunRouter) AddItemHandler(name string, path string, t reflect.Type, singular string, plural string) {
	r.Methods("GET", "DELETE").Name(name).Path(path).Handler(&handlers.ItemHandler{
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

func (r *RailgunRouter) AddTileHandler(name string, path string) {
	r.Methods("GET").Name(name).Path(path).Handler(&handlers.TileHandler{
		BaseHandler: r.NewBaseHandler(),
	})
}

func (r *RailgunRouter) AddMaskHandler(name string, path string) {
	r.Methods("GET").Name(name).Path(path).Handler(&handlers.MaskHandler{
		BaseHandler: r.NewBaseHandler(),
	})
}
