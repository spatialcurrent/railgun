// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"context"
	"reflect"

	"github.com/spatialcurrent/go-dfl/pkg/dfl"

	"github.com/spatialcurrent/railgun/pkg/mapper"
)

type Function struct {
	Name        string   `map:"name,omitempty", rest:"name, the unique name of the fuction" required:"yes"`
	Title       string   `map:"title,omitempty" rest:"title, the title of the function"`
	Description string   `map:"description,omitempty" rest:"description, a verbose description of the process"`
	Aliases     []string `map:"aliases,omitempty" rest:"aliases, aliases for the function that are used by the executor" required:"yes"`
	Node        dfl.Node `map:"expression,omitempty" rest:"expression, the DFL expression of the function" required:"yes"`
	Tags        []string `map:"tags,omitempty" rest:"tags, tags for the function"`
}

func (f Function) GetName() string {
	return f.Name
}

func (f Function) Map(ctx context.Context) map[string]interface{} {
	return mapper.MarshalMapWithContext(ctx, f)
}

/*func (f Function) Map(ctx context.Context) map[string]interface{} {
	m := map[string]interface{}{
		"name":        f.Name,
		"title":       f.Title,
		"description": f.Description,
		"expression":  f.Node.Dfl(dfl.DefaultQuotes, true, 0),
	}
	if len(f.Aliases) > 0 {
		aliases := make([]dfl.Node, 0)
		for _, v := range f.Aliases {
			aliases = append(aliases, dfl.Literal{Value: v})
		}
		m["aliases"] = dfl.Array{Nodes: aliases}.Dfl(dfl.DefaultQuotes, false, 0)
	}
	if len(f.Tags) > 0 {
		tags := make([]dfl.Node, 0)
		for _, v := range f.Tags {
			tags = append(tags, dfl.Literal{Value: v})
		}
		m["tags"] = dfl.Array{Nodes: tags}.Dfl(dfl.DefaultQuotes, false, 0)
	}
	return m
}*/

func (f Function) Dfl(ctx context.Context) string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range f.Map(ctx) {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var FunctionType = reflect.TypeOf(Function{})
