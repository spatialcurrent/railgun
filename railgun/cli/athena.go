// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun/athenaiterator"
	"github.com/spatialcurrent/railgun/railgun/config"
	"github.com/spatialcurrent/railgun/railgun/util"
	"github.com/spatialcurrent/viper"
	"io"
	"os"
	"strings"
)

import (
	//"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/athena"
)

func init() {

	athenaCmd := &cobra.Command{
		Use:   "athena",
		Short: "commands for interacting with Athena",
		Long:  "commands for interacting with Athena",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
	rootCmd.AddCommand(athenaCmd)

	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "download results for Athena Query",
		Long:  "download results for Athena Query",
		Run: func(cmd *cobra.Command, args []string) {

			v := viper.New()
			v.BindPFlags(cmd.Flags())
			v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
			v.AutomaticEnv() // set environment variables to overwrite config
			util.MergeConfigs(v, v.GetStringArray("config-uri"))

			athenaConfig := &config.Athena{
				AWS:              &config.AWS{},
				QueryExecutionId: "",
				Output:           &config.Output{},
				Temp:             &config.Temp{},
				Dfl:              &config.Dfl{},
				ErrorDestination: "",
				ErrorCompression: "",
				LogDestination:   "",
				LogCompression:   "",
			}
			config.LoadConfigFromViper(athenaConfig, v)

			errorWriter, err := grw.WriteToResource(athenaConfig.ErrorDestination, "", true, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error creating error writer\n")
				os.Exit(1)
			}

			err = func(errorWriter grw.ByteWriteCloser) error {

				verbose := v.GetBool("verbose")

				if verbose {
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

				awsSession, err := session.NewSessionWithOptions(athenaConfig.AWSSessionOptions())
				if err != nil {
					fmt.Println(errors.Wrap(err, "error connecting to AWS"))
					os.Exit(1)
				}
				//s3Client.New(awsSession)
				athenaClient := athena.New(awsSession)

				outputWriter, err := grw.WriteToResource(athenaConfig.Output.Uri, athenaConfig.Output.Compression, true, nil)
				if err != nil {
					return errors.Wrap(err, "error opening output file")
				}

				athenaIterator, err := athenaiterator.New(athenaClient, &athenaConfig.QueryExecutionId, athenaConfig.Output.Limit)
				if err != nil {
					return errors.Wrap(err, "error creating athena iterator")
				}

				inputCount := 0
				outputCount := 0
				for {

					line, err := athenaIterator.Next()
					if err != nil {
						if err == io.EOF {
							break
						} else {
							return errors.Wrap(err, "error from athena iterator")
						}
					}

					object := map[string]interface{}{}
					err = json.Unmarshal(line, &object)
					if err != nil {
						return errors.Wrap(err, "error unmarshalling value from athena results: "+string(line))
					}

					str, err := athenaConfig.OutputOptions().SerializeString(gss.StringifyMapKeys(object))
					if err != nil {
						return errors.Wrap(err, "error converting input")
					}

					_, err = outputWriter.WriteLine(str)
					if err != nil {
						return errors.Wrap(err, "error writing output line")
					}

					inputCount += 1
					outputCount += 1

					if athenaConfig.Output.Limit > 0 && outputCount >= athenaConfig.Output.Limit {
						break
					}
				}

				err = outputWriter.Close()
				if err != nil {
					return errors.Wrap(err, "error closing output writer")
				}

				return nil

			}(errorWriter)

			if err != nil {
				errorWriter.WriteError(err)
				errorWriter.Flush()
				os.Exit(1)
			}

		},
	}
	downloadCmd.Flags().StringP("query-execution-id", "q", "", "Athena QueryExecutionId")
	// Output Flags
	downloadCmd.Flags().StringP("output-uri", "o", "stdout", "the output uri (a dfl expression itself)")
	downloadCmd.Flags().StringP("output-compression", "", "", "the output compression: "+strings.Join(GO_RAILGUN_COMPRESSION_ALGORITHMS, ", "))
	downloadCmd.Flags().StringP("output-format", "", "json", "the output format: "+strings.Join(gss.Formats, ", "))
	downloadCmd.Flags().StringSliceP("output-header", "", []string{}, "the output header")
	downloadCmd.Flags().StringP("output-passphrase", "", "", "output passphrase for AES-256 encryption")
	downloadCmd.Flags().StringP("output-salt", "", "", "output salt for AES-256 encryption")
	downloadCmd.Flags().IntP("output-limit", "", gss.NoLimit, "maximum number of objects to send to output")
	downloadCmd.Flags().BoolP("output-append", "", false, "append to output files")
	downloadCmd.Flags().Bool("output-buffer-memory", false, "buffer output in memory")
	downloadCmd.Flags().Bool("output-mkdirs", false, "make directories if missing for output files")
	athenaCmd.AddCommand(downloadCmd)

}
