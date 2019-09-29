// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package process

import (
	"time"

	"github.com/spatialcurrent/pflag"

	"github.com/spatialcurrent/railgun/pkg/cli/dfl"
	"github.com/spatialcurrent/railgun/pkg/cli/input"
	"github.com/spatialcurrent/railgun/pkg/cli/output"
)

// InitProcessFlags initializes the process flags.
func InitProcessFlags(flag *pflag.FlagSet) {

	flag.BoolP("dry-run", "", false, "parse and compile expression, but do not evaluate against context")
	flag.BoolP(FlagStream, "s", false, "stream process (context == row rather than encompassing array)")
	flag.Duration("timeout", 1*time.Minute, "If not zero, then sets the timeout for the program.")

	input.InitInputFlags(flag)

	flag.String("temp-uri", "", "the temporary uri for storing results")

	output.InitOutputFlags(flag, "")

	dfl.InitDflFlags(flag)
}
