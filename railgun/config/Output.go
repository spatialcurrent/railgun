package config

import (
	"reflect"
	"strings"
)

import (
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
)

import (
	"github.com/spatialcurrent/railgun/railgun/util"
)

type Output struct {
	Uri          string   `viper:"output-uri" map:"Uri"`
	Format       string   `viper:"output-format" map:"Format"`
	Header       []string `viper:"output-header" map:"Header"`
	Comment      string   `viper:"output-comment" map:"Comment"`
	LazyQuotes   bool     `viper:"output-lazy-quotes" map:"LazyQuotes"`
	Append       bool     `viper:"output-append" map:"Append"`
	Overwrite    bool     `viper:"output-overwrite" map:"Overwrite"`
	BufferMemory bool     `viper:"output-buffer-memory" map:"BufferMemory"`
	Compression  string   `viper:"output-compression" map:"Compression"`
	Passphrase   string   `viper:"output-passphrase" map:"Passphrase"`
	Salt         string   `viper:"output-salt" map:"Salt"`
	Limit        int      `viper:"output-limit" map:"Limit"`
	Pretty       bool     `viper:"output-pretty" map:"Pretty"`
	Mkdirs       bool     `viper:"output-mkdirs" map:"Mkdirs"`
}

func (o Output) CanStream() bool {
	return o.Format == "csv" || o.Format == "tsv" || o.Format == "jsonl"
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
	_, inputPath := grw.SplitUri(o.Uri)
	return inputPath
}

func (o Output) IsAthenaStoredQuery() bool {
	return strings.HasPrefix(o.Uri, "athena://")
}

func (o Output) IsS3Bucket() bool {
	return strings.HasPrefix(o.Uri, "s3://")
}

func (o Output) Options() gss.Options {
	return gss.Options{
		Format: o.Format,
		Header: o.Header,
		Limit:  o.Limit,
		Pretty: o.Pretty,
	}
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
