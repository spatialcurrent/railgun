// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	//"fmt"
	//"reflect"

	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/railgun/pkg/cli/output"
	"github.com/spatialcurrent/railgun/pkg/catalog"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
)

// NewCommand returns a new instance of the catalog command.
func NewCommand() *cobra.Command {
	catalogCommand := &cobra.Command{
		Use:   CliUse,
		Short: CliShort,
		Long:  CliLong,
	}
	InitCatalogFlags(catalogCommand.PersistentFlags())
	output.InitOutputFlags(catalogCommand.PersistentFlags(), "json")
	
	catalogCommand.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "show catalog",
		Long:  "show catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
		  
  		v, err := initViper(cmd)
			if err != nil {
				return errors.Wrap(err, "error initializing viper")
			}
			
			if err := CheckCatalogConfig(v, gss.Formats, grw.Algorithms); err != nil {
			  return errors.Wrap(err, "error with configuration")
			}
			
			uri := v.GetString(FlagCatalogUri)
			format := v.GetString(FlagCatalogFormat)
			compression := v.GetString(FlagCatalogCompression)
			
			
  		railgunCatalog := catalog.NewRailgunCatalog()
  
  		if err = railgunCatalog.LoadFromViper(v); err != nil {
  			return errors.Wrap(err, "error loading catalog from viper")
  		}
  
  		if len(catalogUri) > 0 {
  			err := railgunCatalog.LoadFromUri(&railgunCatalog.LoadFromUriInput{
          Uri: uri,
          Format: format,
          Compression: compression,
          Logger: logger,
          S3Client: nil,
  			})
  			if err != nil {
  				return errors.Wrapf(err, "error loading catalog from uri %q", uri)
  			}
  		}
  		return nil
		},
	})

	return catalogCommand
}
