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

import (
	"github.com/spatialcurrent/cobra"
)

var dflCmd = &cobra.Command{
	Use:   "dfl",
	Short: "commands for go-dfl (DFL)",
	Long:  "commands for go-dfl (DFL)",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Usage()
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(dflCmd)
}

const (
	flagDflExpression string = "dfl-expression"
	flagDflUri        string = "dfl-uri"
	flagDflVars       string = "dfl-vars"
)

func CheckDflFlags(flag *pflag.FlagSet) error {
	return nil
}

func InitDflFlags(flag *pflag.FlagSet) {
	flag.StringP(flagDflExpression, "d", "", "DFL expression to use")
	flag.String(flagDflUri, "", "URI to DFL file to use")
	flag.String(flagDflVars, "", "initial variables to use when evaluating DFL expression")
}
