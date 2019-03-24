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

type Input struct {
	Uri              string   `viper:"input-uri" map:"Uri"`
	Format           string   `viper:"input-format" map:"Format"`
	Header           []string `viper:"input-header" map:"Header"`
	Comment          string   `viper:"input-comment" map:"Comment"`
	LazyQuotes       bool     `viper:"input-lazy-quotes" map:"LazyQuotes"`
	Compression      string   `viper:"input-compression" map:"Compression"`
	ReaderBufferSize int      `viper:"input-reader-buffer-size" map:"ReaderBufferSize"`
	Passphrase       string   `viper:"input-passphrase" map:"Passphrase"`
	Salt             string   `viper:"input-salt" map:"Salt"`
	SkipLines        int      `viper:"input-skip-lines" map:"SkipLines"`
	Limit            int      `viper:"input-limit" map:"Limit"`
}

func (i Input) CanStream() bool {
	return i.IsAthenaStoredQuery() || i.Format == "csv" || i.Format == "tsv" || i.Format == "jsonl"
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
