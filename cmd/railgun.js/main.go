// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

// railgun.js is the Javascript version of Railgun.
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
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
)

import (
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/console"
)

var GO_RAILGUN_COMPRESSION_ALGORITHMS = []string{"none", "gzip", "snappy"}
var GO_RAILGUN_FORMATS = []string{"csv", "tsv", "hcl", "hcl2", "json", "jsonl", "properties", "toml", "yaml"}

func main() {
	js.Global.Set("railgun", map[string]interface{}{
		"version": railgun.VERSION,
		"process": Process,
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
	output_format := ""
	dfl_exp := ""

	if v, ok := m["header"]; ok {
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

	if v, ok := m["output_format"]; ok {
		switch v.(type) {
		case string:
			output_format = v.(string)
		}
	}

	var ctx interface{}

	switch in.(type) {
	case string:
		input_object, err := gss.NewObject(in.(string), input_format)
		if err != nil {
			console.Log(errors.Wrap(err, "error creating new object for format "+input_format))
			return ""
		}
		switch input_object_typed := input_object.(type) {
		case []map[string]interface{}:
			err = gss.Deserialize(in.(string), input_format, input_header, input_comment, &input_object_typed)
			if err != nil {
				console.Log(errors.Wrap(err, "error deserializing input using format "+input_format).Error())
				return ""
			}
			ctx = input_object_typed // This is a critical line, otherwise the type information is lost.
		default:
			err = gss.Deserialize(in.(string), input_format, input_header, input_comment, &input_object)
			if err != nil {
				console.Log(errors.Wrap(err, "error deserializing input using format "+input_format).Error())
				return ""
			}
			ctx = input_object
		}
	case *js.Object:
		ctx = in.(*js.Object).Interface()
	case map[string]interface{}:
		ctx = in.(map[string]interface{})
	default:
		console.Log("Unknown input type", fmt.Sprint(reflect.TypeOf(in)))
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
		n, err := dfl.Parse(dfl_exp)
		if err != nil {
			console.Log(errors.Wrap(err, "Error parsing dfl node.").Error())
			return ""
		}
		dfl_node = n.Compile()
	}

	var output interface{}
	if dfl_node != nil {
		o, err := dfl_node.Evaluate(ctx, dfl.NewFuntionMapWithDefaults())
		if err != nil {
			console.Log(errors.Wrap(err, "error processing").Error())
			return ""
		}
		output = o
	} else {
		output = ctx
	}

	output = gss.StringifyMapKeys(output)

	if len(output_format) > 0 {
		output_string, err := gss.Serialize(output, output_format)
		if err != nil {
			console.Log(errors.Wrap(err, "Error converting to output format "+output_format).Error())
			return ""
		}
		return output_string
	}

	return output
}
