// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package algorithms

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/viper"

	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-simple-serializer/pkg/serializer"
)

// NewCommand returns a new instance of the version command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   CliUse,
		Short: CliShort,
		Long:  CliLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.New()

			err := v.BindPFlags(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "error binding flags")
			}
			v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
			v.AutomaticEnv() // set environment variables to overwrite config

			err = CheckAlgorithmsConfig(v)
			if err != nil {
				return errors.Wrap(err, "error with configuration")
			}

			f := v.GetString(FlagFormat)

			b, err := serializer.New(f).LineSeparator("\n").Serialize(grw.Algorithms)
			if err != nil {
				return errors.Wrapf(err, "error serializing algorithms with format %q", f)
			}

			fmt.Print(string(b))

			return nil
		},
	}
	InitAlgorithmsFlags(cmd.Flags())
	return cmd
}