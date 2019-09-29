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
	"strings"

	"github.com/pkg/errors"

	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
	"github.com/spatialcurrent/railgun/pkg/cli/logging"
	"github.com/spatialcurrent/railgun/pkg/cli/output"
	"github.com/spatialcurrent/railgun/pkg/config"
	"github.com/spatialcurrent/railgun/pkg/rest"
	"github.com/spatialcurrent/railgun/pkg/util"
	"github.com/spatialcurrent/viper"
)

const (
	CliUse   = "client"
	CliShort = "client commands for interacting with a Railgun server"
	CliLong  = "client commands for interacting with a Railgun server"
)

const (
	FlagJwtToken = "jwt-token"
	FlagServer   = "server"
	FlagUser     = "user"
	FlagPassword = "password"
	FlagName     = "name"

	DefaultServer = "http://localhost:8080"
)

const (
	FlagOutputURI         = output.FlagOutputURI
	FlagOutputCompression = output.FlagOutputCompression
	FlagOutputFormat      = output.FlagOutputFormat
	FlagOutputAppend      = output.FlagOutputAppend
	FlagOutputPretty      = output.FlagOutputPretty
	FlagVerbose           = logging.FlagVerbose
)

const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodDelete = "DELETE"
)

func initViper(cmd *cobra.Command) (*viper.Viper, error) {
	v := viper.New()
	err := v.BindPFlags(cmd.Flags())
	if err != nil {
		return nil, errors.Wrap(err, "error binding flags")
	}
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	util.MergeConfigs(v, v.GetStringArray("config-uri"))
	return v, nil
}

func initOutputWriter(v *viper.Viper) (*grw.Writer, error) {
	w, err := grw.WriteToResource(&grw.WriteToResourceInput{
		Uri:      v.GetString(FlagOutputURI),
		Alg:      v.GetString(FlagOutputCompression),
		Dict:     grw.NoDict,
		Append:   v.GetBool(FlagOutputAppend),
		S3Client: nil,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error opening output writer")
	}
	return w, nil
}

func makeRequest(v *viper.Viper, logger *gsl.Logger, method string, u string, auth string, object interface{}) error {

	w, err := initOutputWriter(v)
	if err != nil {
		return errors.Wrap(err, "error opening output writer")
	}

	in := &rest.MakeRequestInput{
		Url:          u,
		Method:       method,
		Object:       object,
		OutputWriter: w,
		OutputFormat: v.GetString(FlagOutputFormat),
		OutputPretty: v.GetBool(FlagOutputPretty),
		Logger:       logger,
	}

	if len(auth) > 0 {
		in.Authorization = auth
	}

	err = rest.MakeRequest(in)

	return nil
}

func authenticateFunction(cmd *cobra.Command, args []string) error {

	v, err := initViper(cmd)
	if err != nil {
		return errors.Wrap(err, "error initializing viper")
	}

	verbose := v.GetBool(FlagVerbose)

	logger := logging.NewLoggerFromViper(v)

	if verbose {
		config.PrintViperSettings(v)
	}

	u := fmt.Sprintf("%s/authenticate.%s", v.GetString("server"), v.GetString("output-format"))

	u2, err := url.Parse(u)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error parsing url ( %q )", u))
	}

	if strings.Contains(u2.Path, "//") {
		return errors.Errorf("url ( %q ) is invalid, beacuse \"//\" transforms POST to GET", u)
	}

	err = makeRequest(v, logger, "POST", u, "", map[string]interface{}{
		"username": v.GetString("username"),
		"password": v.GetString("password"),
	})
	if err != nil {
		return errors.Wrap(err, "error making request")
	}
	return nil

}
