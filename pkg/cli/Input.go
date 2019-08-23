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
	FlagInputURI              string = "input-uri"
	FlagInputCompression      string = "input-compression"
	FlagInputFormat           string = "input-format"
	FlagInputHeader           string = "input-header"
	FlagInputLimit            string = "input-limit"
	FlagInputComment          string = "input-comment"
	FlagInputTrim             string = "input-trim"
	FlagInputLazyQuotes       string = "input-lazy-quotes"
	FlagInputPassphrase       string = "input-passphrase"
	FlagInputSalt             string = "input-salt"
	FlagInputReaderBufferSize string = "input-reader-buffer-size"
	FlagInputSkipBlanks       string = "input-skip-blanks"
	FlagInputSkipComments     string = "input-skip-comments"
	FlagInputSkipLines        string = "input-skip-lines"
	FlagInputDropCR           string = "input-drop-cr"
	FlagInputLineSeparator    string = "input-line-separator"
	FlagInputEscapePrefix     string = "input-escape-prefix"
	FlagInputUnescapeColon    string = "input-unescape-colon"
	FlagInputUnescapeEqual    string = "input-unescape-equal"
	FlagInputUnescapeSpace    string = "input-unescape-space"
	FlagInputUnescapeNewLine  string = "input-unescape-new-line"

	defaultInputFormat string = ""
)

func CheckInput(v *viper.Viper) error {
	inputFormat := v.GetString(FlagInputFormat)
	inputComment := v.GetString(FlagInputComment)

	if inputFormat == "csv" || inputFormat == "tsv" {
		if len(inputComment) > 1 {
			return errors.New("go's encoding/csv package only supports single character comment characters")
		}
	}
	return nil
}

func InitInputFlags(flag *pflag.FlagSet) {
	flag.StringP(FlagInputURI, "i", "stdin", "the input uri")
	flag.StringP(FlagInputCompression, "", "", "the input compression: "+strings.Join(GO_RAILGUN_COMPRESSION_ALGORITHMS, ", "))
	flag.String(FlagInputFormat, defaultInputFormat, "the input format: "+strings.Join(gss.Formats, ", "))
	flag.StringSliceP(FlagInputHeader, "", []string{}, "the input header")
	flag.IntP(FlagInputLimit, "", gss.NoLimit, "maximum number of objects to read from input")
	flag.StringP(FlagInputComment, "c", "", "the comment character for the input, e.g, #")
	flag.Bool(FlagInputTrim, false, "trim input lines")
	flag.Bool(FlagInputLazyQuotes, false, "allows lazy quotes for CSV and TSV")
	flag.String(FlagInputPassphrase, "", "input passphrase for AES-256 encryption")
	flag.String(FlagInputSalt, "", "input salt for AES-256 encryption")
	flag.Int(FlagInputReaderBufferSize, 4096, "the buffer size for the input reader")
	flag.Int(FlagInputSkipLines, gss.NoSkip, "the number of lines to skip before processing")
	flag.Bool(FlagInputSkipBlanks, false, "skip blank lines")
	flag.Bool(FlagInputSkipComments, false, "skip comments")
	flag.String(FlagInputLineSeparator, "\n", "override line separator.  Used with properties and JSONL formats.")
	flag.Bool(FlagInputDropCR, false, "drop carriage return characters that immediately precede new line characters")
	flag.String(FlagInputEscapePrefix, "", "override escape prefix.  Used with properties format.")
	flag.Bool(FlagInputUnescapeColon, false, "Unescape colon characters in input.  Used with properties format.")
	flag.Bool(FlagInputUnescapeEqual, false, "Unescape equal characters in input.  Used with properties format.")
	flag.Bool(FlagInputUnescapeSpace, false, "Unescape space characters in input.  Used with properties format.")
	flag.Bool(FlagInputUnescapeNewLine, false, "Unescape new line characters in input.  Used with properties format.")
}
