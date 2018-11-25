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

type Input struct {
	Uri              string   `viper:"input-uri"`
	Format           string   `viper:"input-format"`
	Header           []string `viper:"input-header"`
	Comment          string   `viper:"input-comment"`
	LazyQuotes       bool     `viper:"input-lazy-quotes"`
	Compression      string   `viper:"input-compression"`
	ReaderBufferSize int      `viper:"input-reader-buffer-size"`
	Passphrase       string   `viper:"input-passphrase"`
	Salt             string   `viper:"input-salt"`
	SkipLines        int      `viper:"input-skip-lines"`
	Limit            int      `viper:"input-limit"`
}

func (i Input) CanStream() bool {
	return i.Format == "csv" || i.Format == "tsv" || i.Format == "jsonl"
}

func (i Input) HasFormat() bool {
	return len(i.Format) > 0
}

func (i Input) HasCompression() bool {
	return len(i.Compression) > 0
}

func (i Input) Path() string {
	_, inputPath := grw.SplitUri(i.Uri)
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

func (i Input) Options() gss.Options {
	//objects, err := gss.DeserializeBytes(inputBytesPlain, inputConfig.Format, inputConfig.Header, inputConfig.Comment, inputConfig.LazyQuotes, inputConfig.SkipLines, inputConfig.Limit, inputType, verbose)
	return gss.Options{
		Format:     i.Format,
		Header:     i.Header,
		Comment:    i.Comment,
		LazyQuotes: i.LazyQuotes,
		SkipLines:  i.SkipLines,
		Limit:      i.Limit,
	}
}

func (i Input) Map() map[string]interface{} {
	return map[string]interface{}{
		"Uri":              i.Uri,
		"Format":           i.Format,
		"Header":           i.Header,
		"Comment":          i.Comment,
		"LazyQuotes":       i.LazyQuotes,
		"Compression":      i.Compression,
		"ReaderBufferSize": i.ReaderBufferSize,
		"Passphrase":       i.Passphrase,
		"Salt":             i.Salt,
		"SkipLines":        i.SkipLines,
		"Limit":            i.Limit,
	}
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
