package handlers

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-reader/reader"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spf13/viper"
	"net/http"
	"strings"
)

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func HandleItems(v *viper.Viper, w http.ResponseWriter, r *http.Request, vars map[string]string, qs railgun.QueryString, messages chan interface{}, collectionsList []railgun.Collection, collectionsByName map[string]railgun.Collection) {

	verbose := viper.GetBool("verbose")

	collection, ok := collectionsByName[vars["name"]]
	if !ok {
		messages <- errors.New("invalid name")
		return
	}

	_, outputFormat, _ := railgun.SplitNameFormatCompression(r.URL.Path)

	exp, err := qs.FirstString("dfl")
	if err != nil {
		messages <- err
		return
	}

	limit, err := qs.FirstInt("limit")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *railgun.ErrQueryStringParameterNotExist:
		default:
			messages <- err
			return
		}
	} else {
		exp += " | limit(@, " + fmt.Sprint(limit) + ")"
	}

	if vars["ext"] == "geojson" {
		exp += " | {type:FeatureCollection, features:@}"
	}

	// AWS Flags
	awsDefaultRegion := v.GetString("aws-default-region")
	awsAccessKeyId := v.GetString("aws-access-key-id")
	awsSecretAccessKey := v.GetString("aws-secret-access-key")
	awsSessionToken := v.GetString("aws-session-token")

	// Input Flags
	inputReaderBufferSize := viper.GetInt("input-reader-buffer-size")
	inputPassphrase := viper.GetString("input-passphrase")
	inputSalt := viper.GetString("input-salt")

	var aws_session *session.Session
	var s3_client *s3.S3

  inputUri, err := dfl.EvaluateString(collection.DataStore.Uri, map[string]interface{}{}, dfl.NewFuntionMapWithDefaults(), dfl.DefaultQuotes)
  if err != nil {
    messages <- errors.Wrap(err, "error evaluating datastore uri ")
		return
  }

	if strings.HasPrefix(inputUri, "s3://") {
		aws_session = railgun.ConnectToAWS(awsAccessKeyId, awsSecretAccessKey, awsSessionToken, awsDefaultRegion)
		s3_client = s3.New(aws_session)
	}

	inputFormat := collection.DataStore.Format
	inputCompression := collection.DataStore.Compression

	inputReader, inputMetadata, err := reader.OpenResource(
		inputUri,
		inputCompression,
		inputReaderBufferSize,
		false,
		s3_client)
	if err != nil {
		messages <- errors.Wrap(err, "error opening resource from uri "+inputUri)
		return
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
			_, inputPath := reader.SplitUri(inputUri)
			_, inputFormatGuess, inputCompressionGuess := railgun.SplitNameFormatCompression(inputPath)
			if len(inputFormat) == 0 {
				inputFormat = inputFormatGuess
			}
			if len(inputCompression) == 0 {
				inputCompression = inputCompressionGuess
			}
		}
		if len(inputFormat) == 0 {
			messages <- "Error: Provided no --input-format and could not infer from resource.\nRun \"railgun --help\" for more information."
			return
		}
	}

	inputBytesEncrypted, err := inputReader.ReadAll()
	if err != nil {
		messages <- errors.New("error reading from resource")
		return
	}

	inputStringPlain, err := railgun.DecryptInput(inputBytesEncrypted, inputPassphrase, inputSalt)
	if err != nil {
		messages <- errors.Wrap(err, "error decoding input")
		return
	}

	str, err := railgun.ProcessInput(
		inputStringPlain,
		inputFormat,
		[]string{},
		"",
		-1,
		exp,
		"",
		outputFormat,
		verbose)
	if err != nil {
		panic(err)
	}
	w.Write([]byte(str))
	return

}
