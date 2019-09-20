// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package stream

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/pkg/dfl"
)

// NewTransformFunction creates a new transform function for use with go-pipe.
func NewTransformFunction(dflNode dfl.Node, dflVars map[string]interface{}) func(inputObject interface{}) (interface{}, error) {
	return func(inputObject interface{}) (interface{}, error) {
		_, outputObject, err := dflNode.Evaluate(dflVars, inputObject, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
		if err != nil {
			return nil, errors.Wrap(err, "error evaluating dfl filter")
		}
		if _, ok := outputObject.(dfl.Null); ok {
			return nil, nil
		}
		return outputObject, nil
	}
}
