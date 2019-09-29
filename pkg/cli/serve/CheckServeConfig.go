// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package serve

import (
	"github.com/pkg/errors"

	"github.com/spatialcurrent/railgun/pkg/cli/http"
	"github.com/spatialcurrent/railgun/pkg/cli/jwt"
	"github.com/spatialcurrent/railgun/pkg/cli/logging"
	"github.com/spatialcurrent/railgun/pkg/cli/swagger"
	"github.com/spatialcurrent/viper"
)

// CheckServeConfig checks the serve configuration.
func CheckServeConfig(v *viper.Viper, args []string) error {
	err := http.CheckHttpConfig(v)
	if err != nil {
		return errors.Wrap(err, "error with http configuration")
	}
	err = jwt.CheckJwtConfig(v)
	if err != nil {
		return errors.Wrap(err, "error with jwt configuration")
	}
	err = swagger.CheckSwaggerConfig(v)
	if err != nil {
		return errors.Wrap(err, "error with swagger configuration")
	}
	err = logging.CheckLoggingConfig(v)
	if err != nil {
		return errors.Wrap(err, "error with jwt configuration")
	}
	return nil
}
