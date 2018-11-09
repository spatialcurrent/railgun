// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"reflect"
)

type Job struct {
	Service     *Service               `rest:"service, the name of the service" required:"yes"`
	Name        string                 `rest:"name, the name of the job, not required"`
	Title       string                 `rest:"title, the title of the job, not required"`
	Description string                 `rest:"description, a verbose description of the job, not required"`
	Variables   map[string]interface{} `rest:"variables, the input variables for the job"`
	Output      *DataStore             `rest:"output, the output for the job"`
}

func (j Job) GetName() string {
	return j.Name
}

func (j Job) Map() map[string]interface{} {
	m := map[string]interface{}{
		"name":        j.Name,
		"title":       j.Title,
		"description": j.Description,
		"service":     j.Service.Name,
	}
	variables := map[dfl.Node]dfl.Node{}
	for k, v := range j.Variables {
		variables[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	if len(variables) > 0 {
		m["variables"] = dfl.Dictionary{Nodes: variables}.Dfl(dfl.DefaultQuotes, false, 0)
	}
	if j.Output != nil {
		m["output"] = j.Output.Name
	}
	return m
}

func (j Job) Dfl() string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range j.Map() {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var JobType = reflect.TypeOf(Job{})
