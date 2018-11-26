// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	"github.com/spatialcurrent/railgun/railgun/core"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"reflect"
	"sync"
)

type Catalog struct {
	mutex   *sync.Mutex
	objects map[string]interface{}
	indices map[string]map[string]int
}

func NewCatalog() *Catalog {

	catalog := &Catalog{
		mutex:   &sync.Mutex{},
		objects: map[string]interface{}{},
		indices: map[string]map[string]int{},
	}

	return catalog
}

func (c *Catalog) Lock() {
	c.mutex.Lock()
}

func (c *Catalog) Unlock() {
	c.mutex.Unlock()
}

func (c *Catalog) Get(name string, t reflect.Type) (core.Base, bool) {
	typeName := ""
	if t.Kind() == reflect.Ptr {
		typeName = t.Elem().Name()
	} else {
		typeName = t.Name()
	}
	if index, ok := c.indices[typeName]; ok {
		if position, ok := index[name]; ok {
			if objects, ok := c.objects[typeName]; ok {
				obj := reflect.ValueOf(objects).Index(position).Interface()
				if base, ok := obj.(core.Base); ok {
					return base, true
				}
			}
		}
	}
	return nil, false
}

func (c *Catalog) Add(obj interface{}) error {
	objectType := reflect.TypeOf(obj)
	typeName := objectType.Elem().Name()

	if n, ok := obj.(core.Named); ok {

		if _, ok := c.indices[typeName]; !ok {
			c.indices[typeName] = map[string]int{}
		} else {
			if _, ok := c.indices[typeName][n.GetName()]; ok {
				return &rerrors.ErrAlreadyExists{Name: typeName, Value: n.GetName()}
			}
		}

		if _, ok := c.objects[typeName]; !ok {
			c.objects[typeName] = reflect.MakeSlice(reflect.SliceOf(objectType), 0, 0).Interface()
		}

		c.objects[typeName] = reflect.Append(reflect.ValueOf(c.objects[typeName]), reflect.ValueOf(obj)).Interface()
		c.indices[typeName][n.GetName()] = reflect.ValueOf(c.objects[typeName]).Len() - 1
		return nil
	}

	return &rerrors.ErrMissingMethod{Type: typeName, Method: "GetName() string"}
}

func (c *Catalog) Update(obj interface{}) error {
	objectType := reflect.TypeOf(obj)
	typeName := objectType.Elem().Name()
	if n, ok := obj.(core.Named); ok {
		if index, ok := c.indices[typeName]; ok {
			if position, ok := index[n.GetName()]; ok {
				if objects, ok := c.objects[typeName]; ok {
					reflect.ValueOf(objects).Index(position).Set(reflect.ValueOf(obj))
					return nil
				}
			}
		}
		return &rerrors.ErrMissingObject{Type: typeName, Name: n.GetName()}
	}
	return &rerrors.ErrMissingObject{Type: typeName, Name: "unknown"}
}

func (c *Catalog) Delete(name string, t reflect.Type) error {
	typeName := t.Name()
	if index, ok := c.indices[typeName]; ok {
		if position, ok := index[name]; ok {
			if list, ok := c.objects[typeName]; ok {
				listValue := reflect.ValueOf(list)
				c.objects[typeName] = reflect.AppendSlice(listValue.Slice(0, position), listValue.Slice(position+1, listValue.Len())).Interface()
			}
			delete(index, name)
			return nil
		}
	}
	return &rerrors.ErrMissingObject{Type: typeName, Name: name}
}

func (c *Catalog) List(t reflect.Type) interface{} {
	if t.Kind() == reflect.Ptr {
		if list, ok := c.objects[t.Elem().Name()]; ok {
			return list
		}
	} else {
		if list, ok := c.objects[t.Name()]; ok {
			return list
		}
	}
	return reflect.MakeSlice(reflect.SliceOf(t), 0, 0).Interface()
}

func (c *Catalog) Dump() map[string]interface{} {
	dump := map[string]interface{}{}
	for typeName, input := range c.objects {
		output := make([]map[string]interface{}, 0)
		objects := reflect.ValueOf(input)
		numberOfObjects := objects.Len()
		for i := 0; i < numberOfObjects; i++ {
			v := objects.Index(i).Interface()
			if m, ok := v.(core.Mapper); ok {
				output = append(output, m.Map())
			}
		}
		dump[typeName] = output
	}
	return dump
}

func (c *Catalog) SafeDump() map[string]interface{} {
	c.Lock()
	dump := c.Dump()
	c.Unlock()
	return dump
}
