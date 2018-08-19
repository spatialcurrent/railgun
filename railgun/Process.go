// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package railgun

import (
  "fmt"
  "math"
  "reflect"
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-simple-serializer/gss"
)

// Process is a function for processing data.
func Process(input_object interface{}, object_path string, filter dfl.Node, funcs dfl.FunctionMap, max_count int, output_path string) (interface{}, error) {

  var object interface{}

  if len(object_path) > 0 {
    o, err := dfl.Extract(object_path, input_object)
    if err != nil {
      return object, errors.Wrap(err, "error extracting object using path "+object_path)
    }
    object = o
  } else {
    object = input_object
  }

  if filter != nil {
    s := reflect.ValueOf(object)
    if s.Kind() != reflect.Slice {
      return object, errors.New("Object at path " + object_path + " is not of kind slice.")
    }
    output_slice := make([]interface{}, 0)
    for i := 0; i < s.Len(); i++ {
      r := s.Index(i).Elem()
      if r.Kind() != reflect.Map {
        return object, errors.New("Row is not of kind map, but " + fmt.Sprint(r.Kind()))
      }
      m := map[string]interface{}{}
      for _, k := range r.MapKeys() {
        m[fmt.Sprint(k)] = gss.StringifyMapKeys(r.MapIndex(k).Interface())
      }
      valid, err := dfl.EvaluateBool(filter, m, funcs)
      if err != nil {
        return object, errors.Wrap(err, "Error evaluating object "+fmt.Sprint(m))
      }
      if valid {
        if len(output_path) > 0 {
          x, err := dfl.Extract(output_path, m)
          if err != nil {
            return object, errors.Wrap(err, "error extracting object using path "+output_path)
          }
          output_slice = append(output_slice, x)
        } else {
          output_slice = append(output_slice, m)
        }
      }
      if max_count != -1 && len(output_slice) == max_count {
        break
      }
    }
    return output_slice, nil
  }

  if max_count != -1 {
    s := reflect.ValueOf(object)
    if s.Kind() != reflect.Slice {
      return object, errors.New("Object at path " + object_path + " is not of kind slice.")
    }
    count := int(math.Min(float64(s.Len()), float64(max_count)))
    output_slice := make([]interface{}, 0, count)
    for i := 0; i < count; i++ {
      r := s.Index(i).Elem()
      if r.Kind() != reflect.Map {
        return object, errors.New("Row is not of kind map, but " + fmt.Sprint(r.Kind()))
      }
      m := map[string]interface{}{}
      for _, k := range r.MapKeys() {
        m[fmt.Sprint(k)] = gss.StringifyMapKeys(r.MapIndex(k).Interface())
      }
      if len(output_path) > 0 {
        x, err := dfl.Extract(output_path, m)
        if err != nil {
          return object, errors.Wrap(err, "error extracting object using path "+output_path+" from "+fmt.Sprint(m))
        }
        output_slice = append(output_slice, x)
      } else {
        output_slice = append(output_slice, m)
      }
    }
    return output_slice, nil
  }

  return object, nil
}
