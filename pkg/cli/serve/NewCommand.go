// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package serve

import (
	"github.com/spatialcurrent/cobra"
)

type NewCommandInput struct {
	GitBranch string
	GitCommit string
}

// NewCommand returns a new instance of the serve command.
func NewCommand(input *NewCommandInput) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "run railgun server",
		Long:  "run railgun server",
		RunE:  serveFunction(input.GitBranch, input.GitCommit),
	}
	InitServeFlags(cmd.Flags())
	return cmd
}
