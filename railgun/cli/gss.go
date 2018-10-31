// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"github.com/spf13/cobra"
)

var gssCmd = &cobra.Command{
	Use:   "gss",
	Short: "commands for go-simple-serializer (GSS)",
	Long:  "commands for go-simple-serializer (GSS)",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func init() {
	rootCmd.AddCommand(gssCmd)
}
