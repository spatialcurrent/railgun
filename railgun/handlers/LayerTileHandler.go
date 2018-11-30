// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-try-get/gtg"
	"github.com/spatialcurrent/railgun/railgun/core"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/geo"
	"github.com/spatialcurrent/railgun/railgun/pipeline"
	"github.com/spatialcurrent/railgun/railgun/request"
	"github.com/spatialcurrent/railgun/railgun/util"
)

var emptyFeatureCollectionBytes = []byte("{\"type\":\"FeatureCollection\",\"features\":[],\"numberOfFeatures\":0}")

func respondWith404AndEmptyFeatureCollection(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write(emptyFeatureCollectionBytes)
}

func respondWith500AndEmptyFeatureCollection(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(emptyFeatureCollectionBytes)
}

func respondWithEmptyFeatureCollection(w http.ResponseWriter) {
	w.Write(emptyFeatureCollectionBytes)
}

type LayerTileHandler struct {
	*BaseHandler
}

func (h *LayerTileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	vars := mux.Vars(r)

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)

	if m, ok := ctx.Value("request").(map[string]interface{}); ok {
		m["vars"] = vars
		ctx = context.WithValue(ctx, "request", m)
	}
	ctx = context.WithValue(ctx, "handler", reflect.TypeOf(h).Elem().Name())

	switch r.Method {
	case "GET":
		once := &sync.Once{}
		once.Do(func() { h.Catalog.ReadLock() })
		defer once.Do(func() { h.Catalog.ReadUnlock() })
		h.SendDebug("read locked for " + r.URL.String())
		obj, err := h.Get(w, r.WithContext(ctx), format, vars)
		once.Do(func() { h.Catalog.ReadUnlock() })
		h.SendDebug("read unlocked for " + r.URL.String())
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		} else {
			err = h.RespondWithObject(w, http.StatusOK, obj, format)
			if err != nil {
				h.Messages <- err
				err = h.RespondWithError(w, err, format)
				if err != nil {
					panic(err)
				}
			}
		}
	default:
		err := h.RespondWithNotImplemented(w, format)
		if err != nil {
			panic(err)
		}
	}

}

func (h *LayerTileHandler) Get(w http.ResponseWriter, r *http.Request, format string, vars map[string]string) (interface{}, error) {

	qs := request.NewQueryString(r)

	layerName, ok := vars["name"]
	if !ok {
		return nil, &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}

	tileRequest := &request.TileRequest{Layer: layerName, Header: r.Header}
	cacheRequest := &request.CacheRequest{}
	// Defer putting tile request into requests channel, so it can pick up more metadata during execution
	defer func() {
		h.Requests <- tileRequest
		if len(cacheRequest.Key) > 0 {
			h.Requests <- cacheRequest
		}
	}()

	layer, ok := h.Catalog.GetLayer(layerName)
	if !ok {
		return nil, &rerrors.ErrMissingObject{Type: "layer", Name: layerName}
	}

	tile, err := core.NewTileFromRequestVars(vars)
	if err != nil {
		return nil, err
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
			return emptyFeatureCollection, nil
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
			return emptyFeatureCollection, nil
		}
	}

	ctx := tile.Map()
	_, inputUriString, err := dfl.EvaluateString(layer.DataStore.Uri, map[string]interface{}{}, ctx, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		respondWithEmptyFeatureCollection(w)
		return nil, errors.Wrap(err, "error evaluating datastore uri with context "+fmt.Sprint(ctx))
	}

	tileRequest.Source = inputUriString
	cacheRequest.Key = inputUriString

	buffer, err := qs.FirstInt("buffer")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *request.ErrQueryStringParameterMissing:
		default:
			return nil, err
		}
	}

	tileRequest.Bbox = tile.Bbox()

	p := pipeline.New().FilterBoundingBox()

	exp, err := qs.FirstString("dfl")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *request.ErrQueryStringParameterMissing:
		default:
			return nil, err
		}
	}

	if layer.Node != nil {
		if len(exp) > 0 {
			userFilterNode, err := dfl.ParseCompile(exp)
			if err != nil {
				return nil, errors.Wrap(err, "error processing user filter expression "+exp)
			}
			p = p.FilterCustom(dfl.And{&dfl.BinaryOperator{Left: layer.Node, Right: userFilterNode}})
			tileRequest.Expression = exp
		} else {
			p = p.FilterCustom(layer.Node)
		}
	} else {
		if len(exp) > 0 {
			userFilterNode, err := dfl.ParseCompile(exp)
			if err != nil {
				return nil, errors.Wrap(err, "error processing user filter expression "+exp)
			}
			p = p.FilterCustom(userFilterNode)
			tileRequest.Expression = exp
		}
	}

	limit, err := qs.FirstInt("limit")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *request.ErrQueryStringParameterMissing:
		default:
			return nil, err
		}
	} else {
		p = p.Limit()
	}

	p = p.GeoJSON()

	// Input Flags
	inputReaderBufferSize := h.Viper.GetInt("input-reader-buffer-size")
	inputPassphrase := h.Viper.GetString("input-passphrase")
	inputSalt := h.Viper.GetString("input-salt")

	verbose := h.Viper.GetBool("verbose")

	var s3_client *s3.S3
	if strings.HasPrefix(inputUriString, "s3://") {
		client, err := h.GetAWSS3Client()
		if err != nil {
			return nil, errors.Wrap(err, "error connecting to AWS")
		}
		s3_client = client
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
		return nil, errors.Wrap(err, "error getting data from cache for tile "+tile.String())
	}
	cacheRequest.Hit = hit

	bufferedBoundingBox := []float64{
		geo.TileToLongitude(tile.X-buffer, tile.Z),
		geo.TileToLatitude(tile.Y+1+buffer, tile.Z),
		geo.TileToLongitude(tile.X+1+buffer, tile.Z),
		geo.TileToLatitude(tile.Y-buffer, tile.Z),
	}

	variables := map[string]interface{}{}
	for k, v := range layer.Defaults {
		variables[k] = v
	}
	variables["bbox"] = bufferedBoundingBox
	variables["limit"] = limit

	_, outputObject, err := p.Evaluate(
		variables,
		inputObject)
	if err != nil {
		return nil, errors.Wrap(err, "error processing features")
	}

	tileRequest.Features = gtg.TryGetInt(outputObject, "numberOfFeatures", 0)

	return gss.StringifyMapKeys(outputObject), nil

}
