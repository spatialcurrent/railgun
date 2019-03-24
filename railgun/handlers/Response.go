// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"net/http"
	"net/url"
)

type Response struct {
	Url        *url.URL
	Writer     http.ResponseWriter
	StatusCode int
	Format     string
	Filename   string
	Object     interface{}
	Pretty     bool
}
