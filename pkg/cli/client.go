// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"

	//"github.com/spatialcurrent/go-try-get/pkg/gtg"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
	"github.com/spatialcurrent/railgun/pkg/core"
	"github.com/spatialcurrent/railgun/pkg/rest"
	"github.com/spatialcurrent/railgun/pkg/util"
	"github.com/spatialcurrent/viper"
)

func initViper(cmd *cobra.Command) *viper.Viper {
	v := viper.New()
	//v.BindPFlags(cmd.InheritedFlags())
	//v.BindPFlags(cmd.PersistentFlags())
	err := v.BindPFlags(cmd.Flags())
	if err != nil {
		panic(err)
	}
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	util.MergeConfigs(v, v.GetStringArray("config-uri"))
	return v
}

func newPostCommand(use string, short string, long string, path string, params []string, inputType reflect.Type) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		Run: func(cmd *cobra.Command, args []string) {

			v := initViper(cmd)

			verbose := v.GetBool(flagVerbose)

			logger := gsl.CreateApplicationLogger(&gsl.CreateApplicationLoggerInput{
				ErrorDestination: v.GetString(flagErrorDestination),
				ErrorCompression: v.GetString(flagErrorCompression),
				ErrorFormat:      v.GetString(flagErrorFormat),
				InfoDestination:  v.GetString(flagInfoDestination),
				InfoCompression:  v.GetString(flagInfoCompression),
				InfoFormat:       v.GetString(flagInfoFormat),
				Verbose:          verbose,
			})

			if verbose {
				printViperSettings(v)
			}

			outputWriter, err := grw.WriteToResource(
				v.GetString(FlagOutputURI),
				v.GetString(FlagOutputCompression),
				v.GetBool(FlagOutputAppend),
				nil,
			)
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error opening output writer"))
			}

			inputObject := map[string]interface{}{}
			for i := 0; i < inputType.NumField(); i++ {
				f := inputType.Field(i)
				if str, ok := f.Tag.Lookup("rest"); ok && str != "" && str != "-" {
					if strings.Contains(str, ",") {
						arr := strings.SplitN(str, ",", 2)
						fieldValue := v.GetString(arr[0])
						if len(fieldValue) > 0 {
							inputObject[arr[0]] = fieldValue
						}
					} else {
						fieldValue := v.GetString(str)
						if len(fieldValue) > 0 {
							inputObject[str] = fieldValue
						}
					}
				}
			}

			u := v.GetString("server") + strings.Replace(path, "{ext}", "json", 1)
			obj := map[string]interface{}{}
			for _, name := range params {
				value := v.GetString(name)
				if len(value) == 0 {
					logger.FatalF("missing %q", name)
				}
				obj[name] = value
				u = strings.Replace(u, "{"+name+"}", value, 1)
			}

			u2, err := url.Parse(u)
			if err != nil {
				logger.Fatal(errors.Wrap(err, fmt.Sprintf("error parsing url ( %q )", u)))
			}

			if strings.Contains(u2.Path, "//") {
				logger.FatalF("url ( %q ) is invalid, beacuse \"//\" transforms POST to GET", u)
			}

			err = rest.MakeRequest(&rest.MakeRequestInput{
				Url:           u,
				Method:        "POST",
				Object:        inputObject,
				Authorization: v.GetString("jwt-token"),
				OutputWriter:  outputWriter,
				OutputFormat:  v.GetString(FlagOutputFormat),
				OutputPretty:  v.GetBool(FlagOutputPretty),
				Logger:        logger,
			})
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error making request"))
			}

		},
	}
}

func newRestCommand(use string, short string, long string, path string, method string, params []string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		Run: func(cmd *cobra.Command, args []string) {

			v := initViper(cmd)

			verbose := v.GetBool(flagVerbose)

			logger := gsl.CreateApplicationLogger(&gsl.CreateApplicationLoggerInput{
				ErrorDestination: v.GetString(flagErrorDestination),
				ErrorCompression: v.GetString(flagErrorCompression),
				ErrorFormat:      v.GetString(flagErrorFormat),
				InfoDestination:  v.GetString(flagInfoDestination),
				InfoCompression:  v.GetString(flagInfoCompression),
				InfoFormat:       v.GetString(flagInfoFormat),
				Verbose:          verbose,
			})

			if verbose {
				printViperSettings(v)
			}

			outputWriter, err := grw.WriteToResource(
				v.GetString(FlagOutputURI),
				v.GetString(FlagOutputCompression),
				v.GetBool(FlagOutputAppend),
				nil,
			)
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error opening output writer"))
			}

			u := v.GetString("server") + strings.Replace(path, "{ext}", "json", 1)
			obj := map[string]interface{}{}
			for _, name := range params {
				value := v.GetString(name)
				if len(value) == 0 {
					logger.FatalF("missing %q", name)
				}
				obj[name] = value
				u = strings.Replace(u, "{"+name+"}", value, 1)
			}

			u2, err := url.Parse(u)
			if err != nil {
				logger.Fatal(errors.Wrap(err, fmt.Sprintf("error parsing url ( %q )", u)))
			}

			if strings.Contains(u2.Path, "//") {
				logger.FatalF("url ( %q ) is invalid, beacuse \"//\" transforms POST to GET", u)
			}

			err = rest.MakeRequest(&rest.MakeRequestInput{
				Url:           u,
				Method:        method,
				Object:        obj,
				Authorization: v.GetString("jwt-token"),
				OutputWriter:  outputWriter,
				OutputFormat:  v.GetString(FlagOutputFormat),
				OutputPretty:  v.GetBool(FlagOutputPretty),
				Logger:        logger,
			})
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error making request"))
			}

		},
	}
}

