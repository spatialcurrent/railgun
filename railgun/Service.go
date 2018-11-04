package railgun

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"reflect"
)

type Service struct {
	Name        string                 `rest:"name, the unique name of the service" required:"yes"`
	Title       string                 `rest:"title, the title of the service"`
	Description string                 `rest:"description, a verbose description of the service"`
	DataStore   DataStore              `rest:"datastore, the name of the data store" required:"yes"`
	Process     Process                `rest:"process, the name of the process" required:"yes"`
	Defaults    map[string]interface{} `rest:"defaults, the default values of the variables for this service"`
}

func (s Service) Map() map[string]interface{} {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range s.Defaults {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	m := map[string]interface{}{
		"name":        s.Name,
		"title":       s.Title,
		"description": s.Description,
		"datastore":   s.DataStore.Name,
		"process":     s.Process.Name,
	}
	if len(dict) > 0 {
		m["defaults"] = dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
	}
	return m
}

func (s Service) Dfl() string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range s.Map() {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var ServiceType = reflect.TypeOf(Service{})
