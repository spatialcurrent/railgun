// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"

	rerrors "github.com/spatialcurrent/railgun/pkg/errors"
	"github.com/spatialcurrent/railgun/pkg/serializer"
)

type LoadFromUriInput struct {
	Uri         string
	Format      string
	Compression string
	Logger      *gsl.Logger
	S3Client    *s3.S3
}

func (c *RailgunCatalog) LoadFromUri(input *LoadFromUriInput) error {

	uri := input.Uri
	logger := input.Logger

	logger.Info(fmt.Sprintf("* loading catalog from %s", uri))

	raw, err := serializer.New(input.Format, input.Compression, grw.NoDict).S3Client(input.S3Client).Deserialize(input.Uri)
	if err != nil {
		return errors.Wrapf(err, "error loading catalog from uri %q", input.Uri)
	}

	if raw == nil {
		logger.Info("* catalog was empty")
		return nil
	}

	if t := reflect.TypeOf(raw); t.Kind() == reflect.Map {
		v := reflect.ValueOf(raw)

		//logger.Info(fmt.Sprintf("* catalog has keys %v", v.MapKeys()))

		key := "Workspace"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseWorkspace(m)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading workspace"))
						logger.Error(err)
						continue
					}
					err = c.Add(obj)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading workspace"))
						logger.Error(err)
						continue
					}
					logger.InfoF("Loaded workspace %q", obj.Name)
				}
			}
		}

		key = "DataStore"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseDataStore(m)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading data store"))
						logger.Error(err)
						continue
					}
					err = c.Add(obj)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading data store"))
						logger.Error(err)
						continue
					}
					logger.InfoF("Loaded data store %q", obj.Name)
				}
			}
		}

		key = "Layer"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseLayer(m)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading layer"))
						logger.Error(err)
						continue
					}
					err = c.Add(obj)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading layer"))
						logger.Error(err)
						continue
					}
					logger.InfoF("Loaded layer %q", obj.Name)
				}
			}
		}

		key = "Function"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseFunction(m)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading function"))
						logger.Error(err)
						continue
					}
					err = c.Add(obj)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading function"))
						logger.Error(err)
						continue
					}
					logger.InfoF("Loaded function %q", obj.Name)
				}
			}
		}

		key = "Process"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseProcess(m)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading process"))
						logger.Error(err)
						continue
					}
					err = c.Add(obj)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading process"))
						logger.Error(err)
						continue
					}
					logger.InfoF("Loaded process %q", obj.Name)
				}
			}
		}

		key = "Service"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseService(m)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading service"))
						logger.Error(err)
						continue
					}
					err = c.Add(obj)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading service"))
						logger.Error(err)
						continue
					}
					logger.InfoF("Loaded service %q", obj.Name)
				}
			}
		}

		key = "Job"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseJob(m)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading job"))
						logger.Error(err)
						continue
					}
					err = c.Add(obj)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading job"))
						logger.Error(err)
						continue
					}
					logger.InfoF("Loaded job %q", obj.Name)
				}
			}
		}

		key = "Workflow"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseWorkflow(m)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading workflow"))
						logger.Error(err)
						continue
					}
					err = c.Add(obj)
					if err != nil {
						logger.Error(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading workflow"))
						logger.Error(err)
						continue
					}
					logger.InfoF("Loaded workflow %q", obj.Name)
				}
			}
		}

	}

	return nil
}
