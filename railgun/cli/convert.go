// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/viper"
	"io/ioutil"
	"os"
	"strings"
)

var convertViper = viper.New()

func convertFunction(cmd *cobra.Command, args []string) {
	v := convertViper

	verbose := v.GetBool("verbose")
	inputFormat := v.GetString("input-format")
	inputHeader := v.GetStringArray("input-header")
	inputSkipLines := v.GetInt("input-skip-lines")
	inputLimit := v.GetInt("input-limit")
	inputComment := v.GetString("input-comment")
	inputLazyQuotes := v.GetBool("input-lazy-quotes")
	outputFormat := v.GetString("output-format")
	outputHeader := v.GetStringArray("output-header")
	outputLimit := v.GetInt("output-limit")

	inputBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println(errors.Wrap(err, "error reading from stdin"))
		os.Exit(1)
	}

	outputString, err := gss.Convert(inputBytes, inputFormat, inputHeader, inputComment, inputLazyQuotes, inputSkipLines, inputLimit, outputFormat, outputHeader, outputLimit, false, verbose)
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
	convertCmd.MarkFlagRequired("input-format")
	convertCmd.Flags().StringP("input-comment", "c", "", "the input comment character, e.g., #.  Commented lines are not sent to output")
	convertCmd.Flags().BoolP("input-lazy-quotes", "", false, "allows lazy quotes for CSV and TSV")
	convertCmd.Flags().StringSliceP("input-header", "", gss.NoHeader, "the input header, if the input has no header")
	convertCmd.Flags().Int("input-skip-lines", gss.NoSkip, "the number of input lines to skip before processing")
	convertCmd.Flags().Int("input-limit", gss.NoLimit, "maximum number of objects to read from input")

	convertCmd.Flags().StringP("output-format", "o", "", "the output format: "+strings.Join(gss.Formats, ", "))
	convertCmd.MarkFlagRequired("output-format")
	convertCmd.Flags().StringArray("output-header", gss.NoHeader, "the output header")
	convertCmd.Flags().Int("output-limit", gss.NoLimit, "maximum number of objects to send to output")

	// Bind to Viper
	convertViper.BindPFlags(convertCmd.PersistentFlags())
	convertViper.BindPFlags(convertCmd.Flags())

}
