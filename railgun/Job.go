// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package railgun

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"reflect"
)

type Job struct {
	Service     Service                `rest:"service, the name of the service" required:"yes"`
	Name        string                 `rest:"name, the name of the job, not required"`
	Title       string                 `rest:"title, the title of the job, not required"`
	Description string                 `rest:"description, a verbose description of the job, not required"`
	Variables   map[string]interface{} `rest:"variables, the input variables for the job"`
}

func (j Job) Map() map[string]interface{} {
	return map[string]interface{}{
		"name":        j.Name,
		"title":       j.Title,
		"description": j.Description,
		"service":     j.Service.Name,
		"variables":   j.Variables,
	}
}

func (j Job) Dfl() string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range j.Map() {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var JobType = reflect.TypeOf(Job{})
