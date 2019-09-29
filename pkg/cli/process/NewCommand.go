// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package process

import (
	"github.com/spatialcurrent/cobra"
)

// NewCommand returns a new instance of the process command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          CliUse,
		Short:        CliShort,
		Long:         CliLong,
		RunE:         processFunction,
		SilenceUsage: SilenceUsage,
	}
	InitProcessFlags(cmd.Flags())
	return cmd
}
