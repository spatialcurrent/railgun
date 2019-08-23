// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package config

import (
	"github.com/aws/aws-sdk-go/aws/session"
)

type AthenaDownload struct {
	AWS              *AWS
	QueryExecutionId string `viper:"query-execution-id"`
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

func (a *AthenaDownload) AWSSessionOptions() session.Options {
	return a.AWS.SessionOptions()
}

func (a *AthenaDownload) Map() map[string]interface{} {
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
	m["InfoDestination"] = a.InfoDestination
	m["InfoCompression"] = a.InfoCompression
	m["InfoFormat"] = a.InfoFormat
	m["ErrorDestination"] = a.ErrorDestination
	m["ErrorCompression"] = a.ErrorCompression
	m["ErrorFormat"] = a.ErrorFormat
	return m
}

func NewAthenaDownload() *AthenaDownload {
	return &AthenaDownload{
		AWS:              &AWS{},
		QueryExecutionId: "",
		Output:           &Output{},
		Temp:             &Temp{},
		Dfl:              &Dfl{},
		InfoDestination:  "",
		InfoCompression:  "",
		InfoFormat:       "",
		ErrorDestination: "",
		ErrorCompression: "",
		ErrorFormat:      "",
	}
}
