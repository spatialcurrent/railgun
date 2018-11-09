// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"github.com/spatialcurrent/railgun/railgun/cache"
	"github.com/spatialcurrent/go-dfl/dfl"
	"reflect"
)

type Layer struct {
	Name        string     `rest:"name, the unique name of the workspace" required:"yes"`
	Title       string     `rest:"title, the title of the workspace"`
	Description string     `rest:"description, a verbose description of the workspace"`
	DataStore   *DataStore `rest:"datastore, the name of the data store" required:"yes"`
	Extent      []float64  `rest:"extent, the extent of the data"`
	Cache       *cache.Cache
}

func (l Layer) GetName() string {
	return l.Name
}

func (l Layer) Map() map[string]interface{} {
	return map[string]interface{}{
		"name":        l.Name,
		"title":       l.Title,
		"description": l.Description,
		"datastore":   l.DataStore.Name,
		"extent":      l.Extent,
	}
}

func (l Layer) Dfl() string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range l.Map() {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var LayerType = reflect.TypeOf(Layer{})
