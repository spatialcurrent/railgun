// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package util

import (
	"bytes"

	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/viper"
)

// MergeConfig merges a config from the given uri into the Viper config.
func MergeConfig(v *viper.Viper, configUri string) {

	_, configFormat, compression := SplitNameFormatCompression(configUri)
	if len(compression) > 0 {
		panic(errors.New("cannot have compression for config uri " + configUri))
	}

	v.SetConfigType(configFormat)

	configReader, _, err := grw.ReadFromResource(&grw.ReadFromResourceInput{
		Uri:        configUri,
		Alg:        "",
		Dict:       grw.NoDict,
		BufferSize: grw.DefaultBufferSize,
		S3Client:   nil,
	})
	if err != nil {
		panic(err)
	}

	configBytes, err := configReader.ReadAllAndClose()
	if err != nil {
		panic(err)
	}

	if len(configBytes) > 0 {
		err = v.MergeConfig(bytes.NewReader(configBytes))
		if err != nil {
			panic(err)
		}
	}
}
