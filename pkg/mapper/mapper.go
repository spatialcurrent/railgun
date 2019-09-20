// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package mapper

import (
	"context"
	"reflect"
	"strings"

	"github.com/spatialcurrent/go-dfl/pkg/dfl"
)

type named interface {
	GetName() string
}

func MarshalMap(object interface{}) map[string]interface{} {
	v := reflect.ValueOf(object)
	for reflect.TypeOf(v.Interface()).Kind() == reflect.Ptr {
		v = v.Elem()
	}
	c := v.Interface()

	if m, ok := c.(map[string]interface{}); ok {
		return m
	}

	v = reflect.ValueOf(c) // sets value to concerete type
	t := v.Type()

	if t.Kind() == reflect.Struct {

		m := map[string]interface{}{}
		for i := 0; i < v.NumField(); i++ {
			f := t.Field(i)
			fv := v.Field(i).Interface()
			fvt := reflect.TypeOf(fv)

			empty := false
			if fvt != nil {
				switch fvt.Kind() {
				case reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr:
					empty = (!reflect.ValueOf(fv).IsValid()) || reflect.ValueOf(fv).IsNil()
				case reflect.Array, reflect.String, reflect.Map, reflect.Slice:
					empty = reflect.ValueOf(fv).Len() == 0
				default:
					empty = false
				}
			} else {
				empty = true
			}

			key := f.Name
			omitempty := false

			if tag, ok := f.Tag.Lookup("map"); ok && tag != "" {
				if tag != "-" {
					if strings.Contains(tag, ",") {
						parts := strings.Split(tag, ",")
						attrs := map[string]struct{}{}
						for _, p := range parts {
							attrs[p] = struct{}{}
						}
						key = parts[0]
						if _, ok := attrs["omitempty"]; ok {
							omitempty = true
						}
					} else {
						key = tag
					}
				}
			}

			if s, ok := fv.([]float64); ok {
				if empty && omitempty {
					continue
				}
				m[key] = dfl.Literal{Value: s}.Dfl(dfl.DefaultQuotes, false, 0)
				continue
			}

			if !empty {

				if n, ok := fv.(dfl.Node); ok {
					fv = n.Dfl(dfl.DefaultQuotes, false, 0)
				}

				if n, ok := fv.(named); ok {
					fv = n.GetName()
				}

				if s, ok := fv.([]string); ok {
					a := make([]dfl.Node, 0)
					for _, x := range s {
						a = append(a, dfl.Literal{Value: x})
					}
					fv = dfl.Array{Nodes: a}.Dfl(dfl.DefaultQuotes, false, 0)
				}

				if m, ok := fv.(map[string]interface{}); ok {
					dict := map[dfl.Node]dfl.Node{}
					for k, v := range m {
						dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
					}
					fv = dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
				}

				m[key] = fv

			} else {
				m[key] = nil
			}
		}
		return m
	}

	return map[string]interface{}{}
}

func MarshalMapWithContext(ctx context.Context, object interface{}) map[string]interface{} {
	return MarshalMap(object)
}
