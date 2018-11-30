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
	InfoDestination  string `viper:"info-destination"`
	InfoCompression  string `viper:"info-compression"`
	InfoFormat       string `viper:"info-format"`
	ErrorDestination string `viper:"error-destination"`
	ErrorCompression string `viper:"error-compression"`
	ErrorFormat      string `viper:"error-format"`
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
	m["InfoDestination"] = p.InfoDestination
	m["InfoCompression"] = p.InfoCompression
	m["InfoFormat"] = p.InfoFormat
	m["ErrorDestination"] = p.ErrorDestination
	m["ErrorCompression"] = p.ErrorCompression
	m["ErrorFormat"] = p.ErrorFormat
	return m
}
