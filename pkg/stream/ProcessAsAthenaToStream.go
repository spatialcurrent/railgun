// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package stream

/*

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-reader-writer/pkg/io"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"

	"github.com/spatialcurrent/go-pipe/pkg/pipe"

	"github.com/spatialcurrent/railgun/pkg/config"
)



type ProcessAsAthenaToStreamInput struct {
	Reader   io.Reader
	Type     reflect.Type
	DflNode  dfl.Node
	DflVars  map[string]interface{}
	Config   *config.Process
	S3Client *s3.S3
	Logger   *gsl.Logger
}

func ProcessAsAthenaToStream(input *ProcessAsAthenaToStreamInput) error {

	var wgObjects sync.WaitGroup
	outputObjects := make(chan interface{}, 1000)
	wgObjects.Add(1)

	if processConfig.Verbose {
		logger.Info("Processing as athena to stream.")
		logger.Flush()
	}
	athenaIterator, err := processAthenaInput(
		processConfig.Input.Uri,
		processConfig.Input.Limit,
		processConfig.Temp.Uri,
		processConfig.Output.Format,
		athenaClient,
		logger,
		processConfig.Verbose)
	if err != nil {
		return errors.Wrap(err, "error processing athena input")
	}

	err = handleOutput(
		processConfig.Output,
		dflVars,
		outputObjects,
		processConfig.FileDescriptorLimit,
		&wgObjects,
		s3Client,
		stringer,
		logger,
		processConfig.Verbose)
	if err != nil {
		return errors.Wrap(err, "error procssing output")
	}

	inputCount := 0
	for {

		line, err := athenaIterator.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return errors.Wrap(err, "error from athena iterator")
			}
		}

		//messages <- "processing line: " + string(line)

		inputObject := map[string]interface{}{}
		err = json.Unmarshal(line, &inputObject)
		if err != nil {
			return errors.Wrap(err, "error unmarshalling value from athena results: "+string(line))
		}

		outputObject, err := processObject(inputObject, dflNode, dflVars, stringer)
		if err != nil {
			if err != gss.ErrEmptyRow {
				return errors.Wrap(err, "error processing object")
			}
		} else {
			switch outputObject.(type) {
			case dfl.Null:
			default:
				outputObjects <- outputObject
			}
		}

		inputCount += 1
		if processConfig.Input.Limit > 0 && inputCount >= processConfig.Input.Limit {
			break
		}
	}
	logger.Info("closing outputObjects")
	close(outputObjects)
	logger.Info("waiting for wgObjects")
	wgObjects.Wait()

	if processConfig.Time {
		logger.Info(map[string]interface{}{
			"msg": "ended",
		})
	}
	return nil

}

*/
