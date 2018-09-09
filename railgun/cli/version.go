// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

import (
	"github.com/spatialcurrent/railgun/railgun"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version to stdout",
	Long:  "print version to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(railgun.VERSION)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
