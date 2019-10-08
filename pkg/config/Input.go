// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package config

import (
	"reflect"
	"strings"

	"github.com/spatialcurrent/go-reader-writer/pkg/splitter"
	"github.com/spatialcurrent/go-simple-serializer/pkg/serializer"
	"github.com/spatialcurrent/railgun/pkg/util"
)

type Input struct {
	Uri               string        `viper:"input-uri" map:"Uri"`
	Format            string        `viper:"input-format" map:"Format"`
	Header            []interface{} `viper:"input-header" map:"Header"`
	Type              string        `viper:"input-type" map:"Type"`
	Comment           string        `viper:"input-comment" map:"Comment"`
	LazyQuotes        bool          `viper:"input-lazy-quotes" map:"LazyQuotes"`
	Compression       string        `viper:"input-compression" map:"Compression"`
	Dictionary        []byte        `viper:"input-dict" map:"Dictionary"`
	ReaderBufferSize  int           `viper:"input-reader-buffer-size" map:"ReaderBufferSize"`
	Passphrase        string        `viper:"input-passphrase" map:"Passphrase"`
	Salt              string        `viper:"input-salt" map:"Salt"`
	SkipBlanks        bool          `viper:"input-skip-blanks" map:"SkipBlanks"`
	SkipComments      bool          `viper:"input-skip-comments" map:"SkipComments"`
	SkipLines         int           `viper:"input-skip-lines" map:"SkipLines"`
	Limit             int           `viper:"input-limit" map:"Limit"`
	KeyValueSeparator string        `viper:"input-key-value-separator" map:"KeyValueSeparator"`
	LineSeparator     string        `viper:"input-line-separator" map:"LineSeparator"`
	DropCR            bool          `viper:"input-drop-cr" map:"DropCR"`
	Trim              bool          `viper:"input-trim" map:"Trim"`
	EscapePrefix      string        `viper:"input-escape-prefix" map:"EscapePrefix"`
	UnescapeColon     bool          `viper:"input-unescape-colon" map:"UnescapeEqual"`
	UnescapeSpace     bool          `viper:"input-unescape-space" map:"UnescapeSpace"`
	UnescapeNewLine   bool          `viper:"input-unescape-new-line" map:"UnescapeNewLine"`
	UnescapeEqual     bool          `viper:"input-unescape-equal" map:"UnescapeEqual"`
}

func (i Input) CanStream() bool {
	if i.IsAthenaStoredQuery() {
		return true
	}

	switch i.Format {
	case serializer.FormatCSV, serializer.FormatTSV, serializer.FormatJSONL, serializer.FormatGob, serializer.FormatTags:
		return true
	}

	return false
}

func (i Input) HasFormat() bool {
	return len(i.Format) > 0
}

func (i Input) HasCompression() bool {
	return len(i.Compression) > 0
}

func (i Input) Path() string {
	_, inputPath := splitter.SplitUri(i.Uri)
	return inputPath
}

func (i Input) IsAthenaStoredQuery() bool {
	return strings.HasPrefix(i.Uri, "athena://")
}

func (i Input) IsS3Bucket() bool {
	return strings.HasPrefix(i.Uri, "s3://")
}

func (i Input) IsEncrypted() bool {
	return len(i.Passphrase) > 0
}

func (i Input) Map() map[string]interface{} {
	m := map[string]interface{}{}
	v := reflect.ValueOf(i)
	t := v.Type()
	for j := 0; j < v.NumField(); j++ {
		if tag := t.Field(j).Tag.Get("map"); len(tag) > 0 && tag != "-" {
			m[tag] = v.Field(j).Interface()
		}
	}
	return m
}

func (i *Input) Init() {
	if (!i.HasFormat()) && (!i.HasCompression()) {
		_, inputFormatGuess, inputCompressionGuess := util.SplitNameFormatCompression(i.Path())
		if len(i.Format) == 0 {
			i.Format = inputFormatGuess
		}
		if len(i.Compression) == 0 {
			i.Compression = inputCompressionGuess
		}
	}
}
