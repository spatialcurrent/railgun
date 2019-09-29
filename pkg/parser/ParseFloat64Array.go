// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package parser

import (
	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-adaptive-functions/pkg/af"
	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-try-get/pkg/gtg"
	rerrors "github.com/spatialcurrent/railgun/pkg/errors"
)

func ParseFloat64Array(obj interface{}, name string) ([]float64, error) {
	expression := gtg.TryGetString(obj, name, "")
	if len(expression) == 0 {
		return make([]float64, 0), nil
	}
	_, arr, err := dfl.ParseCompileEvaluate(expression, dfl.NoVars, dfl.NoContext, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return make([]float64, 0), errors.Wrap(err, (&rerrors.ErrInvalidParameter{Name: name, Value: expression}).Error())
	}
	extent, err := af.ToFloat64Slice.ValidateRun(arr)
	if err != nil {
		return make([]float64, 0), errors.Wrap(err, (&rerrors.ErrInvalidParameter{Name: name, Value: expression}).Error())
	}
	return extent.([]float64), nil
}
