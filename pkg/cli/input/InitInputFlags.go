// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package input

import (
	"strings"

	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/pflag"
)

// InitInputFlags initializes the input flags.
func InitInputFlags(flag *pflag.FlagSet) {
	flag.StringP(FlagInputURI, "i", "stdin", "the input uri")
	flag.StringP(FlagInputCompression, "", "", "the input compression algorithm, one of: "+strings.Join(grw.Algorithms, ", "))
	flag.String(FlagInputFormat, "", "the input format, one of: "+strings.Join(gss.Formats, ", "))
	flag.StringSliceP(FlagInputHeader, "", []string{}, "the input header")
	flag.IntP(FlagInputLimit, "", gss.NoLimit, "maximum number of objects to read from input")
	flag.StringP(FlagInputComment, "c", "", "the comment character for the input, e.g, #")
	flag.Bool(FlagInputTrim, false, "trim input lines")
	flag.Bool(FlagInputLazyQuotes, false, "allows lazy quotes for CSV and TSV")
	flag.String(FlagInputPassphrase, "", "input passphrase for AES-256 encryption")
	flag.String(FlagInputSalt, "", "input salt for AES-256 encryption")
	flag.Int(FlagInputReaderBufferSize, DefaultInputReaderBufferSize, "the buffer size for the input reader")
	flag.Int(FlagInputSkipLines, gss.NoSkip, "the number of lines to skip before processing")
	flag.Bool(FlagInputSkipBlanks, false, "skip blank lines")
	flag.Bool(FlagInputSkipComments, false, "skip comments")
	flag.String(FlagInputKeyValueSeparator, DefaultInputKeyValueSeparator, "override key-value separator")
	flag.String(FlagInputLineSeparator, DefaultInputLineSeparator, "override line separator.  Used with properties and JSONL formats.")
	flag.Bool(FlagInputDropCR, false, "drop carriage return characters that immediately precede new line characters")
	flag.String(FlagInputEscapePrefix, "", "override escape prefix.  Used with properties format.")
	flag.Bool(FlagInputUnescapeColon, false, "Unescape colon characters in input.  Used with properties format.")
	flag.Bool(FlagInputUnescapeEqual, false, "Unescape equal characters in input.  Used with properties format.")
	flag.Bool(FlagInputUnescapeSpace, false, "Unescape space characters in input.  Used with properties format.")
	flag.Bool(FlagInputUnescapeNewLine, false, "Unescape new line characters in input.  Used with properties format.")
}
