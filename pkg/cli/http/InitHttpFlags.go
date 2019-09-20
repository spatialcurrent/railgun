// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package http

import (
	"github.com/spatialcurrent/pflag"
)

// InitHttpFlags initializes the http flags.
func InitHttpFlags(flag *pflag.FlagSet) {
	flag.StringSlice(FlagHttpSchemes, DefaultHttpSchemes, "the \"public\" schemes")
	flag.StringP(FlagHttpLocation, "", DefaultHttpLocation, "the \"public\" location")
	flag.StringP(FlagHttpAddress, "a", DefaultHttpAddress, "http bind address")
	flag.DurationP(FlagHttpTimeoutIdle, "", DefaultHttpTimeoutIdle, "the idle timeout for the http server")
	flag.DurationP(FlagHttpTimeoutRead, "", DefaultHttpTimeoutRead, "the read timeout for the http server")
	flag.DurationP(FlagHttpTimeoutWrite, "", DefaultHttpTimeoutWrite, "the write timeout for the http server")
	flag.Bool(FlagHttpMiddlewareDebug, false, "enable debug middleware")
	flag.Bool(FlagHttpMiddlewareRecover, false, "enable recovery middleware")
	flag.Bool(FlagHttpMiddlewareGzip, false, "enable gzip middleware")
	flag.Bool(FlagHttpMiddlewareCors, false, "enable CORS middleware")
	flag.Bool(FlagHttpGracefulShutdown, false, "enable graceful shutdown")
	flag.Duration(FlagHttpGracefulShutdownWait, DefaultGracefulShutdownWait, "the duration to wait for graceful shutdown")
}