func init() {

	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "client commands for interacting with Railgun Server",
		Long:  "client commands for interacting with Railgun Server",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Usage()
			if err != nil {
				panic(err)
			}
		},
	}

	rootCmd.AddCommand(clientCmd)
	clientCmd.PersistentFlags().String("jwt-token", "", "The JWT token")
	clientCmd.PersistentFlags().StringP("server", "s", "http://localhost:8080", "the \"server\" location")
	InitOutputFlags(clientCmd.PersistentFlags(), "json")

	authenticateCmd := &cobra.Command{
		Use:   "authenticate",
		Short: "authenticate with Railgun Server",
		Long:  "authenticate with Railgun Server",
		Run: func(cmd *cobra.Command, args []string) {

			v := initViper(cmd)

			verbose := v.GetBool(flagVerbose)

			logger := gsl.CreateApplicationLogger(&gsl.CreateApplicationLoggerInput{
				ErrorDestination: v.GetString(flagErrorDestination),
				ErrorCompression: v.GetString(flagErrorCompression),
				ErrorFormat:      v.GetString(flagErrorFormat),
				InfoDestination:  v.GetString(flagInfoDestination),
				InfoCompression:  v.GetString(flagInfoCompression),
				InfoFormat:       v.GetString(flagInfoFormat),
				Verbose:          verbose,
			})

			if verbose {
				printViperSettings(v)
			}

			outputWriter, err := grw.WriteToResource(
				v.GetString(FlagOutputURI),
				v.GetString(FlagOutputCompression),
				v.GetBool(FlagOutputAppend),
				nil,
			)
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error opening output writer"))
			}

			inputObject := map[string]interface{}{
				"username": v.GetString("username"),
				"password": v.GetString("password"),
			}

			u := v.GetString("server") + "/authenticate." + v.GetString("output-format")

			u2, err := url.Parse(u)
			if err != nil {
				logger.Fatal(errors.Wrap(err, fmt.Sprintf("error parsing url ( %q )", u)))
			}

			if strings.Contains(u2.Path, "//") {
				logger.FatalF("url ( %q ) is invalid, beacuse \"//\" transforms POST to GET", u)
			}

			err = rest.MakeRequest(&rest.MakeRequestInput{
				Url:    u,
				Method: "POST",
				Object: inputObject,
				// Authorization: v.GetString("jwt-token"),
				OutputWriter: outputWriter,
				OutputFormat: v.GetString(FlagOutputFormat),
				OutputPretty: v.GetBool(FlagOutputPretty),
				Logger:       logger,
			})
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error making request"))
			}

		},
	}
	authenticateCmd.Flags().String("username", "", "username")
	authenticateCmd.Flags().String("password", "", "password")
	clientCmd.AddCommand(authenticateCmd)

	// Workspaces
	workspacesCmd := &cobra.Command{
		Use:   "workspaces",
		Short: "interact with workspaces on Railgun Server",
		Long:  "interact with workspaces on Railgun Server",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Usage()
			if err != nil {
				panic(err)
			}
		},
	}
	clientCmd.AddCommand(workspacesCmd)
	initRestCommands(workspacesCmd, "/workspaces", "workspace", "workspace", core.WorkspaceType)

	// Data Stores
	datastoresCmd := &cobra.Command{
		Use:   "datastores",
		Short: "interact with datastores on Railgun Server",
		Long:  "interact with datastores on Railgun Server",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Usage()
			if err != nil {
				panic(err)
			}
		},
	}
	clientCmd.AddCommand(datastoresCmd)
	initRestCommands(datastoresCmd, "/datastores", "data store", "data stores", core.DataStoreType)

	// Layers
	layersCmd := &cobra.Command{
		Use:   "layers",
		Short: "interact with layers on Railgun Server",
		Long:  "interact with layers on Railgun Server",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Usage()
			if err != nil {
				panic(err)
			}
		},
	}
	clientCmd.AddCommand(layersCmd)
	initRestCommands(layersCmd, "/layers", "layer", "layers", core.LayerType)

	// Processes
	processesCmd := &cobra.Command{
		Use:   "processes",
		Short: "interact with processes on Railgun Server",
		Long:  "interact with processes on Railgun Server",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Usage()
			if err != nil {
				panic(err)
			}
		},
	}
	clientCmd.AddCommand(processesCmd)
	initRestCommands(processesCmd, "/processes", "process", "processes", core.ProcessType)

	// Services
	servicesCmd := &cobra.Command{
		Use:   "services",
		Short: "interact with services on Railgun Server",
		Long:  "interact with services on Railgun Server",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Usage()
			if err != nil {
				panic(err)
			}
		},
	}
	clientCmd.AddCommand(servicesCmd)
	initRestCommands(servicesCmd, "/services", "service", "services", core.ServiceType)
	serviceExecCmd := newPostCommand(
		"exec",
		"execute a service on the Railgun Server with the given input",
		"execute a service on the Railgun Server with the given input",
		"/services/{name}/exec.{ext}",
		[]string{"name"},
		core.JobType)
	servicesCmd.AddCommand(serviceExecCmd)
	initFlags(serviceExecCmd, core.JobType)

	// Jobs
	jobsCmd := &cobra.Command{
		Use:   "jobs",
		Short: "interact with jobs on Railgun Server",
		Long:  "interact with jobs on Railgun Server",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Usage()
			if err != nil {
				panic(err)
			}
		},
	}
	clientCmd.AddCommand(jobsCmd)
	initRestCommands(jobsCmd, "/jobs", "job", "jobs", core.JobType)
	jobExecCmd := newRestCommand(
		"exec",
		fmt.Sprintf("execute %s on Railgun Server", "job"),
		fmt.Sprintf("execute %s on Railgun Server", "job"),
		"/jobs/{name}/exec.{ext}",
		"POST",
		[]string{"name"})
	jobExecCmd.Flags().String("name", "", fmt.Sprintf("name of %s on Railgun Server", "job"))
	jobsCmd.AddCommand(jobExecCmd)

	// Workflows
	workflowsCmd := &cobra.Command{
		Use:   "workflows",
		Short: "interact with workflows on Railgun Server",
		Long:  "interact with workflows on Railgun Server",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Usage()
			if err != nil {
				panic(err)
			}
		},
	}
	clientCmd.AddCommand(workflowsCmd)
	initRestCommands(workflowsCmd, "/workflows", "workflow", "workflows", core.WorkflowType)
	workflowExecCmd := newRestCommand(
		"exec",
		fmt.Sprintf("execute %s on Railgun Server", "workflow"),
		fmt.Sprintf("execute %s on Railgun Server", "workflow"),
		"/workflows/{name}/exec.{ext}",
		"POST",
		[]string{"name"})
	workflowExecCmd.Flags().String("name", "", fmt.Sprintf("name of %s on Railgun Server", "workflow"))
	workflowsCmd.AddCommand(workflowExecCmd)

}

