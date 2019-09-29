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

type ProcessAsSinkToStreamInput struct {
	Reader   io.Reader
	Type     reflect.Type
	DflNode  dfl.Node
	DflVars  map[string]interface{}
	Config   *config.Process
	S3Client *s3.S3
	Logger   *gsl.Logger
}

func ProcessAsSinkToStream(input *ProcessAsSinkToStreamInput) error {
	if processConfig.Verbose {
		logger.Info("Processing as sink to stream.")
		logger.Flush()
	}

	var wgObjects sync.WaitGroup

	outputObjects := make(chan interface{}, 1000)

	wgObjects.Add(1)
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

	p = pipe.NewBuilder().OutputLimit(processConfig.Output.Limit)

	//
	// Input
	//

	p = p.InputLimit(processConfig.Input.Limit).Input(NewSink(func() (interface{}, error) {
		inputBytes, err := util.DecryptReader(inputReader, processConfig.Input.Passphrase, processConfig.Input.Salt)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding input")
		}

		inputType, err := gss.GetType(inputBytes, processConfig.Input.Format)
		if err != nil {
			return nil, errors.Wrap(err, "error getting type for input")
		}

		if !(inputType.Kind() == reflect.Array || inputType.Kind() == reflect.Slice) {
			return nil, errors.New("input type cannot be streamed as it is not an array or slice but " + fmt.Sprint(inputType))
		}

		inputObjects, err := gss.DeserializeBytes(&gss.DeserializeBytesInput{
			Bytes:         inputBytes,
			Format:        processConfig.Input.Format,
			Header:        processConfig.Input.Header,
			Comment:       processConfig.Input.Comment,
			LazyQuotes:    processConfig.Input.LazyQuotes,
			SkipLines:     processConfig.Input.SkipLines,
			Limit:         processConfig.Input.Limit,
			Type:          inputType,
			LineSeparator: processConfig.Input.LineSeparator,
		})
		if err != nil {
			return nil, errors.Wrap(err, "error deserializing input using format "+processConfig.Input.Format)
		}
		return inputObjects, nil
	}))

	//
	// Transform
	//

	if dflNode != nil {
		p = p.Transform(stream.NewTransformFunction(dflNode, dflVars))
	}

	//
	// Output
	//

	p = p.Output(pipe.NewChannelWriter(outputObjects)).CloseOutput(true)

	err = p.Run()
	if err != nil {
		return errors.Wrap(err, "error processing values as sink to stream")
	}

	wgObjects.Wait()

	if processConfig.Time {
		logger.Info(map[string]interface{}{
			"msg": "ended",
		})
	}
	return nil // exits function
}
*/
