// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package util

import (
	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
)

func ProcessInput(inputBytes []byte, inputFormat string, inputHeader []interface{}, inputComment string, inputLazyQuotes bool, inputSkipLines int, inputLimit int, dflExpression string, dflVars map[string]interface{}, dflUri string, outputFormat string, outputHeader []interface{}, outputLimit int) (string, error) {

	if len(outputFormat) > 0 {

		dflNode, err := ParseDfl(dflUri, dflExpression)
		if err != nil {
			return "", errors.Wrap(err, "error parsing")
		}

		inputObject, err := gss.DeserializeBytes(&gss.DeserializeBytesInput{
			Bytes:         inputBytes,
			Format:        inputFormat,
			Header:        inputHeader,
			Comment:       inputComment,
			LazyQuotes:    inputLazyQuotes,
			SkipLines:     inputSkipLines,
			Limit:         inputLimit,
			LineSeparator: "\n",
		})
		if err != nil {
			return "", errors.Wrap(err, "error deserializing input using format "+inputFormat)
		}

		var outputObject interface{}
		if dflNode != nil {
			_, object, err := dflNode.Evaluate(dflVars, inputObject, dfl.DefaultFunctionMap, []string{"'", "\"", "`"})
			if err != nil {
				return "", errors.Wrap(err, "error evaluating filter")
			}
			outputObject = object
		} else {
			outputObject = inputObject
		}

		outputBytes, err := gss.SerializeBytes(&gss.SerializeBytesInput{
			Object:            outputObject,
			Format:            outputFormat,
			Header:            outputHeader,
			Limit:             outputLimit,
			LineSeparator:     "\n",
			KeyValueSeparator: "=",
		})
		if err != nil {
			return "", errors.Wrap(err, "error converting output")
		}

		return string(outputBytes), nil
	}

	return string(inputBytes), nil

}
