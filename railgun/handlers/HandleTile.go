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
	"github.com/spatialcurrent/go-reader/reader"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spatialcurrent/railgun/railgun/geo"
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

func respondWithEmptyFeatureCollection(w http.ResponseWriter) {
	w.Write([]byte("{\"type\":\"FeatureCollection\",\"features\":[]}"))
}

func HandleTile(v *viper.Viper, w http.ResponseWriter, r *http.Request, vars map[string]string, qs railgun.QueryString, messages chan interface{}, collectionsList []railgun.Collection, collectionsByName map[string]railgun.Collection) {

	verbose := v.GetBool("verbose")

	collection, ok := collectionsByName[vars["name"]]
	if !ok {
		messages <- errors.New("invalid name " + vars["name"])
		return
	}

	_, outputFormat, _ := railgun.SplitNameFormatCompression(r.URL.Path)

	exp, err := qs.FirstString("dfl")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *railgun.ErrQueryStringParameterNotExist:
		default:
			messages <- err
			return
		}
	}

	z, err := strconv.Atoi(vars["z"])
	if err != nil {
		messages <- err
		return
	}

	x, err := strconv.Atoi(vars["x"])
	if err != nil {
		messages <- err
		return
	}

	y, err := strconv.Atoi(vars["y"])
	if err != nil {
		messages <- err
		return
	}

	/*ctx := map[string]interface{}{
	  "z": 10, // z
	  "x": int(float64(x) * math.Pow(2.0, float64(10 - z))),
	  "y": int(float64(y) * math.Pow(2.0, float64(10 - z)))}*/
	ctx := map[string]interface{}{"z": z, "x": x, "y": y}
	_, inputUri, err := collection.DataStore.Uri.Evaluate(map[string]interface{}{}, ctx, dfl.NewFuntionMapWithDefaults(), dfl.DefaultQuotes)
	if err != nil {
		messages <- errors.Wrap(err, "error evaluating datastore uri with context "+fmt.Sprint(ctx))
		return
	}

	if verbose {
		messages <- "requesting " + fmt.Sprint(z) + "-" + fmt.Sprint(x) + "-" + fmt.Sprint(y) + " from " + fmt.Sprint(inputUri)
	}

	inputUriString := ""
	switch inputUri.(type) {
	case string:
		inputUriString = inputUri.(string)
	default:
		respondWithEmptyFeatureCollection(w)
		return
	}

	bbox := geo.TileToBoundingBox(z, x, geo.FlipY(y, z, 256, geo.WebMercatorExtent, geo.WebMercatorResolutions))

	g := "filter(@, '((@_tile_z == 10) and (@_tile_x == " + fmt.Sprint(x) + ") and (@_tile_y == " + fmt.Sprint(y) + ")) or ((@geometry?.coordinates != null) and (($c := @geometry.coordinates) | ($c[0] within " + fmt.Sprint(bbox[0]) + " and " + fmt.Sprint(bbox[2]) + ") and ($c[1] within " + fmt.Sprint(bbox[1]) + " and " + fmt.Sprint(bbox[3]) + ")))')"

	if len(exp) > 0 {
		exp = g + " | " + exp
	} else {
		exp = g
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
		exp += " | map(@, '@properties -= {`_tile_x`, `_tile_y`, `_tile_z`}') | {type:FeatureCollection, features:@}"
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

	if strings.HasPrefix(inputUriString, "s3://") {
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
			messages <- errors.New("error reading from resource")
			return
		}

		b, err := railgun.DecryptInput(inputBytesEncrypted, inputPassphrase, inputSalt)
		if err != nil {
			messages <- errors.Wrap(err, "error decoding input")
			return
		}
		inputBytesPlain = &b

		collection.Cache.Set(inputUriString, inputBytesPlain, cache.DefaultExpiration)

	}

	str, err := railgun.ProcessInput(
		*inputBytesPlain,
		inputFormat,
		[]string{},
		"",
		false,
		-1,
		exp,
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
