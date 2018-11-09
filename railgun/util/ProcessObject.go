// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package util

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-simple-serializer/gss"
)

func ProcessObject(inputBytes []byte, inputFormat string, inputHeader []string, inputComment string, inputLazyQuotes bool, inputLimit int, dflExpression string, dflVars map[string]interface{}, dflUri string, verbose bool) (interface{}, error) {

	funcs := dfl.NewFuntionMapWithDefaults()

	dflNode, err := ParseDfl(dflUri, dflExpression)
	if err != nil {
		return "", errors.Wrap(err, "error parsing")
	}

	if verbose && dflNode != nil {
		dflNodeAsYaml, err := gss.SerializeString(dflNode.Map(), "yaml", []string{}, 0)
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

	inputObject, err := gss.DeserializeBytes(inputBytes, inputFormat, inputHeader, inputComment, inputLazyQuotes, inputLimit, inputType, verbose)
	if err != nil {
		return "", errors.Wrap(err, "error deserializing input using format "+inputFormat)
	}

	if dflNode != nil {
		_, outputObject, err := dflNode.Evaluate(dflVars, inputObject, funcs, []string{"'", "\"", "`"})
		if err != nil {
			return "", errors.Wrap(err, "error evaluating filter")
		}
		return outputObject, nil
	}

	return inputObject, nil

}
