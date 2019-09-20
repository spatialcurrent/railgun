// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package client

import (
	"fmt"
	"reflect"

	"github.com/spatialcurrent/cobra"
)

func NewRestCommands(baseurl string, singular string, plural string, inputType reflect.Type) []*cobra.Command {

	addCmd := NewRestCommand(&NewRestCommandInput{
		Use:    "add",
		Short:  fmt.Sprintf("add %s to Railgun Server", singular),
		Long:   fmt.Sprintf("add %s to Railgun Server", singular),
		Method: MethodPost,
		Path:   baseurl + ".{ext}",
		Params: []string{},
		Type:   inputType,
	})
	InitRestFlags(addCmd.Flags(), inputType)

	getCmd := NewRestCommand(&NewRestCommandInput{
		Use:    "get",
		Short:  fmt.Sprintf("get %s on Railgun Server", singular),
		Long:   fmt.Sprintf("get %s on Railgun Server", singular),
		Path:   baseurl + "/{name}.{ext}",
		Method: MethodGet,
		Params: []string{"name"},
	})
	getCmd.Flags().String("name", "", fmt.Sprintf("name of %s on Railgun Server", singular))

	updateCmd := NewRestCommand(&NewRestCommandInput{
		Use:    "update",
		Short:  fmt.Sprintf("update %s on Railgun Server", singular),
		Long:   fmt.Sprintf("update %s on Railgun Server", singular),
		Method: MethodPost,
		Path:   baseurl + "/{name}.{ext}",
		Params: []string{"name"},
		Type:   inputType,
	})
	InitRestFlags(updateCmd.Flags(), inputType)

	deleteCmd := NewRestCommand(&NewRestCommandInput{
		Use:    "delete",
		Short:  fmt.Sprintf("delete %s on Railgun Server", singular),
		Long:   fmt.Sprintf("delete %s on Railgun Server", singular),
		Path:   baseurl + "/{name}.{ext}",
		Method: MethodDelete,
		Params: []string{"name"},
	})
	deleteCmd.Flags().String("name", "", fmt.Sprintf("name of %s on Railgun Server", singular))

	listCmd := NewRestCommand(&NewRestCommandInput{
		Use:    "list",
		Short:  fmt.Sprintf("list %s on Railgun Server", plural),
		Long:   fmt.Sprintf("list %s on Railgun Server", plural),
		Path:   baseurl + ".{ext}",
		Method: MethodGet,
		Params: []string{},
	})

	return []*cobra.Command{addCmd, getCmd, updateCmd, deleteCmd, listCmd}

}
