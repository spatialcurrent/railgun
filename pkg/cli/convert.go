// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/go-stringify/pkg/stringify"
	"github.com/spatialcurrent/railgun/pkg/util"
	"github.com/spatialcurrent/viper"
)

func convertFunction(cmd *cobra.Command, args []string) {

	v := viper.New()
	err := v.BindPFlags(cmd.Flags())
	if err != nil {
		panic(err)
	}
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	util.MergeConfigs(v, v.GetStringArray("config-uri"))

	inputBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println(errors.Wrap(err, "error reading from stdin"))
		os.Exit(1)
	}

	outputString, err := gss.Convert(&gss.ConvertInput{
		InputBytes:      inputBytes,
		InputFormat:     v.GetString("input-format"),
		InputHeader:     stringify.StringSliceToInterfaceSlice(v.GetStringSlice("input-header")),
		InputComment:    v.GetString("input-comment"),
		InputLazyQuotes: v.GetBool("input-lazy-quotes"),
		InputSkipLines:  v.GetInt("input-skip-lines"),
		InputLimit:      v.GetInt("input-limit"),
		OutputFormat:    v.GetString("output-format"),
		OutputHeader:    stringify.StringSliceToInterfaceSlice(v.GetStringSlice("output-header")),
		OutputLimit:     v.GetInt("output-limit"),
	})
	if err != nil {
		fmt.Println(errors.Wrap(err, "error converting"))
		os.Exit(1)
	}
	fmt.Println(outputString)

}

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "use go-simple-serializer (GSS) to convert between formats",
	Long:  `Use go-simple-serializer (GSS) to convert between formats.`,
	Run:   convertFunction,
}

func init() {
	gssCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringP("input-format", "i", "", "the input format: "+strings.Join(gss.Formats, ", "))
	err := convertCmd.MarkFlagRequired("input-format")
	if err != nil {
		panic(err)
	}
	convertCmd.Flags().StringP("input-comment", "c", "", "the input comment character, e.g., #.  Commented lines are not sent to output")
	convertCmd.Flags().Bool("input-lazy-quotes", false, "allows lazy quotes for CSV and TSV")
	convertCmd.Flags().StringSlice("input-header", []string{}, "the input header, if the input has no header")
	convertCmd.Flags().Int("input-skip-lines", gss.NoSkip, "the number of input lines to skip before processing")
	convertCmd.Flags().Int("input-limit", gss.NoLimit, "maximum number of objects to read from input")

	convertCmd.Flags().StringP("output-format", "o", "", "the output format: "+strings.Join(gss.Formats, ", "))
	err = convertCmd.MarkFlagRequired("output-format")
	if err != nil {
		panic(err)
	}
	convertCmd.Flags().StringSlice("output-header", []string{}, "the output header")
	convertCmd.Flags().Int("output-limit", gss.NoLimit, "maximum number of objects to send to output")

}
