package middleware

import (
	"time"
)

type Request struct {
  Client string
	Host    string
	Url     string
	Method  string
	Start  *time.Time
	End    *time.Time
	Subject string
	Vars map[string]string
	Handler string
	Error error
}

func NewRequest(client string, host string, url string, method string, start *time.Time, subject string, handler string, err error) Request {
	return Request{
	  Client: client,
		Host:    host,
		Url:     url,
		Method:  method,
		Start:    start,
		End: nil,
		Subject: subject,
		Vars: map[string]string{},
		Handler: handler,
		Error: err,
	}
}

func (r Request) Map() map[string]interface{} {
	m := map[string]interface{}{
	  "client": r.Client,
		"host":   r.Host,
		"url":    r.Url,
		"method": r.Method,
	}
	if r.Start != nil {
	  m["start"] = r.Start.Format(time.RFC3339)
	}
	if r.End != nil {
	  m["end"] = r.End.Format(time.RFC3339)
	}
	if r.Start != nil && r.End != nil {
	  m["duration"] = r.End.Sub(*r.Start).String()
	}
	if len(r.Subject) > 0 {
		m["subject"] = r.Subject
	}
	if len(r.Vars) > 0 {
		m["vars"] = r.Vars
	}
	if len(r.Handler) > 0 {
		m["handler"] = r.Handler
	}
	if r.Error != nil {
		m["error"] = r.Error.Error()
	}
	return m
}
