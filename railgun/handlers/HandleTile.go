package handlers

import (
	"fmt"
	//"math"
	"strconv"
	"strings"
	//"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-reader/reader"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spatialcurrent/railgun/railgun/geo"
	"github.com/spatialcurrent/railgun/railgun/railgunerrors"
	"github.com/spf13/viper"
	"net/http"
)

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func respondWith404AndEmptyFeatureCollection(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("{\"type\":\"FeatureCollection\",\"features\":[]}"))
}

func respondWith500AndEmptyFeatureCollection(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("{\"type\":\"FeatureCollection\",\"features\":[]}"))
}

func respondWithEmptyFeatureCollection(w http.ResponseWriter) {
	w.Write([]byte("{\"type\":\"FeatureCollection\",\"features\":[]}"))
}

func HandleTile(v *viper.Viper, w http.ResponseWriter, r *http.Request, vars map[string]string, qs railgun.QueryString, requests chan railgun.Request, messages chan interface{}, errorsChannel chan error, collectionsList []railgun.Collection, collectionsByName map[string]railgun.Collection) {

	verbose := v.GetBool("verbose")
	
	tileRequest := &railgun.TileRequest{Collection: vars["name"], Header: r.Header}
	// Defer putting tile request into requests channel, so it can pick up more metadata during execution
  defer func() {
    requests <- tileRequest
  }()

	collection, ok := collectionsByName[vars["name"]]
	if !ok {
		errorsChannel <- &railgunerrors.ErrMissingCollection{Name: vars["name"]}
		return
	}

	_, outputFormat, _ := railgun.SplitNameFormatCompression(r.URL.Path)

	exp, err := qs.FirstString("dfl")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *railgun.ErrQueryStringParameterNotExist:
		default:
			errorsChannel <- err
			return
		}
	}

	z, err := strconv.Atoi(vars["z"])
	if err != nil {
		errorsChannel <- &railgunerrors.ErrInvalidParameter{Name: "z", Value: vars["z"]}
		return
	}
	tileRequest.Z = z

	x, err := strconv.Atoi(vars["x"])
	if err != nil {
		errorsChannel <- &railgunerrors.ErrInvalidParameter{Name: "x", Value: vars["x"]}
		return
	}
	tileRequest.X = x

	y, err := strconv.Atoi(vars["y"])
	if err != nil {
		errorsChannel <- &railgunerrors.ErrInvalidParameter{Name: "y", Value: vars["y"]}
		return
	}
	tileRequest.Y = y

	/*ctx := map[string]interface{}{
	  "z": 10, // z
	  "x": int(float64(x) * math.Pow(2.0, float64(10 - z))),
	  "y": int(float64(y) * math.Pow(2.0, float64(10 - z)))}*/
	ctx := map[string]interface{}{"z": z, "x": x, "y": y}
	_, inputUri, err := collection.DataStore.Uri.Evaluate(map[string]interface{}{}, ctx, dfl.NewFuntionMapWithDefaults(), dfl.DefaultQuotes)
	if err != nil {
		errorsChannel <- errors.Wrap(err, "error evaluating datastore uri with context "+fmt.Sprint(ctx))
		return
	}

	inputUriString := ""
	switch inputUri.(type) {
	case string:
		inputUriString = inputUri.(string)
	default:
		respondWithEmptyFeatureCollection(w)
		return
	}
	
	tileRequest.Source = inputUriString

	bbox := geo.TileToBoundingBox(z, x, geo.FlipY(y, z, 256, geo.WebMercatorExtent, geo.WebMercatorResolutions))

	g := "filter(@, '((@_tile_z == "+fmt.Sprint(z)+") and (@_tile_x == " + fmt.Sprint(x) + ") and (@_tile_y == " + fmt.Sprint(y) + ")) or ((@geometry?.coordinates != null) and (($c := @geometry.coordinates) | ($c[0] within " + fmt.Sprint(bbox[0]) + " and " + fmt.Sprint(bbox[2]) + ") and ($c[1] within " + fmt.Sprint(bbox[1]) + " and " + fmt.Sprint(bbox[3]) + ")))')"

	if len(exp) > 0 {
		exp = g + " | " + exp
	} else {
		exp = g
	}
	
	tileRequest.Expression = exp

	limit, err := qs.FirstInt("limit")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *railgun.ErrQueryStringParameterNotExist:
		default:
			errorsChannel <- err
			return
		}
	} else {
		exp += " | limit(@, " + fmt.Sprint(limit) + ")"
	}

	if vars["ext"] == "geojson" {
		exp += " | map(@, '@properties -= {`_tile_x`, `_tile_y`, `_tile_z`}') | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}"
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

	var aws_session *session.Session
	var s3_client *s3.S3

	if strings.HasPrefix(inputUriString, "s3://") {
	  if verbose {
	    fmt.Println("Connecting to AWS with AWS_ACCESS_KEY_ID "+awsAccessKeyId)
	  }
		aws_session = railgun.ConnectToAWS(awsAccessKeyId, awsSecretAccessKey, awsSessionToken, awsDefaultRegion)
		s3_client = s3.New(aws_session)
	}

	inputFormat := collection.DataStore.Format
	inputCompression := collection.DataStore.Compression

	var inputBytesPlain *[]byte
	item, found := collection.Cache.Get(inputUriString)
	if found {
		inputBytesPlain = item.(*[]byte)
		if len(*inputBytesPlain) == 0 {
			if verbose {
				messages <- "requested tile " + fmt.Sprint(z) + "-" + fmt.Sprint(x) + "-" + fmt.Sprint(y) + " had no data in cache"
			}
			respondWithEmptyFeatureCollection(w)
			return
		}
	}

	if !found {
		inputReader, inputMetadata, err := reader.OpenResource(
			inputUriString,
			inputCompression,
			inputReaderBufferSize,
			false,
			s3_client)
		if err != nil {
			if verbose {
				messages <- "when opening resource " + fmt.Sprint(z) + "-" + fmt.Sprint(x) + "-" + fmt.Sprint(y) + " had error " + err.Error()
			}
			respondWithEmptyFeatureCollection(w)
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
			if len(inputFormat) == 0 {
				messages <- "Error: Provided no --input-format and could not infer from resource.\nRun \"railgun --help\" for more information."
				return
			}
			collection.DataStore.Format = inputFormat
		}

		inputBytesEncrypted, err := inputReader.ReadAll()
		if err != nil {
			if verbose {
				messages <- "when opening resource " + fmt.Sprint(z) + "-" + fmt.Sprint(x) + "-" + fmt.Sprint(y) + " had error reading bytes"
			}
			errorsChannel <- errors.New("error reading from resource")
			return
		}

		b, err := railgun.DecryptInput(inputBytesEncrypted, inputPassphrase, inputSalt)
		if err != nil {
			errorsChannel <- errors.Wrap(err, "error decoding input")
			return
		}
		inputBytesPlain = &b

		collection.Cache.Set(inputUriString, inputBytesPlain, cache.DefaultExpiration)

	}
	
	if len(outputFormat) == 0 {
	  w.Write(*inputBytesPlain)
	  return
	}

	outputObject, err := railgun.ProcessObject(
		*inputBytesPlain,
		inputFormat,
		[]string{},
		"",
		false,
		-1,
		exp,
		"",
		verbose)
		
	outputObject = gss.StringifyMapKeys(outputObject)
	
	tileRequest.Features = int(dfl.TryGetInt64(outputObject, "numberOfFeatures", 0))

	outputBytes, err := gss.SerializeBytes(outputObject, outputFormat, []string{}, -1)
	if err != nil {
		errorsChannel <- errors.Wrap(err, "error converting output")
		respondWith500AndEmptyFeatureCollection(w)
		return
	}
	w.Write(outputBytes)
	return

}
