// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/named"
	"github.com/spatialcurrent/railgun/railgun/request"
	"github.com/spatialcurrent/railgun/railgun/util"
	"net/http"
	"strings"
)

type ItemsHandler struct {
	*BaseHandler
	AwsSessionCache *gocache.Cache
	DflFuncs        dfl.FunctionMap
}

func (h *ItemsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	qs := request.NewQueryString(r)
	err := h.Run(w, r, vars, qs)
	if err != nil {
		h.Errors <- err
		respondWithEmptyFeatureCollection(w)
		//w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *ItemsHandler) Run(w http.ResponseWriter, r *http.Request, vars map[string]string, qs request.QueryString) error {

	layer, ok := h.Catalog.GetLayer(vars["name"])
	if !ok {
		return &rerrors.ErrMissingObject{Type: "layer", Name: vars["name"]}
	}

	ext := vars["ext"]

	_, outputFormat, _ := util.SplitNameFormatCompression(r.URL.Path)

	pipeline := []dfl.Node{}

	exp, err := qs.FirstString("dfl")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *request.ErrQueryStringParameterMissing:
		default:
			return err
		}
	}

	if len(exp) > 0 {
		_, _, err := dfl.Parse(exp)
		if err != nil {
			return errors.Wrap(err, "error processing filter expression "+exp)
		}

		pipeline = append(pipeline, dfl.Function{Name: "filter", MultiOperator: &dfl.MultiOperator{Arguments: []dfl.Node{
			dfl.Attribute{Name: ""},
			dfl.Literal{Value: exp},
		}}})
	} else {
		pipeline = append(pipeline, dfl.Attribute{Name: ""})
	}

	limit, err := qs.FirstInt("limit")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *request.ErrQueryStringParameterMissing:
		default:
			return err
		}
	} else {
		pipeline = append(pipeline, named.Limit)
	}

	if ext == "geojson" || ext == "toml" {
		pipeline = append(pipeline, named.GeoJSONLinesToGeoJSON)
	}

	// Input Flags
	inputReaderBufferSize := h.Viper.GetInt("input-reader-buffer-size")
	inputPassphrase := h.Viper.GetString("input-passphrase")
	inputSalt := h.Viper.GetString("input-salt")

	_, inputUriString, err := dfl.EvaluateString(layer.DataStore.Uri, map[string]interface{}{}, map[string]interface{}{}, h.DflFuncs, dfl.DefaultQuotes)
	if err != nil {
		return errors.Wrap(err, "error evaluating datastore uri")
	}

	var s3_client *s3.S3
	if strings.HasPrefix(inputUriString, "s3://") {
		client, err := h.GetAWSS3Client()
		if err != nil {
			return errors.Wrap(err, "error connecting to AWS")
		}
		s3_client = client
	}

	inputFormat := layer.DataStore.Format
	inputCompression := layer.DataStore.Compression

	inputReader, inputMetadata, err := grw.ReadFromResource(
		inputUriString,
		inputCompression,
		inputReaderBufferSize,
		false,
		s3_client)
	if err != nil {
		return errors.Wrap(err, "error opening resource from uri "+inputUriString)
	}

	if len(inputFormat) == 0 {
		if inputMetadata != nil {
			if len(inputMetadata.ContentType) > 0 {
				switch inputMetadata.ContentType {
				case "application/json":
					inputFormat = "json"
				case "application/vnd.geo+json":
					inputFormat = "json"
				case "application/toml":
					inputFormat = "toml"
				}
			}
		}
		if len(inputFormat) == 0 || len(inputCompression) == 0 {
			_, inputPath := grw.SplitUri(inputUriString)
			_, inputFormatGuess, _ := util.SplitNameFormatCompression(inputPath)
			if len(inputFormat) == 0 {
				inputFormat = inputFormatGuess
			}
			/* Has no effect since too late, since already created grw.Reader
			  if len(inputCompression) == 0 {
				inputCompression = inputCompressionGuess
			}*/
		}
		if len(inputFormat) == 0 {
			return errors.New("Error: Provided no --input-format and could not infer from resource.\nRun \"railgun --help\" for more information.")
		}
	}

	inputBytes, err := util.DecryptReader(inputReader, inputPassphrase, inputSalt)
	if err != nil {
		return errors.Wrap(err, "error decoding input")
	}

	outputType, err := gss.GetType(inputBytes, inputFormat)
	if err != nil {
		return errors.Wrap(err, "error decoding input")
	}

	inputObject, err := gss.DeserializeBytes(
		inputBytes,
		inputFormat,
		gss.NoHeader,
		"",
		false,
		gss.NoSkip,
		gss.NoLimit,
		outputType,
		false,
		false)
	if err != nil {
		return errors.Wrap(err, "error deserializing input")
	}

	_, outputObject, err := dfl.Pipeline{Nodes: pipeline}.Evaluate(
		map[string]interface{}{"limit": limit},
		inputObject,
		h.DflFuncs,
		dfl.DefaultQuotes)
	if err != nil {
		return errors.Wrap(err, "error processing features")
	}

	outputBytes, err := gss.SerializeBytes(gss.StringifyMapKeys(outputObject), outputFormat, []string{}, gss.NoLimit)
	if err != nil {
		return errors.Wrap(err, "error converting output")
	}
	w.Write(outputBytes) // #nosec

	return nil

}
