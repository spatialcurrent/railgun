// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

// railgun.js is the Javascript package for Railgun.
//
package main

import (
	"fmt"
	"reflect"
	"strings"
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-dfl/dfljs"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-simple-serializer/gssjs"
	"github.com/spatialcurrent/railgun/railgun"
)

import (
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/console"
)

func main() {

	js.Global.Set("railgun", map[string]interface{}{
		"version": railgun.Version,
		"process": Process,
		"dfl": map[string]interface{}{
			"version":        dfl.Version,
			"Parse":          dfljs.Parse,
			"EvaluateBool":   dfljs.EvaluateBool,
			"EvaluateInt":    dfljs.EvaluateInt,
			"EvaluateFloat":  dfljs.EvaluateFloat64,
			"EvaluateString": dfljs.EvaluateString,
		},
		"gss": map[string]interface{}{
			"version":     gss.Version,
			"formats":     gss.Formats,
			"convert":     gssjs.Convert,
			"deserialize": gssjs.Deserialize,
			"serialize":   gssjs.Serialize,
		},
	})
}

func Process(in interface{}, options *js.Object) interface{} {

	// Convert Javascript options object into Golang map
	m := map[string]interface{}{}
	for _, key := range js.Keys(options) {
		m[key] = options.Get(key).Interface()
	}

	input_header := []string{}
	input_comment := ""
	input_format := ""
	input_lazy_quotes := false
	input_limit := gss.NoLimit
	
	output_format := ""
	output_header := []string{}
	output_limit := gss.NoLimit
	dfl_exp := ""

	if v, ok := m["input_header"]; ok {
		switch v.(type) {
		case []string:
			input_header = v.([]string)
		case []interface{}:
			input_header = make([]string, 0, len(v.([]interface{})))
			for _, h := range v.([]interface{}) {
				input_header = append(input_header, fmt.Sprint(h))
			}
		}
	}

	if v, ok := m["input_comment"]; ok {
		switch v.(type) {
		case string:
			input_comment = v.(string)
		}
	}

	if v, ok := m["input_format"]; ok {
		switch v.(type) {
		case string:
			input_format = v.(string)
		}
	}
	
	if v, ok := m["input_lazy_quotes"]; ok {
		switch v.(type) {
		case bool:
			input_lazy_quotes = v.(bool)
		}
	}
	
	if v, ok := m["input_limit"]; ok {
		switch v.(type) {
		case int:
			input_limit = v.(int)
		}
	}


	if v, ok := m["output_format"]; ok {
		switch v.(type) {
		case string:
			output_format = v.(string)
		}
	}
	
	if v, ok := m["output_header"]; ok {
		switch v.(type) {
		case []string:
			output_header = v.([]string)
		case []interface{}:
			output_header = make([]string, 0, len(v.([]interface{})))
			for _, h := range v.([]interface{}) {
				output_header = append(output_header, fmt.Sprint(h))
			}
		}
	}
	
	if v, ok := m["output_limit"]; ok {
		switch v.(type) {
		case int:
			output_limit = v.(int)
		}
	}

	var ctx interface{}

	switch in.(type) {
	case string:
		input_type, err := gss.GetType([]byte(in.(string)), input_format)
		if err != nil {
			console.Error(errors.Wrap(err, "error geting type for input").Error())
			return ""
		}
		input_object, err := gss.DeserializeString(in.(string), input_format, input_header, input_comment, input_lazy_quotes, input_limit, input_type, false)
		if err != nil {
			console.Error(errors.Wrap(err, "error deserializing input using format "+input_format).Error())
			return ""
		}
		ctx = input_object
	case *js.Object:
		ctx = in.(*js.Object).Interface()
	case map[string]interface{}:
		ctx = in.(map[string]interface{})
	case []interface{}:
		ctx = in.([]interface{})
	case []map[string]interface{}:
		ctx = in.([]map[string]interface{})
	default:
		console.Error("Unknown input type", fmt.Sprint(reflect.TypeOf(in)))
		return ""
	}

	if v, ok := m["dfl"]; ok {
		switch v.(type) {
		case string:
			dfl_exp = v.(string)
		case []interface{}:
			arr := dfl.TryConvertArray(v.([]interface{}))
			switch arr.(type) {
			case []string:
				dfl_exp = strings.Join(arr.([]string), " | ")
			}
		}
	}

	var dfl_node dfl.Node
	if len(dfl_exp) > 0 {
		n, err := dfl.ParseCompile(dfl_exp)
		if err != nil {
			console.Error(errors.Wrap(err, "Error parsing DFL node.").Error())
			return ""
		}
		dfl_node = n
	}

	var output interface{}
	if dfl_node != nil {
		_, o, err := dfl_node.Evaluate(map[string]interface{}{}, ctx, dfl.NewFuntionMapWithDefaults(), []string{"\"", "'", "`"})
		if err != nil {
			console.Error(errors.Wrap(err, "error processing").Error())
			return ""
		}
		output = o
	} else {
		output = ctx
	}

	output = gss.StringifyMapKeys(output)

	if len(output_format) > 0 {
		output_string, err := gss.SerializeString(output, output_format, output_header, output_limit)
		if err != nil {
			console.Error(errors.Wrap(err, "Error converting to output format "+output_format).Error())
			return ""
		}
		return output_string
	}

	return output
}
