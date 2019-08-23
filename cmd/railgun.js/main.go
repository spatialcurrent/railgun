// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
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
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/console"
)

import (
	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-dfl/pkg/dfljs"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gssjs"
	"github.com/spatialcurrent/go-stringify/pkg/stringify"
)

func main() {

	js.Global.Set("railgun", map[string]interface{}{
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

	input_header := gss.NoHeader
	inputComment := gss.NoComment
	inputFormat := ""
	inputLazyQuotes := false
	inputSkipLines := gss.NoSkip
	inputLimit := gss.NoLimit

	outputFormat := ""
	outputPretty := false
	outputHeader := gss.NoHeader
	outputLimit := gss.NoLimit

	async := false

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

	if v, ok := m["inputComment"]; ok {
		switch v.(type) {
		case string:
			inputComment = v.(string)
		}
	}

	if v, ok := m["inputFormat"]; ok {
		switch v.(type) {
		case string:
			inputFormat = v.(string)
		}
	}

	if v, ok := m["inputLazyQuotes"]; ok {
		switch v.(type) {
		case bool:
			inputLazyQuotes = v.(bool)
		}
	}

	if v, ok := m["inputSkipLines"]; ok {
		switch v.(type) {
		case int:
			inputSkipLines = v.(int)
		}
	}

	if v, ok := m["inputLimit"]; ok {
		switch v.(type) {
		case int:
			inputLimit = v.(int)
		}
	}

	if v, ok := m["outputFormat"]; ok {
		switch v.(type) {
		case string:
			outputFormat = v.(string)
		}
	}

	if v, ok := m["outputPretty"]; ok {
		switch v.(type) {
		case bool:
			outputPretty = v.(bool)
		}
	}

	if v, ok := m["outputHeader"]; ok {
		switch v.(type) {
		case []string:
			outputHeader = v.([]string)
		case []interface{}:
			outputHeader = make([]string, 0, len(v.([]interface{})))
			for _, h := range v.([]interface{}) {
				outputHeader = append(outputHeader, fmt.Sprint(h))
			}
		}
	}

	if v, ok := m["outputLimit"]; ok {
		switch v.(type) {
		case int:
			outputLimit = v.(int)
		}
	}

	if v, ok := m["async"]; ok {
		switch v.(type) {
		case bool:
			async = v.(bool)
		}
	}

	var ctx interface{}

	switch in.(type) {
	case string:
		input_type, err := gss.GetType([]byte(in.(string)), inputFormat)
		if err != nil {
			console.Error(errors.Wrap(err, "error geting type for input").Error())
			return ""
		}
		input_object, err := gss.DeserializeString(in.(string), inputFormat, input_header, inputComment, inputLazyQuotes, inputSkipLines, inputLimit, input_type, async, false)
		if err != nil {
			console.Error(errors.Wrap(err, "error deserializing input using format "+inputFormat).Error())
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

	output = stringify.StringifyMapKeys(output)

	if len(outputFormat) > 0 {
		output_string, err := gss.SerializeString(&gss.SerializeInput{
			Object: output,
			Format: outputFormat,
			Header: outputHeader,
			Limit:  outputLimit,
			Pretty: outputPretty,
		})
		if err != nil {
			console.Error(errors.Wrap(err, "Error converting to output format "+outputFormat).Error())
			return ""
		}
		return output_string
	}

	return output
}
