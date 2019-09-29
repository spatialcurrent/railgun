// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package config

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-simple-serializer/pkg/properties"
	"github.com/spatialcurrent/go-stringify/pkg/stringify"
	"github.com/spatialcurrent/viper"
)

func PrintViperSettings(v *viper.Viper) {
	fmt.Println("=================================================")
	fmt.Println("Viper:")
	fmt.Println("-------------------------------------------------")
	err := properties.Write(&properties.WriteInput{
		Writer:            os.Stdout,
		LineSeparator:     "\n",
		KeyValueSeparator: "=",
		Object:            v.AllSettings(),
		KeySerializer:     stringify.NewDefaultStringer(),
		ValueSerializer:   stringify.NewDefaultStringer(),
		Sorted:            true,
		Reversed:          false,
		EscapePrefix:      "\\",
		EscapeSpace:       false,
		EscapeEqual:       false,
		EscapeColon:       false,
		EscapeNewLine:     false,
	})
	if err != nil {
		fmt.Println(errors.Wrap(err, "error serializing viper settings").Error())
		os.Exit(1)
	}
	fmt.Println("")
	fmt.Println("=================================================")
}
