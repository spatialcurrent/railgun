// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"context"
	"github.com/spatialcurrent/go-dfl/dfl"
	"reflect"
)

type Process struct {
	Name        string   `rest:"name, the unique name of the process" required:"yes"`
	Title       string   `rest:"title, the title of the process"`
	Description string   `rest:"description, a verbose description of the process"`
	Node        dfl.Node `rest:"expression, the DFL expression of the process" required:"yes"`
	Tags        []string `rest:"tags, tags for the service"`
}

func (p Process) GetName() string {
	return p.Name
}

func (p Process) Map(ctx context.Context) map[string]interface{} {
	m := map[string]interface{}{
		"name":        p.Name,
		"title":       p.Title,
		"description": p.Description,
		"expression":  p.Node.Dfl(dfl.DefaultQuotes, true, 0),
		"variables":   p.Node.Variables(),
	}
	tags := make([]dfl.Node, 0)
	for _, v := range p.Tags {
		tags = append(tags, dfl.Literal{Value: v})
	}
	if len(tags) > 0 {
		m["tags"] = dfl.Array{Nodes: tags}.Dfl(dfl.DefaultQuotes, false, 0)
	}
	return m
}

func (p Process) Dfl(ctx context.Context) string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range p.Map(ctx) {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var ProcessType = reflect.TypeOf(Process{})
