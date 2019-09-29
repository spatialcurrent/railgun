// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package output

import (
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/pflag"
)

// InitOutputFlags initializes the output flags.
func InitOutputFlags(flag *pflag.FlagSet, defaultOutputFormat string) {
	flag.StringP(FlagOutputURI, "o", "stdout", "the output uri (a dfl expression itself)")
	flag.String(FlagOutputCompression, "", "the output compression algorithm")
	flag.String(FlagOutputFormat, defaultOutputFormat, "the output format")
	flag.String(FlagOutputFormatSpecifier, "", "The output format specifier")
	flag.Bool(FlagOutputFit, false, "fit output")
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
