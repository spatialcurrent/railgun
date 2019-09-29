// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package process

import (
	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/railgun/pkg/cli/dfl"
	"github.com/spatialcurrent/railgun/pkg/cli/input"
	"github.com/spatialcurrent/railgun/pkg/cli/output"
	"github.com/spatialcurrent/viper"
)

// CheckProcessConfig checks the process configuration.
func CheckProcessConfig(v *viper.Viper, args []string) error {
	err := input.CheckInputConfig(v, gss.Formats)
	if err != nil {
		return errors.Wrap(err, "error with input configuration")
	}
	err = output.CheckOutputConfig(v, gss.Formats)
	if err != nil {
		return errors.Wrap(err, "error with output configuration")
	}
	err = dfl.CheckDflConfig(v)
	if err != nil {
		return errors.Wrap(err, "error with dfl configuration")
	}
	inputFormat := v.GetString(FlagInputFormat)
	stream := v.GetBool(FlagStream)
	if inputFormat == "gob" && !stream {
		return errors.New("input format \"gob\" requires streaming")
	}
	return nil
}
