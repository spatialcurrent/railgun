package railgun

import (
	"reflect"
)

type Layer struct {
	Name        string    `cli:"name, the unique name of the workspace"`
	Title       string    `cli:"title, the title of the workspace"`
	Description string    `cli:"description, a verbose description of the workspace"`
	DataStore   DataStore `cli:"datastore, the name of the data store"`
	Extent      []float64
	Cache       *Cache
}

func (l Layer) Map() map[string]interface{} {
	return map[string]interface{}{
		"name":        l.Name,
		"title":       l.Title,
		"description": l.Description,
		"datastore":   l.DataStore.Map(),
		"extent":      l.Extent,
	}
}

var LayerType = reflect.TypeOf(Layer{})