func initRestCommands(parentCmd *cobra.Command, baseurl string, singular string, plural string, inputType reflect.Type) {

	addCmd := newPostCommand(
		"add",
		fmt.Sprintf("add %s to Railgun Server", singular),
		fmt.Sprintf("add %s to Railgun Server", singular),
		baseurl+".{ext}",
		[]string{},
		inputType)
	initFlags(addCmd, inputType)

	getCmd := newRestCommand(
		"get",
		fmt.Sprintf("get %s on Railgun Server", singular),
		fmt.Sprintf("get %s on Railgun Server", singular),
		baseurl+"/{name}.{ext}",
		"GET",
		[]string{"name"})
	getCmd.Flags().String("name", "", fmt.Sprintf("name of %s on Railgun Server", singular))

	updateCmd := newPostCommand(
		"update",
		fmt.Sprintf("update %s on Railgun Server", singular),
		fmt.Sprintf("update %s on Railgun Server", singular),
		baseurl+"/{name}.{ext}",
		[]string{"name"},
		inputType)
	initFlags(updateCmd, inputType)

	deleteCmd := newRestCommand(
		"delete",
		fmt.Sprintf("delete %s on Railgun Server", singular),
		fmt.Sprintf("delete %s on Railgun Server", singular),
		baseurl+"/{name}.{ext}",
		"DELETE",
		[]string{"name"})
	deleteCmd.Flags().String("name", "", fmt.Sprintf("name of %s on Railgun Server", singular))

	listCmd := newRestCommand(
		"list",
		fmt.Sprintf("list %s on Railgun Server", plural),
		fmt.Sprintf("list %s on Railgun Server", plural),
		baseurl+".{ext}",
		"GET",
		[]string{})

	parentCmd.AddCommand(addCmd, getCmd, updateCmd, deleteCmd, listCmd)

}

func initFlags(c *cobra.Command, t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if str, ok := f.Tag.Lookup("rest"); ok && str != "" && str != "-" {
			//t := f.Type
			if strings.Contains(str, ",") {
				arr := strings.SplitN(str, ",", 2)
				c.Flags().String(arr[0], "", arr[1])
			} else {
				c.Flags().String(str, "", "")
			}
			// use dfl for arrays
			/*if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
				if strings.Contains(str, ",") {
					arr := strings.SplitN(str, ",", 2)
					c.Flags().StringArray(arr[0], []string{}, arr[1])
				} else {
					c.Flags().StringArray(str, []string{}, "")
				}
			} else {
				if strings.Contains(str, ",") {
					arr := strings.SplitN(str, ",", 2)
					c.Flags().String(arr[0], "", arr[1])
				} else {
					c.Flags().String(str, "", "")
				}
			}*/
		}
	}
}
