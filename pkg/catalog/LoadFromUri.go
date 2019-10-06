// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"github.com/spatialcurrent/viper"
)

import (
	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-reader-writer/pkg/splitter"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
	"github.com/spatialcurrent/go-try-get/pkg/gtg"
)

import (
	"github.com/spatialcurrent/railgun/pkg/cache"
	"github.com/spatialcurrent/railgun/pkg/core"
	rerrors "github.com/spatialcurrent/railgun/pkg/errors"
	"github.com/spatialcurrent/railgun/pkg/parser"
	"github.com/spatialcurrent/railgun/pkg/util"
)

type LoadFromUriInput struct {
  Uri string
  Format string
  Compression string
  Logger *gsl.Logger
  S3Client *s3.S3
}

func (c *RailgunCatalog) LoadFromUri(input *LoadFromUriInput) error {
  
  logger := input.Logger

	logger.Info(fmt.Sprintf("* loading catalog from %s", uri))

	raw, err := func() (interface{}, error) {

		inputBytes, err := grw.ReadAllAndClose(&grw.ReadAllAndCloseInput{
			Uri:        input.Uri,
			Alg:        input.Compression,
			Dict:       grw.NoDict,
			BufferSize: grw.DefaultBufferSize,
			S3Client:   input.s3Client,
		})
		if err != nil {
			return nil, err
		}

		if len(inputBytes) == 0 {
			return nil, nil
		}

		inputObject, err := gss.DeserializeBytes(&gss.DeserializeBytesInput{
			Bytes:         inputBytes,
			Format:        input.Format,
			Header:        gss.NoHeader,
			Comment:       gss.NoComment,
			LazyQuotes:    false,
			SkipLines:     gss.NoSkip,
			Limit:         gss.NoLimit,
			LineSeparator: "\n",
			DropCR:        true,
		})
		if err != nil {
			return nil, err
		}

		return inputObject, nil

	}()

	if err != nil {
		return errors.Wrap(err, "error loading catalog")
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
					logger.Info(map[string]interface{}{
						"init": map[string]interface{}{
							"workspace": map[string]interface{}{"name": obj.Name},
						}
					})
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
					logger.Info(map[string]interface{}{
						"init": map[string]interface{}{
							"datastore": map[string]interface{}{"name": obj.Name},
						},
					})
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
					logger.Info(map[string]interface{}{
						"init": map[string]interface{}{
							"layer": map[string]interface{}{"name": obj.Name},
						},
					})
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
					logger.Info(map[string]interface{}{
						"init": map[string]interface{}{
							"function": map[string]interface{}{"name": obj.Name},
						},
					})
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
					logger.Info(map[string]interface{}{
						"init": map[string]interface{}{
							"process": map[string]interface{}{"name": obj.Name},
						},
					})
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
					logger.Info(map[string]interface{}{
						"init": map[string]interface{}{
							"service": map[string]interface{}{"name": obj.Name},
						},
					})
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
					logger.Info(map[string]interface{}{
						"init": map[string]interface{}{
							"job": map[string]interface{}{"name": obj.Name},
						},
					})
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
					logger.Info(map[string]interface{}{
						"init": map[string]interface{}{
							"workflow": map[string]interface{}{"name": obj.Name},
						},
					})
				}
			}
		}

	}

	return nil
}