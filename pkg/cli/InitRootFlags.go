// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"github.com/spatialcurrent/pflag"

	"github.com/spatialcurrent/railgun/pkg/cli/aws"
	"github.com/spatialcurrent/railgun/pkg/cli/logging"
	"github.com/spatialcurrent/railgun/pkg/cli/runtime"
)

// InitRootFlags initializes the root flags.
func InitRootFlags(flag *pflag.FlagSet) {
	aws.InitAwsFlags(flag)
	logging.InitLoggingFlags(flag)
	runtime.InitRuntimeFlags(flag)

	flag.StringArrayP("config-uri", "", []string{}, "the uri(s) to the config file")
}
