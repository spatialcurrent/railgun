// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package parser

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-adaptive-functions/af"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-try-get/gtg"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
)

func ParseStringArray(obj interface{}, name string) ([]string, error) {
	expression := gtg.TryGetString(obj, name, "")
	if len(expression) == 0 {
		return make([]string, 0), nil
	}
	_, arr, err := dfl.ParseCompileEvaluate(expression, dfl.NoVars, dfl.NoContext, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return make([]string, 0), errors.Wrap(err, (&rerrors.ErrInvalidParameter{Name: name, Value: expression}).Error())
	}
	strs, err := af.ToStringArray.ValidateRun([]interface{}{arr})
	if err != nil {
		return make([]string, 0), errors.Wrap(err, (&rerrors.ErrInvalidParameter{Name: name, Value: expression}).Error())
	}
	return strs.([]string), nil
}
