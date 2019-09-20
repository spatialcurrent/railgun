// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package logging

import (
	"github.com/spatialcurrent/pflag"
)

// InitLoggingFlags initializes the logging flags.
func InitLoggingFlags(flag *pflag.FlagSet) {
	flag.BoolP("time", "t", false, "print timing output to info log")

	flag.BoolP(FlagVerbose, "v", false, "print verbose output to stdout")

	flag.String(FlagInfoDestination, DefaultInfoDestination, "destination for info logs as a uri")
	flag.String(FlagInfoFormat, DefaultFormat, "info log format (as provided by gss)")
	flag.String(FlagInfoCompression, "", "the compression algorithm for the info messages (as provided by grw)")

	flag.String(FlagErrorDestination, DefaultErrorDestination, "destination for errors as a uri")
	flag.String(FlagErrorCompression, "", "the compression algorithm for the error messages (as provided by grw)")
	flag.String(FlagErrorFormat, DefaultFormat, "error log format (as provided by gss)")
}
