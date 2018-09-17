// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package railgun

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	//"github.com/spatialcurrent/go-reader/reader"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	//"os"
	//"strings"
)

func ProcessInput(inputBytes []byte, inputFormat string, inputHeader []string, inputComment string, inputLimit int, dflExpression string, dflUri string, outputFormat string, verbose bool) (string, error) {

	if len(outputFormat) > 0 {

		funcs := dfl.NewFuntionMapWithDefaults()

		dflNode, err := ParseDfl(dflUri, dflExpression)
		if err != nil {
			return "", errors.Wrap(err, "error parsing")
		}

		if verbose && dflNode != nil {
			dflNodeAsYaml, err := gss.Serialize(dflNode.Map(), "yaml")
			if err != nil {
				return "", errors.Wrap(err, "error dumping dflNode as yaml to stdout")
			}
			fmt.Println(dflNodeAsYaml)
			fmt.Println(dflNode.Dfl(dfl.DefaultQuotes, true, 0))
		}

		inputType, err := gss.GetType(inputBytes, inputFormat)
		if err != nil {
			return "", errors.Wrap(err, "error getting type for input")
		}

		inputObject, err := gss.Deserialize(string(inputBytes), inputFormat, inputHeader, inputComment, inputLimit, inputType, verbose)
		if err != nil {
			return "", errors.Wrap(err, "error deserializing input using format "+inputFormat)
		}

		var output interface{}
		if dflNode != nil {
			_, o, err := dflNode.Evaluate(map[string]interface{}{}, inputObject, funcs, []string{"'", "\"", "`"})
			if err != nil {
				return "", errors.Wrap(err, "error evaluating filter")
			}
			output = o
		} else {
			output = inputObject
		}

		output = gss.StringifyMapKeys(output)

		outputString, err := gss.Serialize(output, outputFormat)
		if err != nil {
			return "", errors.Wrap(err, "error converting output")
		}

		return outputString, nil
	}

	return string(inputBytes), nil

}
