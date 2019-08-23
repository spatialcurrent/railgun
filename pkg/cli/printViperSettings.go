// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"fmt"
	"os"
)

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/viper"
)

import (
	"github.com/spatialcurrent/go-stringify/pkg/stringify"
)

import (
	"github.com/spatialcurrent/go-simple-serializer/pkg/properties"
)

func printViperSettings(v *viper.Viper) {
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
