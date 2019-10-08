// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package config

import (
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
)

type Process struct {
	AWS                 *AWS          `map:"AWS"`
	Input               *Input        `map:"Input"`
	Output              *Output       `map:"Output"`
	Temp                *Temp         `map:"Temp"`
	Dfl                 *Dfl          `map:"Dfl"`
	Stream              bool          `viper:"stream" map:"Stream"`
	InfoDestination     string        `viper:"info-destination" map:"InfoDestination"`
	InfoCompression     string        `viper:"info-compression" map:"InfoCompression"`
	InfoFormat          string        `viper:"info-format" map:"InfoFormat"`
	ErrorDestination    string        `viper:"error-destination" map:"ErrorDestination"`
	ErrorCompression    string        `viper:"error-compression" map:"ErrorCompression"`
	ErrorFormat         string        `viper:"error-format" map:"ErrorFormat"`
	FileDescriptorLimit int           `viper:"file-descriptor-limit" map:"FileDescriptorLimit"`
	Time                bool          `viper:"time" map:"Time"`
	Timeout             time.Duration `viper:"timeout" map:"Timeout"`
	Verbose             bool          `viper:"verbose" map:"Verbose"`
}

func NewProcessConfig() *Process {
	return &Process{
		AWS:                 &AWS{},
		Input:               &Input{},
		Output:              &Output{},
		Temp:                &Temp{},
		Dfl:                 &Dfl{},
		Stream:              false,
		InfoDestination:     "",
		InfoCompression:     "",
		InfoFormat:          "",
		ErrorDestination:    "",
		ErrorCompression:    "",
		ErrorFormat:         "",
		FileDescriptorLimit: -1,
		Time:                false,
		Timeout:             0 * time.Second,
		Verbose:             false,
	}
}

func (p *Process) AWSSessionOptions() session.Options {
	return p.AWS.SessionOptions()
}

func (p *Process) HasAWSResource() bool {
	return p.HasAthenaStoredQuery() || p.HasS3Bucket()
}

func (p *Process) HasAthenaStoredQuery() bool {
	return (p.Input != nil && p.Input.IsAthenaStoredQuery())
}

func (p *Process) HasS3Bucket() bool {
	return (p.Input != nil && p.Input.IsS3Bucket()) || (p.Temp != nil && p.Temp.IsS3Bucket()) || (p.Output != nil && p.Output.IsS3Bucket()) || strings.HasPrefix(p.InfoDestination, "s3://") || strings.HasPrefix(p.ErrorDestination, "s3://")
}

func (p *Process) Map() map[string]interface{} {
	m := map[string]interface{}{}
	v := reflect.ValueOf(p)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	for i := 0; i < v.NumField(); i++ {
		if tag := t.Field(i).Tag.Get("map"); len(tag) > 0 && tag != "-" {
			fieldValue := v.Field(i).Interface()
			if fieldMap, ok := fieldValue.(mapper); ok {
				m[tag] = fieldMap.Map()
			} else {
				m[tag] = fieldValue
			}
		}
	}
	return m
}
