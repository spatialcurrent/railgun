// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spatialcurrent/go-dfl/pkg/dfl"
)

//"filter(@, '@visibility in $authorizations') | dict(map(@, '[((`,` in @rest) ? first(split(@rest, `,`)) : @rest'), @value]'))"

var intersects = dfl.MustParseCompile("intersects(set($a), set($b))")

func Map(object interface{}, authorizations []string) map[string]interface{} {

	t := reflect.TypeOf(object)
	v := reflect.ValueOf(object)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	m := map[string]interface{}{}

	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)
		visibilities := UserVisibilities[i]
		fmt.Println("Visibilites:", visibilities)
		if len(visibilities) > 0 {
			vars := map[string]interface{}{"a": UserVisibilities[i], "b": authorizations}
			fmt.Println("vars:", vars)
			_, authorized, err := dfl.EvaluateBool(intersects, vars, nil, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
			fmt.Println("Authorized:", authorized)
			fmt.Println("Err:", err)
			if err == nil && authorized {
				if str, ok := ft.Tag.Lookup("rest"); ok && str != "" && str != "-" {
					if strings.Contains(str, ",") {
						m[strings.SplitN(str, ",", 2)[0]] = fv.Interface()
					} else {
						m[str] = fv.Interface()
					}
				}
			}
		}
	}

	return m
}
