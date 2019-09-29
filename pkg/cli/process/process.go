// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package process

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	//"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gob"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
	"github.com/spatialcurrent/railgun/pkg/batch"
	"github.com/spatialcurrent/railgun/pkg/cli/input"
	"github.com/spatialcurrent/railgun/pkg/config"
	"github.com/spatialcurrent/railgun/pkg/stream"
	"github.com/spatialcurrent/railgun/pkg/util"
	"github.com/spatialcurrent/viper"
	//"unicode"
)

const (
	CliUse       = "process"
	CliShort     = "process data found in a file, supporting multiple protocols, formats, and compression algorithms"
	CliLong      = "process data found in a file, supporting multiple protocols, formats, and compression algorithms"
	SilenceUsage = true
)

const (
	FlagStream      = "stream"
	FlagInputFormat = input.FlagInputFormat
)

var GO_RAILGUN_COMPRESSION_ALGORITHMS = []string{"none", "bzip2", "gzip", "snappy"}
var GO_RAILGUN_DEFAULT_SALT = "4F56C8C88B38CD8CD96BF8A9724F4BFE"

func processFunction(cmd *cobra.Command, args []string) error {

	// Register gob types
	gob.RegisterTypes()

	v := viper.New()

	err := v.BindPFlags(cmd.Flags())
	if err != nil {
		return errors.Wrap(err, "error binding flags")
	}
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	util.MergeConfigs(v, v.GetStringArray("config-uri"))

	verbose := v.GetBool("verbose")

	if verbose {
		config.PrintViperSettings(v)
	}

	err = CheckProcessConfig(v, args)
	if err != nil {
		return errors.Wrap(err, "error with configuration")
	}

	processConfig := config.NewProcessConfig()
	config.LoadConfigFromViper(processConfig, v)

	if verbose {
		config.PrintConfig(processConfig)
	}

	//var athenaClient *athena.Athena
	var s3Client *s3.S3

	if processConfig.HasAWSResource() {
		awsSession, err := session.NewSessionWithOptions(processConfig.AWSSessionOptions())
		if err != nil {
			return errors.Wrap(err, "error connecting to AWS")
		}

		//if processConfig.HasAthenaStoredQuery() {
		//	athenaClient = athena.New(awsSession)
		//}

		if processConfig.HasS3Bucket() {
			s3Client = s3.New(awsSession)
		}
	}

	logger := gsl.CreateApplicationLogger(&gsl.CreateApplicationLoggerInput{
		ErrorDestination: v.GetString("error-destination"),
		ErrorCompression: processConfig.ErrorCompression,
		ErrorFormat:      processConfig.ErrorFormat,
		InfoDestination:  processConfig.InfoDestination,
		InfoCompression:  processConfig.InfoCompression,
		InfoFormat:       processConfig.InfoFormat,
		Verbose:          processConfig.Verbose,
	})

	start := time.Now()
	if processConfig.Time {
		logger.Info(map[string]interface{}{
			"msg": "started",
			"ts":  start.Format(time.RFC3339),
		})
	}

	if processConfig.Timeout.Seconds() > 0 {
		deadline := time.Now().Add(processConfig.Timeout)
		logger.Debug(fmt.Sprintf("Deadline: %v", deadline))
		go func() {
			for {
				if time.Now().After(deadline) {
					logger.FatalF("program exceeded timeout %v", processConfig.Timeout)
				}
				time.Sleep(15 * time.Second)
			}
		}()
	}

	processConfig.Input.Init()

	inputReader, _, err := grw.ReadFromResource(&grw.ReadFromResourceInput{
		Uri:        processConfig.Input.Uri,
		Alg:        processConfig.Input.Compression,
		Dict:       processConfig.Input.Dictionary,
		BufferSize: processConfig.Input.ReaderBufferSize,
		S3Client:   s3Client,
	})
	if err != nil {
		return errors.Wrapf(err, "error opening resource from uri %q", processConfig.Input.Uri)
	}

	processConfig.Output.Init()

	inputType, err := InitInputType(processConfig.Input.Type, processConfig.Input.Format)
	if err != nil {
		return errors.Wrap(err, "error initializing input type")
	}

	if processConfig.Stream {

		err := stream.ProcessAsStream(&stream.ProcessAsStreamInput{
			Reader:   inputReader,
			Type:     inputType,
			Config:   processConfig,
			S3Client: s3Client,
			Logger:   logger,
		})
		if err != nil {
			return errors.Wrap(err, "error processing as stream")
		}
	} else {
		err := batch.ProcessAsBatch(&batch.ProcessAsBatchInput{
			Reader:   inputReader,
			Type:     inputType,
			Config:   processConfig,
			S3Client: s3Client,
			Logger:   logger,
		})
		if err != nil {
			return errors.Wrap(err, "error processing as batch")
		}
	}

	if processConfig.Time {
		end := time.Now()
		logger.Info(map[string]interface{}{
			"msg":      "ended",
			"ts":       end.Format(time.RFC3339),
			"duration": end.Sub(start).String(),
		})
	}

	logger.Close()

	return nil
}
