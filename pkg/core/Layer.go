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
	"github.com/spatialcurrent/railgun/pkg/cache"
)

type Layer struct {
	Name        string                 `rest:"name, the unique name of the workspace" required:"yes"`
	Title       string                 `rest:"title, the title of the workspace"`
	Description string                 `rest:"description, a verbose description of the workspace"`
	DataStore   *DataStore             `rest:"datastore, the name of the data store" required:"yes"`
	Node        dfl.Node               `rest:"expression, the DFL expression of the layer" required:"yes"`
	Defaults    map[string]interface{} `rest:"defaults, the default values of the variables for this service"`
	Extent      []float64              `rest:"extent, the extent of the data"`
	Tags        []string               `rest:"tags, tags for the service"`
	Cache       *cache.Cache
}

func (l Layer) GetName() string {
	return l.Name
}

func (l Layer) Map(ctx context.Context) map[string]interface{} {
	m := map[string]interface{}{
		"name":        l.Name,
		"title":       l.Title,
		"description": l.Description,
		"datastore":   l.DataStore.Name,
		"extent":      l.Extent,
	}
	if l.Node != nil {
		m["expression"] = l.Node.Dfl(dfl.DefaultQuotes, false, 0)
	}
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range l.Defaults {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	if len(dict) > 0 {
		m["defaults"] = dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
	}
	tags := make([]dfl.Node, 0)
	for _, v := range l.Tags {
		tags = append(tags, dfl.Literal{Value: v})
	}
	if len(tags) > 0 {
		m["tags"] = dfl.Array{Nodes: tags}.Dfl(dfl.DefaultQuotes, false, 0)
	}
	return m
}

func (l Layer) Dfl(ctx context.Context) string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range l.Map(ctx) {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var LayerType = reflect.TypeOf(Layer{})
