package config

import (
	"strings"
)

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spatialcurrent/go-simple-serializer/gss"
)

type Process struct {
	AWS              *AWS
	Input            *Input
	Output           *Output
	Temp             *Temp
	Dfl              *Dfl
	ErrorDestination string `viper:"error-destination"`
	ErrorCompression string `viper:"error-compression"`
	LogDestination   string `viper:"log-destination"`
	LogCompression   string `viper:"log-compression"`
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
	return (p.Input != nil && p.Input.IsS3Bucket()) || (p.Temp != nil && p.Temp.IsS3Bucket()) || (p.Output != nil && p.Output.IsS3Bucket()) || strings.HasPrefix(p.ErrorDestination, "s3://") || strings.HasPrefix(p.LogDestination, "s3://")
}

func (p *Process) InputOptions() gss.Options {
	return p.Input.Options()
}

func (p *Process) OutputOptions() gss.Options {
	return p.Output.Options()
}

func (p *Process) Map() map[string]interface{} {
	m := map[string]interface{}{}
	if p.AWS != nil {
		m["AWS"] = p.AWS.Map()
	}
	if p.Input != nil {
		m["Input"] = p.Input.Map()
	}
	if p.Output != nil {
		m["Output"] = p.Output.Map()
	}
	if p.Temp != nil {
		m["Temp"] = p.Temp.Map()
	}
	if p.Dfl != nil {
		m["Dfl"] = p.Dfl.Map()
	}
	return m
}
