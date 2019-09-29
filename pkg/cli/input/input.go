// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package input

import (
	"github.com/pkg/errors"
)

const (
	FlagInputURI               string = "input-uri"
	FlagInputCompression       string = "input-compression"
	FlagInputDictionary        string = "input-dict"
	FlagInputFormat            string = "input-format"
	FlagInputType              string = "input-type"
	FlagInputHeader            string = "input-header"
	FlagInputLimit             string = "input-limit"
	FlagInputComment           string = "input-comment"
	FlagInputTrim              string = "input-trim"
	FlagInputLazyQuotes        string = "input-lazy-quotes"
	FlagInputPassphrase        string = "input-passphrase"
	FlagInputSalt              string = "input-salt"
	FlagInputReaderBufferSize  string = "input-reader-buffer-size"
	FlagInputSkipBlanks        string = "input-skip-blanks"
	FlagInputSkipComments      string = "input-skip-comments"
	FlagInputSkipLines         string = "input-skip-lines"
	FlagInputKeyValueSeparator string = "input-key-value-separator"
	FlagInputLineSeparator     string = "input-line-separator"
	FlagInputDropCR            string = "input-drop-cr"
	FlagInputEscapePrefix      string = "input-escape-prefix"
	FlagInputUnescapeColon     string = "input-unescape-colon"
	FlagInputUnescapeEqual     string = "input-unescape-equal"
	FlagInputUnescapeSpace     string = "input-unescape-space"
	FlagInputUnescapeNewLine   string = "input-unescape-new-line"

	DefaultInputKeyValueSeparator string = "="
	DefaultInputLineSeparator     string = "\n"
	DefaultInputReaderBufferSize         = 4096
)

var (
	ErrMissingInputKeyValueSeparator = errors.New("missing input key-value separator")
	ErrMissingInputLineSeparator     = errors.New("missing input line separator")
	ErrMissingInputEscapePrefix      = errors.New("missing input escape prefix")
)

func stringSliceContains(slc []string, str string) bool {
	for _, x := range slc {
		if x == str {
			return true
		}
	}
	return false
}
