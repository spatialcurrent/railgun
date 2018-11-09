// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"fmt"
	"github.com/spatialcurrent/go-dfl/dfl"
	"reflect"
)

type Process struct {
	Name        string   `rest:"name, the unique name of the process" required:"yes"`
	Title       string   `rest:"title, the title of the process"`
	Description string   `rest:"description, a verbose description of the process"`
	Node        dfl.Node `rest:"expression, the DFL expression of the process" required:"yes"`
}

func (p Process) GetName() string {
	return p.Name
}

func (p Process) Map() map[string]interface{} {
	fmt.Println("p.Node:", p.Node)
	return map[string]interface{}{
		"name":        p.Name,
		"title":       p.Title,
		"description": p.Description,
		"expression":  p.Node.Dfl(dfl.DefaultQuotes, false, 0),
		"variables":   p.Node.Variables(),
	}
}

func (p Process) Dfl() string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range p.Map() {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var ProcessType = reflect.TypeOf(Process{})
