package config

import (
//"strings"
)

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spatialcurrent/go-simple-serializer/gss"
)

type Athena struct {
	AWS              *AWS
	QueryExecutionId string `viper:"query-execution-id"`
	Output           *Output
	Temp             *Temp
	Dfl              *Dfl
	ErrorDestination string `viper:"error-destination"`
	ErrorCompression string `viper:"error-compression"`
	LogDestination   string `viper:"log-destination"`
	LogCompression   string `viper:"log-compression"`
}

func (a *Athena) AWSSessionOptions() session.Options {
	return a.AWS.SessionOptions()
}

func (a *Athena) OutputOptions() gss.Options {
	return a.Output.Options()
}

func (a *Athena) Map() map[string]interface{} {
	m := map[string]interface{}{}
	if a.AWS != nil {
		m["AWS"] = a.AWS.Map()
	}
	m["QueryExecutionId"] = a.QueryExecutionId
	if a.Output != nil {
		m["Output"] = a.Output.Map()
	}
	if a.Temp != nil {
		m["Temp"] = a.Temp.Map()
	}
	if a.Dfl != nil {
		m["Dfl"] = a.Dfl.Map()
	}
	m["ErrorDestination"] = a.ErrorDestination
	m["ErrorCompression"] = a.ErrorCompression
	m["LogDestination"] = a.LogDestination
	m["LogCompression"] = a.LogCompression
	return m
}
