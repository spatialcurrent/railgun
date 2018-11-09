package router

import (
	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
	"github.com/spatialcurrent/railgun/railgun/request"
)

type Router struct {
	*mux.Router
	Requests        chan request.Request
	Messages        chan interface{}
	Errors          chan error
	AwsSessionCache *gocache.Cache
}

func NewRouter(requests chan request.Request, messages chan interface{}, errors chan error, awsSessionCache *gocache.Cache) *Router {
	return &Router{
		Router:          mux.NewRouter(),
		Requests:        requests,
		Messages:        messages,
		Errors:          errors,
		AwsSessionCache: awsSessionCache,
	}
}
