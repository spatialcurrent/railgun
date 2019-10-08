// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	"github.com/spatialcurrent/go-dfl/pkg/dfl"
)

func parseCompileEvaluateMap(str string) (interface{}, error) {
	_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	return m, err
}
