// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package logging

const (
	FlagErrorDestination = "error-destination"
	FlagErrorCompression = "error-compression"
	FlagErrorFormat      = "error-format"
	FlagInfoDestination  = "info-destination"
	FlagInfoCompression  = "info-compression"
	FlagInfoFormat       = "info-format"
	FlagVerbose          = "verbose"

	DefaultFormat           = "tags"
	DefaultInfoDestination  = "stdout"
	DefaultErrorDestination = "stderr"
)
