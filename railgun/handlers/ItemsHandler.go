package handlers

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spatialcurrent/railgun/railgun/railgunerrors"
	"net/http"
	"strings"
)

type ItemsHandler struct {
	*BaseHandler
	CollectionsList   []railgun.Collection
	CollectionsByName map[string]railgun.Collection
	AwsSessionCache   *gocache.Cache
	DflFuncs          dfl.FunctionMap
}

func (h *ItemsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	v := h.Viper

	vars := mux.Vars(r)
	qs := railgun.NewQueryString(r)

	verbose := v.GetBool("verbose")

	collection, ok := h.CollectionsByName[vars["name"]]
	if !ok {
		h.Errors <- &railgunerrors.ErrMissingCollection{Name: vars["name"]}
		return
	}

	_, outputFormat, _ := railgun.SplitNameFormatCompression(r.URL.Path)

	exp, err := qs.FirstString("dfl")
	if err != nil {
		h.Errors <- err
		return
	}

	limit, err := qs.FirstInt("limit")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *railgun.ErrQueryStringParameterNotExist:
		default:
			h.Errors <- err
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
	inputReaderBufferSize := v.GetInt("input-reader-buffer-size")
	inputPassphrase := v.GetString("input-passphrase")
	inputSalt := v.GetString("input-salt")

	_, inputUriString, err := dfl.EvaluateString(collection.DataStore.Uri, map[string]interface{}{}, map[string]interface{}{}, h.DflFuncs, dfl.DefaultQuotes)
	if err != nil {
		respondWithEmptyFeatureCollection(w)
		h.Errors <- errors.Wrap(err, "error evaluating datastore uri")
		return
	}

	var awsSession *session.Session
	var s3_client *s3.S3

	if strings.HasPrefix(inputUriString, "s3://") {
		if verbose {
			fmt.Println("Connecting to AWS with AWS_ACCESS_KEY_ID " + awsAccessKeyId)
		}
		s, found := h.AwsSessionCache.Get(awsAccessKeyId + "\n" + awsSessionToken)
		if found {
			awsSession = s.(*session.Session)
		} else {
			awsSession = railgun.ConnectToAWS(awsAccessKeyId, awsSecretAccessKey, awsSessionToken, awsDefaultRegion)
			h.AwsSessionCache.Set(awsAccessKeyId+"\n"+awsSessionToken, awsSession, gocache.DefaultExpiration)
		}
		s3_client = s3.New(awsSession)
	}

	inputFormat := collection.DataStore.Format
	inputCompression := collection.DataStore.Compression

	inputReader, inputMetadata, err := grw.ReadFromResource(
		inputUriString,
		inputCompression,
		inputReaderBufferSize,
		false,
		s3_client)
	if err != nil {
		h.Errors <- errors.Wrap(err, "error opening resource from uri "+inputUriString)
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
			_, inputPath := grw.SplitUri(inputUriString)
			_, inputFormatGuess, inputCompressionGuess := railgun.SplitNameFormatCompression(inputPath)
			if len(inputFormat) == 0 {
				inputFormat = inputFormatGuess
			}
			if len(inputCompression) == 0 {
				inputCompression = inputCompressionGuess
			}
		}
		if len(inputFormat) == 0 {
			h.Errors <- errors.New("Error: Provided no --input-format and could not infer from resource.\nRun \"railgun --help\" for more information.")
			return
		}
	}

	inputBytesEncrypted, err := inputReader.ReadAll()
	if err != nil {
		h.Errors <- errors.New("error reading from resource")
		return
	}

	inputStringPlain, err := railgun.DecryptInput(inputBytesEncrypted, inputPassphrase, inputSalt)
	if err != nil {
		h.Errors <- errors.Wrap(err, "error decoding input")
		return
	}

	str, err := railgun.ProcessInput(
		inputStringPlain,
		inputFormat,
		[]string{},
		"",
		false,
		-1,
		exp,
		map[string]interface{}{},
		"",
		outputFormat,
		[]string{},
		-1,
		verbose)
	if err != nil {
		panic(err)
	}
	w.Write([]byte(str))
	return

}
