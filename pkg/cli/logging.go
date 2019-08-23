// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"github.com/spatialcurrent/pflag"
)

const (
	flagErrorDestination string = "error-destination"
	flagErrorCompression string = "error-compression"
	flagErrorFormat      string = "error-format"
	flagInfoDestination  string = "info-destination"
	flagInfoCompression  string = "info-compression"
	flagInfoFormat       string = "info-format"
	flagVerbose          string = "verbose"
)

func InitLoggingFlags(flag *pflag.FlagSet) {
	flag.BoolP("time", "t", false, "print timing output to info log")
	flag.BoolP(flagVerbose, "v", false, "print verbose output to stdout")
	flag.String(flagInfoDestination, "stdout", "destination for info logs as a uri")
	flag.String(flagInfoFormat, "tags", "info log format: text, properties, json, yaml, etc.")
	flag.String(flagInfoCompression, "", "the compression algorithm for the info logs: none, gzip, or snappy")
	flag.String(flagErrorDestination, "stderr", "destination for errors as a uri")
	flag.String(flagErrorCompression, "", "the compression algorithm for the errors: none, gzip, or snappy")
	flag.String(flagErrorFormat, "tags", "error log format: text, properties, json, yaml, etc.")
}
