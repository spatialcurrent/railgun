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

type Workflow struct {
	Name        string                 `rest:"name, the name of the workflow" required:"yes"`
	Title       string                 `rest:"title, the title of the workflow, not required"`
	Description string                 `rest:"description, a verbose description of the workflow, not required"`
	Variables   map[string]interface{} `rest:"variables, global variables for the workflow"`
	Jobs        []*Job                 `rest:"jobs, the jobs for the workflow, in order of execution" required:"yes"`
}

func (w Workflow) GetName() string {
	return w.Name
}

func (w Workflow) Map() map[string]interface{} {
	jobs := make([]string, 0, len(w.Jobs))
	for _, j := range w.Jobs {
		jobs = append(jobs, j.Name)
	}
	m := map[string]interface{}{
		"name":        w.Name,
		"title":       w.Title,
		"description": w.Description,
		"jobs":        jobs,
	}
	variables := map[dfl.Node]dfl.Node{}
	for k, v := range w.Variables {
		variables[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	if len(variables) > 0 {
		m["variables"] = dfl.Dictionary{Nodes: variables}.Dfl(dfl.DefaultQuotes, false, 0)
	}
	return m
}

func (w Workflow) Dfl() string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range w.Map() {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var WorkflowType = reflect.TypeOf(Workflow{})
