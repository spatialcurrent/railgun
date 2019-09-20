// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package client

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/railgun/pkg/cli/logging"
	"github.com/spatialcurrent/railgun/pkg/config"
)

type NewRestCommandInput struct {
	Use    string
	Short  string
	Long   string
	Path   string
	Method string
	Params []string
	Type   reflect.Type
}

func NewRestCommand(input *NewRestCommandInput) *cobra.Command {
	return &cobra.Command{
		Use:   input.Use,
		Short: input.Short,
		Long:  input.Long,
		RunE: func(cmd *cobra.Command, args []string) error {

			v, err := initViper(cmd)
			if err != nil {
				return errors.Wrap(err, "error initializing viper")
			}

			verbose := v.GetBool(FlagVerbose)

			logger := logging.NewLoggerFromViper(v)

			if verbose {
				config.PrintViperSettings(v)
			}

			u := v.GetString("server") + strings.Replace(input.Path, "{ext}", "json", 1)

			obj := map[string]interface{}{}
			for _, name := range input.Params {
				value := v.GetString(name)
				if len(value) == 0 {
					return errors.Errorf("missing %q", name)
				}
				if input.Type == nil {
					obj[name] = value
				}
				u = strings.Replace(u, "{"+name+"}", value, 1)
			}

			if input.Type != nil {
				for i := 0; i < input.Type.NumField(); i++ {
					f := input.Type.Field(i)
					if str, ok := f.Tag.Lookup("rest"); ok && str != "" && str != "-" {
						if strings.Contains(str, ",") {
							arr := strings.SplitN(str, ",", 2)
							fieldValue := v.GetString(arr[0])
							if len(fieldValue) > 0 {
								obj[arr[0]] = fieldValue
							}
						} else {
							fieldValue := v.GetString(str)
							if len(fieldValue) > 0 {
								obj[str] = fieldValue
							}
						}
					}
				}
			}

			u2, err := url.Parse(u)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error parsing url ( %q )", u))
			}

			if strings.Contains(u2.Path, "//") {
				return errors.Errorf("url ( %q ) is invalid, beacuse \"//\" transforms POST to GET", u)
			}

			err = makeRequest(v, logger, input.Method, u, v.GetString(FlagJwtToken), obj)
			if err != nil {
				return errors.Wrap(err, "error making request")
			}
			return nil
		},
	}
}
