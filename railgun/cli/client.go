// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun/core"
	"github.com/spatialcurrent/railgun/railgun/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
)

func initViper(cmd *cobra.Command) *viper.Viper {
	v := viper.New()
	//v.BindPFlags(cmd.InheritedFlags())
	//v.BindPFlags(cmd.PersistentFlags())
	v.BindPFlags(cmd.Flags())
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	util.MergeConfigs(v, v.GetStringArray("config-uri"))
	return v
}

func printConfig(v *viper.Viper) {
	fmt.Println("=================================================")
	fmt.Println("Configuration:")
	fmt.Println("-------------------------------------------------")
	str, err := gss.SerializeString(v.AllSettings(), "properties", []string{}, -1)
	if err != nil {
		fmt.Println("error getting all settings")
		os.Exit(1)
	}
	fmt.Println(str)
	fmt.Println("=================================================")
}

type RequestInput struct {
	Url    string
	Method string
	Object interface{}
	Format string
}

//func handlePost(url string, inputObject interface{}, outputFormat string, outputWriter grw.ByteWriteCloser, errorWriter grw.ByteWriteCloser) error {

func MakeRequest(input *RequestInput, outputWriter grw.ByteWriteCloser, errorWriter grw.ByteWriteCloser, verbose bool) error {

	var req *http.Request
	if input.Method != "GET" {
		inputBytes, err := gss.SerializeBytes(input.Object, "json", []string{}, gss.NoLimit)
		if err != nil {
			return err
		}
		if verbose {
			fmt.Println("Body:\n", string(inputBytes))
		}
		r, err := http.NewRequest(input.Method, input.Url, bytes.NewBuffer(inputBytes))
		r.Header.Set("Content-Type", "application/json")
		if err != nil {
			return err
		}
		req = r
	} else {
		r, err := http.NewRequest(input.Method, input.Url, nil)
		if err != nil {
			return err
		}
		req = r
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	respType, err := gss.GetType(respBytes, "json")
	if err != nil {
		return err
	}

	respObject, err := gss.DeserializeBytes(respBytes, "json", []string{}, "", false, gss.NoLimit, respType, false)
	if err != nil {
		return err
	}

	outputBytes, err := gss.SerializeBytes(respObject, input.Format, []string{}, gss.NoLimit)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(outputBytes))
	}

	outputWriter.Write(outputBytes)
	outputWriter.WriteString("\n")
	outputWriter.Flush()

	return nil
}

func handleList(url string, outputWriter grw.ByteWriteCloser, errorWriter grw.ByteWriteCloser) {

	brc, _, err := grw.ReadHTTPFile(url, "", false)
	if err != nil {
		errorWriter.WriteError(err)
		errorWriter.Flush()
		os.Exit(1)
	}

	outputBytes, err := brc.ReadAllAndClose()
	if err != nil {
		errorWriter.WriteError(err)
		errorWriter.Flush()
		os.Exit(1)
	}

	outputWriter.Write(outputBytes)
	outputWriter.WriteString("\n")
	outputWriter.Flush()
}

func newPostCommand(use string, short string, long string, path string, inputType reflect.Type) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		Run: func(cmd *cobra.Command, args []string) {

			errorWriter, err := grw.WriteToResource("stderr", "", true, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error creating error writer\n")
				os.Exit(1)
			}

			err = func(errorWriter grw.ByteWriteCloser) error {
				v := initViper(cmd)

				verbose := v.GetBool("verbose")

				if verbose {
					printConfig(v)
				}

				outputWriter, err := grw.WriteToResource("stdout", "", true, nil)
				if err != nil {
					return errors.Wrap(err, "error opening output file")
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

				u2, err := url.Parse(u)
				if err != nil {
					return err
				}

				if strings.Contains(u2.Path, "//") {
					return errors.New("url is invalid: " + u)
				}

				return MakeRequest(&RequestInput{
					Url:    u,
					Method: "POST",
					Object: inputObject,
					Format: v.GetString("output-format"),
				}, outputWriter, errorWriter, v.GetBool("verbose"))

			}(errorWriter)

			if err != nil {
				errorWriter.WriteError(err)
				errorWriter.Flush()
				os.Exit(1)
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

			errorWriter, err := grw.WriteToResource("stderr", "", true, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error creating error writer\n")
				os.Exit(1)
			}

			err = func(errorWriter grw.ByteWriteCloser) error {
				v := initViper(cmd)

				verbose := v.GetBool("verbose")

				if verbose {
					printConfig(v)
				}

				outputWriter, err := grw.WriteToResource("stdout", "", true, nil)
				if err != nil {
					return errors.Wrap(err, "error opening output file")
				}

				u := v.GetString("server") + strings.Replace(path, "{ext}", "json", 1)
				obj := map[string]interface{}{}
				for _, name := range params {
					value := v.GetString(name)
					if len(value) == 0 {
						return errors.New("missing " + name)
					}
					obj[name] = value
					u = strings.Replace(u, "{"+name+"}", value, 1)
				}

				u2, err := url.Parse(u)
				if err != nil {
					return err
				}

				if strings.Contains(u2.Path, "//") {
					return errors.New("url is invalid: " + u)
				}

				return MakeRequest(&RequestInput{
					Url:    u,
					Method: method,
					Object: obj,
					Format: v.GetString("output-format"),
				}, outputWriter, errorWriter, v.GetBool("verbose"))

			}(errorWriter)

			if err != nil {
				errorWriter.WriteError(err)
				errorWriter.Flush()
				os.Exit(1)
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
			cmd.Usage()
		},
	}

	rootCmd.AddCommand(clientCmd)
	clientCmd.PersistentFlags().StringP("server", "s", "http://localhost:8080", "the \"server\" location")
	clientCmd.PersistentFlags().StringP("output-format", "f", "json", "the output format: "+strings.Join(gss.Formats, ", "))

	// Workspaces
	workspacesCmd := &cobra.Command{
		Use:   "workspaces",
		Short: "interact with workspaces on Railgun Server",
		Long:  "interact with workspaces on Railgun Server",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
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
			cmd.Usage()
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
			cmd.Usage()
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
			cmd.Usage()
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
			cmd.Usage()
		},
	}
	clientCmd.AddCommand(servicesCmd)
	initRestCommands(servicesCmd, "/services", "service", "services", core.ServiceType)
	servicesExecCmd := newPostCommand(
		"exec",
		"execute a service on the Railgun Server with the given input",
		"execute a service on the Railgun Server with the given input",
		"/services/exec.{ext}",
		core.JobType)
	servicesCmd.AddCommand(servicesExecCmd)
	initFlags(servicesExecCmd, core.JobType)

	// Jobs
	jobsCmd := &cobra.Command{
		Use:   "jobs",
		Short: "interact with jobs on Railgun Server",
		Long:  "interact with jobs on Railgun Server",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
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
			cmd.Usage()
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

	parentCmd.AddCommand(addCmd, getCmd, deleteCmd, listCmd)

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
