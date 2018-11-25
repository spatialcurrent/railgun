package config

import (
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
	Uri          string   `viper:"output-uri"`
	Format       string   `viper:"output-format"`
	Header       []string `viper:"output-header"`
	Comment      string   `viper:"output-comment"`
	LazyQuotes   bool     `viper:"output-lazy-quotes"`
	Append       bool     `viper:"output-append"`
	BufferMemory bool     `viper:"output-buffer-memory"`
	Compression  string   `viper:"output-compression"`
	Passphrase   string   `viper:"output-passphrase"`
	Salt         string   `viper:"output-salt"`
	Limit        int      `viper:"output-limit"`
	Mkdirs       bool     `viper:"output-mkdirs"`
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
	}
}

func (o Output) Map() map[string]interface{} {
	return map[string]interface{}{
		"Uri":          o.Uri,
		"Format":       o.Format,
		"Header":       o.Header,
		"Comment":      o.Comment,
		"LazyQuotes":   o.LazyQuotes,
		"Append":       o.Append,
		"BufferMemory": o.BufferMemory,
		"Compression":  o.Compression,
		"Passphrase":   o.Passphrase,
		"Salt":         o.Salt,
		"Limit":        o.Limit,
		"Mkdirs":       o.Mkdirs,
	}
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
