// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package runtime

import (
	"github.com/spatialcurrent/pflag"
)

// InitRuntimeFlags initializes the Runtime flags.
func InitRuntimeFlags(flag *pflag.FlagSet) {
	flag.Int(FlagRuntimeMaxProcs, 1, "Sets the maximum number of parallel processes.  If set to zero, then sets the maximum number of parallel processes to the number of CPUs on the machine. (https://godoc.org/runtime#GOMAXPROCS).")
}
