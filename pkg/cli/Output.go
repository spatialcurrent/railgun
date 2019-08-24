// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"strings"
)

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/pflag"
	"github.com/spatialcurrent/viper"
)

import (
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
)

const (
	FlagOutputURI           string = "output-uri"
	FlagOutputCompression   string = "output-compression"
	FlagOutputFormat        string = "output-format"
	FlagOutputPretty        string = "output-pretty"
	FlagOutputHeader        string = "output-header"
	FlagOutputLimit         string = "output-limit"
	FlagOutputAppend        string = "output-append"
	FlagOutputOverwrite     string = "output-overwrite"
	FlagOutputBufferMemory  string = "output-buffer-memory"
	FlagOutputMkdirs        string = "output-mkdirs"
	FlagOutputPassphrase    string = "output-passphrase"
	FlagOutputSalt          string = "output-salt"
	FlagOutputDecimal       string = "output-decimal"
	FlagOutputKeyLower      string = "output-key-lower"
	FlagOutputKeyUpper      string = "output-key-upper"
	FlagOutputValueLower    string = "output-value-lower"
	FlagOutputValueUpper    string = "output-value-upper"
	FlagOutputNoDataValue   string = "output-no-data-value"
	FlagOutputKeyValueSeparator string = "output-key-value-separator"
	FlagOutputLineSeparator string = "output-line-separator"
	FlagOutputExpandHeader  string = "output-expand-header"
	FlagOutputEscapePrefix  string = "output-escape-prefix"
	FlagOutputEscapeColon   string = "output-escape-colon"
	FlagOutputEscapeEqual   string = "output-escape-equal"
	FlagOutputEscapeNewLine string = "output-escape-new-line"
	FlagOutputEscapeSpace   string = "output-escape-space"
	FlagOutputSorted        string = "output-sorted"
	FlagOutputReversed      string = "output-reversed"

	DefaultOutputLineSeparator = "\n"
)

func CheckOutput(v *viper.Viper) error {
	if newline := v.GetString(FlagOutputLineSeparator); len(newline) != 1 {
		return errors.New("newline must be 1 character")
	}
	return nil
}

func InitOutputFlags(flag *pflag.FlagSet, defaultOutputFormat string) {
	flag.StringP(FlagOutputURI, "o", "stdout", "the output uri (a dfl expression itself)")
	flag.StringP(FlagOutputCompression, "", "", "the output compression: "+strings.Join(GO_RAILGUN_COMPRESSION_ALGORITHMS, ", "))
	flag.String(FlagOutputFormat, defaultOutputFormat, "the output format: "+strings.Join(gss.Formats, ", "))
	flag.BoolP(FlagOutputPretty, "p", false, "output pretty format")
	flag.StringSliceP(FlagOutputHeader, "", []string{}, "the output header")
	flag.IntP(FlagOutputLimit, "", gss.NoLimit, "maximum number of objects to send to output")
	flag.BoolP(FlagOutputAppend, "", false, "append to output files")
	flag.BoolP(FlagOutputOverwrite, "", false, "overwrite output if it already exists")
	flag.BoolP(FlagOutputBufferMemory, "b", false, "buffer output in memory")
	flag.Bool(FlagOutputMkdirs, false, "make directories if missing for output files")
	flag.StringP(FlagOutputPassphrase, "", "", "output passphrase for AES-256 encryption")
	flag.StringP(FlagOutputSalt, "", "", "output salt for AES-256 encryption")
	flag.Bool(FlagOutputDecimal, false, "when converting floats to strings use decimals rather than scientific notation")
	flag.Bool(FlagOutputKeyLower, false, "lower case output keys, including CSV column headers, tag names, and property names")
	flag.Bool(FlagOutputKeyUpper, false, "upper case output keys, including CSV column headers, tag names, and property names")
	flag.Bool(FlagOutputValueLower, false, "lower case output values, including tag values, and property values")
	flag.Bool(FlagOutputValueUpper, false, "upper case output values, including tag values, and property values")
	flag.String(FlagOutputNoDataValue, "", "no data value, e.g., used for missing values when converting JSON to CSV")
	flag.String(FlagOutputKeyValueSeparator, "=", "override key-value separator.")
	flag.String(FlagOutputLineSeparator, DefaultOutputLineSeparator, "override new line value.  Used with properties and JSONL formats.")
	flag.Bool(FlagOutputExpandHeader, false, "expand output header.  Used with CSV and TSV formats.")
	flag.String(FlagOutputEscapePrefix, "", "override escape prefix.  Used with properties format.")
	flag.Bool(FlagOutputEscapeColon, false, "Escape colon characters in output.  Used with properties format.")
	flag.Bool(FlagOutputEscapeEqual, false, "Escape equal characters in output.  Used with properties format.")
	flag.Bool(FlagOutputEscapeSpace, false, "Escape space characters in output.  Used with properties format.")
	flag.Bool(FlagOutputEscapeNewLine, false, "Escape new line characters in output.  Used with properties format.")
	flag.Bool(FlagOutputSorted, false, "sort output")
	flag.Bool(FlagOutputReversed, false, "if sorting output, sort in reverse alphabetical order")
}
