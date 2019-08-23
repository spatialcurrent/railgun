// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	"context"
	"reflect"
)

import (
	"github.com/spatialcurrent/go-sync-catalog/gsc"
)

import (
	"github.com/spatialcurrent/railgun/pkg/core"
	rerrors "github.com/spatialcurrent/railgun/pkg/errors"
)

type Catalog struct {
	*gsc.Catalog
}

func NewCatalog() *Catalog {
	return &Catalog{Catalog: gsc.NewCatalog()}
}

func (c *Catalog) Get(name string, t reflect.Type) (core.Base, bool) {
	if obj, exists := c.Catalog.Get(name, t); exists {
		if base, ok := obj.(core.Base); ok {
			return base, true
		}
	}
	return nil, false
}

func (c *Catalog) Add(obj interface{}) error {
	if n, ok := obj.(core.Named); ok {
		return c.Catalog.Add(n.GetName(), obj)
	}
	return &rerrors.ErrMissingMethod{Type: reflect.TypeOf(obj).Elem().Name(), Method: "GetName() string"}
}

func (c *Catalog) Update(obj interface{}) error {
	if n, ok := obj.(core.Named); ok {
		return c.Catalog.Update(n.GetName(), obj)
	}
	return &rerrors.ErrMissingObject{Type: reflect.TypeOf(obj).Elem().Name(), Name: "unknown"}
}

func (c *Catalog) Dump(ctx context.Context) map[string]interface{} {
	dump := map[string]interface{}{}
	for typeName, input := range c.Objects() {
		output := make([]map[string]interface{}, 0)
		objects := reflect.ValueOf(input)
		numberOfObjects := objects.Len()
		for i := 0; i < numberOfObjects; i++ {
			v := objects.Index(i).Interface()
			if m, ok := v.(core.Mapper); ok {
				output = append(output, m.Map(ctx))
			}
		}
		dump[typeName] = output
	}
	return dump
}

func (c *Catalog) SafeDump(ctx context.Context) map[string]interface{} {
	c.Lock() // lock the mutex for writing
	dump := c.Dump(ctx)
	c.Unlock() // Unlock the mutex for writing
	return dump
}
