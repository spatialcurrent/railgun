// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-try-get/gtg"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spatialcurrent/railgun/railgun/geo"
	"github.com/spatialcurrent/railgun/railgun/img"
	"github.com/spatialcurrent/railgun/railgun/named"
	"github.com/spatialcurrent/railgun/railgun/railgunerrors"
	"image/color"
	"net/http"
	"strings"
)

var emptyFeatureCollection = []byte("{\"type\":\"FeatureCollection\",\"features\":[]}")

func respondWith404AndEmptyFeatureCollection(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write(emptyFeatureCollection)
}

func respondWith500AndEmptyFeatureCollection(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(emptyFeatureCollection)
}

func respondWithEmptyFeatureCollection(w http.ResponseWriter) {
	w.Write(emptyFeatureCollection)
}

type TileHandler struct {
	*BaseHandler
}

func (h *TileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	qs := railgun.NewQueryString(r)
	err := h.Run(w, r, vars, qs)
	if err != nil {
		h.Errors <- err
		w.WriteHeader(http.StatusInternalServerError)
		img.RespondWithImage(vars["ext"], w, img.CreateImage(color.RGBA{255, 0, 0, 220}))
	}
}

func (h *TileHandler) Run(w http.ResponseWriter, r *http.Request, vars map[string]string, qs railgun.QueryString) error {

	config := h.Config

	ext := vars["ext"]

	layerName, ok := vars["name"]
	if !ok {
		return &railgunerrors.ErrMissingRequiredParameter{Name: "name"}
	}

	tileRequest := &railgun.TileRequest{Layer: layerName, Header: r.Header}
	cacheRequest := &railgun.CacheRequest{}
	// Defer putting tile request into requests channel, so it can pick up more metadata during execution
	defer func() {
		h.Requests <- tileRequest
		if len(cacheRequest.Key) > 0 {
			h.Requests <- cacheRequest
		}
	}()

	layer, ok := h.Config.GetLayer(layerName)
	if !ok {
		return &railgunerrors.ErrMissingObject{Type: "layer", Name: layerName}
	}

	_, outputFormat, _ := railgun.SplitNameFormatCompression(r.URL.Path)

	tile, err := railgun.NewTileFromRequestVars(vars)
	if err != nil {
		return err
	}
	tileRequest.Tile = tile

	// if outside layer extent return empty feature collection
	if maxExtent := layer.Extent; len(maxExtent) > 0 {
		minX := geo.LongitudeToTile(maxExtent[0], tile.Z)
		minY := geo.LatitudeToTile(maxExtent[3], tile.Z) // flip y
		maxX := geo.LongitudeToTile(maxExtent[2], tile.Z)
		maxY := geo.LatitudeToTile(maxExtent[1], tile.Z) // flip y
		fmt.Println(minX, minY, maxX, maxY)
		if tile.X < minX || tile.X > maxX || tile.Y < minY || tile.Y > maxY {
			tileRequest.OutsideExtent = true
			respondWithEmptyFeatureCollection(w)
			return nil
		}
	}

	// if outside data store extent return empty feature collection
	if maxExtent := layer.DataStore.Extent; len(maxExtent) > 0 {
		minX := geo.LongitudeToTile(maxExtent[0], tile.Z)
		minY := geo.LatitudeToTile(maxExtent[3], tile.Z) // flip y
		maxX := geo.LongitudeToTile(maxExtent[2], tile.Z)
		maxY := geo.LatitudeToTile(maxExtent[1], tile.Z) // flip y
		fmt.Println(minX, minY, maxX, maxY)
		if tile.X < minX || tile.X > maxX || tile.Y < minY || tile.Y > maxY {
			tileRequest.OutsideExtent = true
			respondWithEmptyFeatureCollection(w)
			return nil
		}
	}

	ctx := tile.Map()
	_, inputUriString, err := dfl.EvaluateString(layer.DataStore.Uri, map[string]interface{}{}, ctx, h.DflFuncs, dfl.DefaultQuotes)
	if err != nil {
		respondWithEmptyFeatureCollection(w)
		return errors.Wrap(err, "error evaluating datastore uri with context "+fmt.Sprint(ctx))
	}

	tileRequest.Source = inputUriString
	cacheRequest.Key = inputUriString

	buffer, err := qs.FirstInt("buffer")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *railgun.ErrQueryStringParameterNotExist:
		default:
			return err
		}
	}

	//bbox := geo.TileToBoundingBox(z, x, geo.FlipY(y, z, 256, geo.WebMercatorExtent, geo.WebMercatorResolutions))
	//bbox := geo.TileToBoundingBox(z, x, y)
	//bbox := tile.Bbox()

	//fmt.Println("Bounding box:", bbox)

	tileRequest.Bbox = tile.Bbox()

	//g := "filter(@, '((@_tile_z == "+fmt.Sprint(z)+") and (@_tile_x == " + fmt.Sprint(x) + ") and (@_tile_y == " + fmt.Sprint(y) + ")) or ((@geometry?.coordinates != null) and (($c := @geometry.coordinates) | ($c[0] within " + fmt.Sprint(bbox[0]) + " and " + fmt.Sprint(bbox[2]) + ") and ($c[1] within " + fmt.Sprint(bbox[1]) + " and " + fmt.Sprint(bbox[3]) + ")))')"

	/*g := "filter(@, '(@geometry?.coordinates != null) and (($c := @geometry.coordinates) | ($c[0] within " + fmt.Sprint(bbox[0]) + " and " + fmt.Sprint(bbox[2]) + ") and ($c[1] within " + fmt.Sprint(bbox[1]) + " and " + fmt.Sprint(bbox[3]) + "))')"

	if len(exp) > 0 {
		exp = g + " | " + exp
	} else {
		exp = g
	}*/

	pipeline := []dfl.Node{named.GeometryFilter}

	exp, err := qs.FirstString("dfl")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *railgun.ErrQueryStringParameterNotExist:
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

		tileRequest.Expression = exp
	}

	limit, err := qs.FirstInt("limit")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *railgun.ErrQueryStringParameterNotExist:
		default:
			return err
		}
	} else {
		pipeline = append(pipeline, named.Limit)
	}

	if ext == "geojson" || ext == "toml" {
		pipeline = append(pipeline, named.GeoJSONLinesToGeoJSON)
	}

	// AWS Flags
	awsDefaultRegion := config.GetString("aws-default-region")
	awsAccessKeyId := config.GetString("aws-access-key-id")
	awsSecretAccessKey := config.GetString("aws-secret-access-key")
	awsSessionToken := config.GetString("aws-session-token")

	// Input Flags
	inputReaderBufferSize := config.GetInt("input-reader-buffer-size")
	inputPassphrase := config.GetString("input-passphrase")
	inputSalt := config.GetString("input-salt")

	var awsSession *session.Session
	var s3_client *s3.S3

	verbose := config.GetBool("verbose")

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

	hit, inputObject, err := layer.Cache.Get(
		inputUriString,
		layer.DataStore.Format,
		layer.DataStore.Compression,
		inputReaderBufferSize,
		inputPassphrase,
		inputSalt,
		s3_client,
		verbose)
	if err != nil {
		return errors.Wrap(err, "error getting data from cache for tile "+tile.String())
	}
	cacheRequest.Hit = hit

	bufferedBoundingBox := []float64{
		geo.TileToLongitude(tile.X-buffer, tile.Z),
		geo.TileToLatitude(tile.Y+1+buffer, tile.Z),
		geo.TileToLongitude(tile.X+1+buffer, tile.Z),
		geo.TileToLatitude(tile.Y-buffer, tile.Z),
	}

	_, outputObject, err := dfl.Pipeline{Nodes: pipeline}.Evaluate(
		map[string]interface{}{"bbox": bufferedBoundingBox, "limit": limit},
		inputObject,
		h.DflFuncs,
		dfl.DefaultQuotes)
	if err != nil {
		return errors.Wrap(err, "error processing features")
	}

	tileRequest.Features = gtg.TryGetInt(outputObject, "numberOfFeatures", 0)

	outputBytes, err := gss.SerializeBytes(gss.StringifyMapKeys(outputObject), outputFormat, []string{}, gss.NoLimit)
	if err != nil {
		return errors.Wrap(err, "error converting output")
	}
	w.Write(outputBytes)

	return nil
}
