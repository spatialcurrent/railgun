// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"context"
	"github.com/spatialcurrent/go-dfl/dfl"
	"reflect"
)

type Service struct {
	Name        string                 `rest:"name, the unique name of the service" required:"yes"`
	Title       string                 `rest:"title, the title of the service"`
	Description string                 `rest:"description, a verbose description of the service"`
	DataStore   *DataStore             `rest:"datastore, the name of the data store" required:"yes"`
	Process     *Process               `rest:"process, the name of the process" required:"yes"`
	Defaults    map[string]interface{} `rest:"defaults, the default values of the variables for this service"`
	Tags        []string               `rest:"tags, tags for the service"`
	Transform   dfl.Node               `rest:"transform, transform applied to request variables before passing to datastore and process" required:"no" visibility:"private"`
}

func (s Service) GetName() string {
	return s.Name
}

func (s Service) Map(ctx context.Context) map[string]interface{} {
	m := map[string]interface{}{
		"name":        s.Name,
		"title":       s.Title,
		"description": s.Description,
		"datastore":   s.DataStore.Name,
		"process":     s.Process.Name,
	}
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range s.Defaults {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	if len(dict) > 0 {
		m["defaults"] = dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
	}
	tags := make([]dfl.Node, 0)
	for _, v := range s.Tags {
		tags = append(tags, dfl.Literal{Value: v})
	}
	if len(tags) > 0 {
		m["tags"] = dfl.Array{Nodes: tags}.Dfl(dfl.DefaultQuotes, false, 0)
	}
	if s.Transform != nil {
		m["transform"] = s.Transform.Dfl(dfl.DefaultQuotes, false, 0)
	}
	return m
}

func (s Service) Dfl(ctx context.Context) string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range s.Map(ctx) {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var ServiceType = reflect.TypeOf(Service{})
