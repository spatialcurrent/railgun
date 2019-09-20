// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package http

import (
	"time"

	"github.com/pkg/errors"
)

const (
	FlagHttpSchemes              = "http-schemes"
	FlagHttpLocation             = "http-location"
	FlagHttpAddress              = "http-address"
	FlagHttpTimeoutIdle          = "http-timeout-idle"
	FlagHttpTimeoutRead          = "http-timeout-read"
	FlagHttpTimeoutWrite         = "http-timeout-write"
	FlagHttpMiddlewareDebug      = "http-middleware-debug"
	FlagHttpMiddlewareRecover    = "http-middleware-recover"
	FlagHttpMiddlewareGzip       = "http-middleware-gzip"
	FlagHttpMiddlewareCors       = "http-middleware-cors"
	FlagHttpGracefulShutdown     = "http-graceful-shutdown"
	FlagHttpGracefulShutdownWait = "http-graceful-shutdown-wait"

	DefaultHttpLocation         = "http://localhost:8080"
	DefaultHttpAddress          = ":8080"
	DefaultHttpTimeoutIdle      = time.Second * 60
	DefaultHttpTimeoutRead      = time.Second * 15
	DefaultHttpTimeoutWrite     = time.Second * 15
	DefaultGracefulShutdownWait = time.Second * 15

	MinIdleTimeout          = time.Second * 15
	MinReadTimeout          = time.Second * 5
	MinWriteTimeout         = time.Second * 5
	MinGracefulShutdownWait = time.Second * 5
)

var (
	DefaultHttpSchemes = []string{"http"}
)

var (
	ErrMissingSchemes = errors.New("missing schemes, expecting at least one of http and https")
)
