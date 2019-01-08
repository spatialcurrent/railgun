// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package util

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-simple-serializer/gss"
)

func ProcessInput(inputBytes []byte, inputFormat string, inputHeader []string, inputComment string, inputLazyQuotes bool, inputSkipLines int, inputLimit int, dflExpression string, dflVars map[string]interface{}, dflUri string, outputFormat string, outputHeader []string, outputLimit int, verbose bool) (string, error) {

	if len(outputFormat) > 0 {

		dflNode, err := ParseDfl(dflUri, dflExpression)
		if err != nil {
			return "", errors.Wrap(err, "error parsing")
		}

		inputType, err := gss.GetType(inputBytes, inputFormat)
		if err != nil {
			return "", errors.Wrap(err, "error getting type for input")
		}

		inputObject, err := gss.DeserializeBytes(inputBytes, inputFormat, inputHeader, inputComment, inputLazyQuotes, inputSkipLines, inputLimit, inputType, false, verbose)
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

		outputString, err := gss.SerializeString(
			gss.StringifyMapKeys(outputObject),
			outputFormat,
			outputHeader,
			outputLimit)
		if err != nil {
			return "", errors.Wrap(err, "error converting output")
		}

		return outputString, nil
	}

	return string(inputBytes), nil

}
