// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package parser

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-try-get/gtg"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"reflect"
)

func ParseMap(obj interface{}, name string) (map[string]interface{}, error) {
	expression := gtg.TryGetString(obj, name, "")
	if len(expression) == 0 {
		return make(map[string]interface{}, 0), nil
	}
	_, m, err := dfl.ParseCompileEvaluateMap(expression, dfl.NoVars, dfl.NoContext, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return make(map[string]interface{}, 0), errors.Wrap(err, (&rerrors.ErrInvalidParameter{Name: name, Value: expression}).Error())
	}
	if reflect.TypeOf(m).Kind() == reflect.Map {
		if reflect.ValueOf(m).Len() == 0 {
			return map[string]interface{}{}, nil
		}
	}
	return gss.StringifyMapKeys(m).(map[string]interface{}), err
}
