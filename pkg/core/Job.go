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

type Job struct {
	Service     *Service               `map:"service" rest:"service, the name of the service" required:"yes"`
	Name        string                 `map:"name" rest:"name, the name of the job, not required"`
	Title       string                 `map:"title" rest:"title, the title of the job, not required"`
	Description string                 `map:"description" rest:"description, a verbose description of the job, not required"`
	Variables   map[string]interface{} `map:"variables,omitempty" rest:"variables, the input variables for the job"`
	Output      *DataStore             `map:"output,omitempty" rest:"output, the output for the job"`
}

func (j Job) GetName() string {
	return j.Name
}

func (j Job) Map(ctx context.Context) map[string]interface{} {
	return mapper.MarshalMapWithContext(ctx, j)
}

func (j Job) Dfl(ctx context.Context) string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range j.Map(ctx) {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var JobType = reflect.TypeOf(Job{})
