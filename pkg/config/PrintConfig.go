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
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-simple-serializer/pkg/yaml"
)

func PrintConfig(c mapper) {
	fmt.Println("=================================================")
	fmt.Println("Configuration:")
	fmt.Println("-------------------------------------------------")
	b, err := yaml.Marshal(c.Map())
	if err != nil {
		fmt.Println(errors.Wrap(err, "error serializing process config").Error())
		os.Exit(1)
	}
	fmt.Println(string(b))
	fmt.Println("=================================================")
}
