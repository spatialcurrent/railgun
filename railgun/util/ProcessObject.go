// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package util

/*
import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-simple-serializer/gss"

	"github.com/spatialcurrent/railgun/railgun/config"
)

func ProcessObject(inputBytes []byte, inputConfig *config.Input, dflExpression string, dflVars map[string]interface{}, dflUri string, verbose bool) (interface{}, error) {

	dflNode, err := inputConfig.Dfl.Node()
	if err != nil {
		return "", errors.Wrap(err, "error parsing")
	}

	inputType, err := gss.GetType(inputBytes, inputConfig.Format)
	if err != nil {
		return "", errors.Wrap(err, "error getting type for input")
	}

	options := inputConfig.Options()
	options.Type = inputType
	inputObject, err := options.DeserializeBytes(inputBytes, verbose)
	if err != nil {
		return "", errors.Wrap(err, "error deserializing input using format "+inputConfig.Format)
	}

	if dflNode != nil {
		_, outputObject, err := dflNode.Evaluate(dflVars, inputObject, dfl.DefaultFunctionMap, []string{"'", "\"", "`"})
		if err != nil {
			return "", errors.Wrap(err, "error evaluating filter")
		}
		return outputObject, nil
	}

	return inputObject, nil

}
*/
