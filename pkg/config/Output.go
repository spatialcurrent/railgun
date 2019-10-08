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
	"github.com/spatialcurrent/go-stringify/pkg/stringify"
	"github.com/spatialcurrent/railgun/pkg/util"
)

type Output struct {
	Uri               string        `viper:"output-uri" map:"Uri"`
	Format            string        `viper:"output-format" map:"Format"`
	FormatSpecifier   string        `viper:"output-format-specifier" map:"Format"`
	Header            []interface{} `viper:"output-header" map:"Header"`
	Comment           string        `viper:"output-comment" map:"Comment"`
	LazyQuotes        bool          `viper:"output-lazy-quotes" map:"LazyQuotes"`
	Append            bool          `viper:"output-append" map:"Append"`
	Overwrite         bool          `viper:"output-overwrite" map:"Overwrite"`
	BufferMemory      bool          `viper:"output-buffer-memory" map:"BufferMemory"`
	Compression       string        `viper:"output-compression" map:"Compression"`
	Dictionary        []byte        `viper:"output-dict" map:"Dictionary"`
	Passphrase        string        `viper:"output-passphrase" map:"Passphrase"`
	Salt              string        `viper:"output-salt" map:"Salt"`
	Limit             int           `viper:"output-limit" map:"Limit"`
	Fit               bool          `viper:"output-fit" map:"Fit"`
	Pretty            bool          `viper:"output-pretty" map:"Pretty"`
	Mkdirs            bool          `viper:"output-mkdirs" map:"Mkdirs"`
	Decimal           bool          `viper:"output-decimal" map:"Decimal"`
	KeyLower          bool          `viper:"output-key-lower" map:"KeyLower"`
	KeyUpper          bool          `viper:"output-key-upper" map:"KeyUpper"`
	ValueLower        bool          `viper:"output-value-lower" map:"ValueLower"`
	ValueUpper        bool          `viper:"output-value-upper" map:"ValueUpper"`
	NoDataValue       string        `viper:"output-no-data-value" map:"NoDataValue"`
	LineSeparator     string        `viper:"output-line-separator" map:"LineSeparator"`
	KeyValueSeparator string        `viper:"output-key-value-separator" map:"KeyValueSeparator"`
	Sorted            bool          `viper:"output-sorted" map:"Sorted"`
	Reversed          bool          `viper:"output-reversed" map:"Sorted"`
	ExpandHeader      bool          `viper:"output-expand-header" map:"ExpandHeader"`
	EscapePrefix      string        `viper:"output-escape-prefix" map:"EscapePrefix"`
	EscapeColon       bool          `viper:"output-escape-colon" map:"EscapeColon"`
	EscapeEqual       bool          `viper:"output-escape-equal" map:"EscapeEqual"`
	EscapeSpace       bool          `viper:"output-escape-space" map:"EscapeSpace"`
	EscapeNewLine     bool          `viper:"output-escape-new-line" map:"EscapeNewLine"`
}

func (o Output) CanStream() bool {
	if o.Sorted {
		return false
	}

	switch o.Format {
	case serializer.FormatCSV, serializer.FormatFmt, serializer.FormatGo, serializer.FormatGob, serializer.FormatJSONL, serializer.FormatTags, serializer.FormatTSV:
		return true
	}

	return false
}

func (o Output) HasFormat() bool {
	return len(o.Format) > 0
}

func (o Output) HasCompression() bool {
	return len(o.Compression) > 0
}

func (o Output) IsEncrypted() bool {
	return len(o.Passphrase) > 0
}

func (o Output) IsPretty() bool {
	return o.Pretty
}

func (o Output) Path() string {
	_, inputPath := splitter.SplitUri(o.Uri)
	return inputPath
}

func (o Output) IsAthenaStoredQuery() bool {
	return strings.HasPrefix(o.Uri, "athena://")
}

func (o Output) IsS3Bucket() bool {
	return strings.HasPrefix(o.Uri, "s3://")
}

func (o Output) Map() map[string]interface{} {
	m := map[string]interface{}{}
	v := reflect.ValueOf(o)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if tag := t.Field(i).Tag.Get("map"); len(tag) > 0 && tag != "-" {
			m[tag] = v.Field(i).Interface()
		}
	}
	return m
}

func (o *Output) Init() {
	if (!o.HasFormat()) && (!o.HasCompression()) {
		_, outputFormatGuess, outputCompressionGuess := util.SplitNameFormatCompression(o.Path())
		if len(o.Format) == 0 {
			o.Format = outputFormatGuess
		}
		if len(o.Compression) == 0 {
			o.Compression = outputCompressionGuess
		}
	}
}

func (o *Output) KeySerializer() stringify.Stringer {
	return stringify.NewStringer(
		o.NoDataValue,
		o.Decimal,
		o.KeyLower,
		o.KeyUpper)
}

func (o *Output) ValueSerializer() stringify.Stringer {
	return stringify.NewStringer(
		o.NoDataValue,
		o.Decimal,
		o.ValueLower,
		o.ValueUpper)
}
