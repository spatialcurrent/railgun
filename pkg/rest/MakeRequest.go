// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package rest

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-reader-writer/pkg/io"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/go-simple-serializer/pkg/json"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
)

type MakeRequestInput struct {
	Url           string
	Method        string
	Object        interface{}
	Authorization string
	OutputFormat  string
	OutputWriter  io.ByteWriter
	OutputPretty  bool
	Logger        *gsl.Logger
}

func MakeRequest(input *MakeRequestInput) error {

	var req *http.Request
	if input.Method != "GET" {
		inputBytes, err := json.Marshal(input.Object, false)
		if err != nil {
			return err
		}

		input.Logger.DebugF("Url: %q", input.Url)

		input.Logger.DebugF("Body: %v", string(inputBytes))

		r, err := http.NewRequest(input.Method, input.Url, bytes.NewBuffer(inputBytes))
		r.Header.Set("Content-Type", "application/json")
		if err != nil {
			return err
		}
		if len(input.Authorization) > 0 {
			r.Header.Set("Authorization", "bearer "+input.Authorization)
		}
		req = r
	} else {
		r, err := http.NewRequest(input.Method, input.Url, nil)
		if err != nil {
			return err
		}
		req = r
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if len(respBytes) == 0 {
		return errors.New("no response from server")
	}

	input.Logger.DebugF("Response: %v", string(respBytes))

	respObject, err := json.Unmarshal(respBytes)
	if err != nil {
		return err
	}

	outputBytes, err := gss.SerializeBytes(&gss.SerializeBytesInput{
		Object:            respObject,
		Format:            input.OutputFormat,
		Header:            gss.NoHeader,
		Limit:             gss.NoLimit,
		Pretty:            input.OutputPretty,
		LineSeparator:     "\n",
		KeyValueSeparator: "=",
	})
	if err != nil {
		return err
	}

	input.Logger.DebugF("Response Code: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(outputBytes))
	}

	_, err = input.OutputWriter.Write(outputBytes)
	if err != nil {
		return errors.Wrap(err, "error writing response to output writer")
	}

	err = input.OutputWriter.WriteByte('\n')
	if err != nil {
		return errors.Wrap(err, "error writing final newline to output writer")
	}

	err = io.Flush(input.OutputWriter)
	if err != nil {
		return errors.Wrap(err, "error flushing output writer")
	}

	return nil
}
