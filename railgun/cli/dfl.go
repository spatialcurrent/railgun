// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"github.com/spatialcurrent/cobra"
)

var dflCmd = &cobra.Command{
	Use:   "dfl",
	Short: "commands for go-dfl (DFL)",
	Long:  "commands for go-dfl (DFL)",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func init() {
	rootCmd.AddCommand(dflCmd)
}
