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
	"github.com/spatialcurrent/railgun/pkg/core"
)

// NewCommand returns a new instance of the process command.
func NewCommand() *cobra.Command {
	clientCmd := &cobra.Command{
		Use:   CliUse,
		Short: CliShort,
		Long:  CliLong,
	}
	InitClientFlags(clientCmd.PersistentFlags())

	clientCmd.AddCommand(func() *cobra.Command {
		c := &cobra.Command{
			Use:   "authenticate",
			Short: "authenticate with Railgun Server",
			Long:  "authenticate with Railgun Server",
			RunE:  authenticateFunction,
		}
		InitClientAuthFlags(c.Flags())
		return c
	}())

	clientCmd.AddCommand(func() *cobra.Command {
		c := &cobra.Command{
			Use:   "workspaces",
			Short: "interact with workspaces on Railgun Server",
			Long:  "interact with workspaces on Railgun Server",
		}
		c.AddCommand(NewRestCommands("/workspaces", "workspace", "workspace", core.WorkspaceType)...)
		return c
	}())

	clientCmd.AddCommand(func() *cobra.Command {
		c := &cobra.Command{
			Use:   "datastores",
			Short: "interact with datastores on Railgun Server",
			Long:  "interact with datastores on Railgun Server",
		}
		c.AddCommand(NewRestCommands("/datastores", "data store", "data stores", core.DataStoreType)...)
		return c
	}())

	clientCmd.AddCommand(func() *cobra.Command {
		c := &cobra.Command{
			Use:   "layers",
			Short: "interact with layers on Railgun Server",
			Long:  "interact with layers on Railgun Server",
		}
		c.AddCommand(NewRestCommands("/layers", "layer", "layers", core.LayerType)...)
		return c
	}())

	clientCmd.AddCommand(func() *cobra.Command {
		c := &cobra.Command{
			Use:   "processes",
			Short: "interact with processes on Railgun Server",
			Long:  "interact with processes on Railgun Server",
		}
		c.AddCommand(NewRestCommands("/processes", "process", "processes", core.ProcessType)...)
		return c
	}())

	clientCmd.AddCommand(func(singular string, plural string, t reflect.Type) *cobra.Command {
		c := &cobra.Command{
			Use:   plural,
			Short: fmt.Sprintf("interact with %s on the server", plural),
			Long:  fmt.Sprintf("interact with %s on the server", plural),
		}
		c.AddCommand(NewRestCommands("/"+plural, singular, plural, t)...)

		c.AddCommand(func(t reflect.Type) *cobra.Command {
			e := NewRestCommand(&NewRestCommandInput{
				Use:    "exec",
				Short:  fmt.Sprintf("execute %s on the server", singular),
				Long:   fmt.Sprintf("execute %s on the server", singular),
				Method: MethodPost,
				Path:   fmt.Sprintf("/%s/{name}/exec.{ext}", plural),
				Params: []string{"name"},
				Type:   t,
			})
			InitRestFlags(e.Flags(), t)
			return e
		}(core.JobType))

		return c
	}("service", "services", core.ServiceType))

	clientCmd.AddCommand(func(singular string, plural string, t reflect.Type) *cobra.Command {
		c := &cobra.Command{
			Use:   plural,
			Short: fmt.Sprintf("interact with %s on the server", plural),
			Long:  fmt.Sprintf("interact with %s on the server", plural),
		}
		c.AddCommand(NewRestCommands("/"+plural, singular, plural, t)...)

		c.AddCommand(func() *cobra.Command {
			e := NewRestCommand(&NewRestCommandInput{
				Use:    "exec",
				Short:  fmt.Sprintf("execute %s on the server", singular),
				Long:   fmt.Sprintf("execute %s on the server", singular),
				Path:   fmt.Sprintf("/%s/{name}/exec.{ext}", plural),
				Method: MethodPost,
				Params: []string{"name"},
			})
			e.Flags().String(FlagName, "", fmt.Sprintf("name of %s on the Server", singular))
			return e
		}())
		return c
	}("job", "jobs", core.JobType))

	clientCmd.AddCommand(func(singular string, plural string, t reflect.Type) *cobra.Command {
		c := &cobra.Command{
			Use:   plural,
			Short: fmt.Sprintf("interact with %s on the server", plural),
			Long:  fmt.Sprintf("interact with %s on the server", plural),
		}
		c.AddCommand(NewRestCommands("/"+plural, singular, plural, t)...)
		c.AddCommand(func() *cobra.Command {
			e := NewRestCommand(&NewRestCommandInput{
				Use:    "exec",
				Short:  fmt.Sprintf("execute %s on the server", singular),
				Long:   fmt.Sprintf("execute %s on the server", singular),
				Path:   fmt.Sprintf("/%s/{name}/exec.{ext}", plural),
				Method: "POST",
				Params: []string{"name"},
			})
			e.Flags().String(FlagName, "", fmt.Sprintf("name of the %s on the server", singular))
			return e
		}())
		return c
	}("workflow", "workflows", core.WorkflowType))

	return clientCmd
}
