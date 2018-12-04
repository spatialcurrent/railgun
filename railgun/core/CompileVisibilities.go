package core

import (
	"reflect"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

func CompileVisibilities(t reflect.Type) []map[string]struct{} {
	visibilities := make([]map[string]struct{}, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		if str, ok := ft.Tag.Lookup("visibility"); ok {
			_, visibility := dfl.MustParseCompileEvaluate(
				str,
				map[string]interface{}{},
				map[string]interface{}{},
				dfl.DefaultFunctionMap,
				dfl.DefaultQuotes)
			if set, ok := visibility.(map[string]struct{}); ok {
				visibilities = append(visibilities, set)
			} else {
				visibilities = append(visibilities, map[string]struct{}{})
			}
		}
	}
	return visibilities
}
