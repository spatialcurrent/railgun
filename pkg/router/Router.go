// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package router

import (
	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
	"github.com/spatialcurrent/railgun/pkg/request"
)

type Router struct {
	*mux.Router
	Requests        chan request.Request
	Messages        chan interface{}
	Errors          chan interface{}
	AwsSessionCache *gocache.Cache
}

func NewRouter(requests chan request.Request, messages chan interface{}, errors chan interface{}, awsSessionCache *gocache.Cache) *Router {
	return &Router{
		Router:          mux.NewRouter(),
		Requests:        requests,
		Messages:        messages,
		Errors:          errors,
		AwsSessionCache: awsSessionCache,
	}
}
