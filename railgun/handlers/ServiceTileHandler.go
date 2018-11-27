// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"context"
	"net/http"
	"reflect"
	"strings"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-try-get/gtg"
	"github.com/spatialcurrent/railgun/railgun/core"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/geo"
	"github.com/spatialcurrent/railgun/railgun/pipeline"
	"github.com/spatialcurrent/railgun/railgun/request"
	"github.com/spatialcurrent/railgun/railgun/util"
)

var emptyFeatureCollection = map[string]interface{}{
	"type":             "FeatureCollection",
	"features":         []interface{}{},
	"numberOfFeatures": 0,
}

type ServiceTileHandler struct {
	*BaseHandler
	Cache *gocache.Cache
}

func (h *ServiceTileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := context.WithValue(r.Context(), "handler", map[string]interface{}{
		"name": reflect.TypeOf(h).Elem().Name(),
	})

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)

	switch r.Method {
	case "GET":
		h.Catalog.Lock()
		obj, err := h.Get(w, r.WithContext(ctx), format)
		h.Catalog.Unlock()
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

func (h *ServiceTileHandler) Get(w http.ResponseWriter, r *http.Request, format string) (interface{}, error) {

	vars := mux.Vars(r)
	ctx := context.WithValue(r.Context(), "vars", vars)

	defer func() {
		h.SendMessage(map[string]interface{}{
			"request": ctx.Value("request"),
			"handler": ctx.Value("handler"),
			"vars":    ctx.Value("vars"),
		})
	}()

	qs := request.NewQueryString(r)

	serviceName, ok := vars["name"]
	if !ok {
		return nil, &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}

	tileRequest := &request.TileRequest{Layer: serviceName, Header: r.Header}
	cacheRequest := &request.CacheRequest{}
	// Defer putting tile request into requests channel, so it can pick up more metadata during execution
	defer func() {
		h.Requests <- tileRequest
		if len(cacheRequest.Key) > 0 {
			h.Requests <- cacheRequest
		}
	}()

	service, ok := h.Catalog.GetService(serviceName)
	if !ok {
		return nil, &rerrors.ErrMissingObject{Type: "service", Name: serviceName}
	}

	tileRequest.Expression = service.Process.Node.Dfl(dfl.DefaultQuotes, true, 0)

	tile, err := core.NewTileFromRequestVars(vars)
	if err != nil {
		return nil, err
	}
	tileRequest.Tile = tile

	// if outside data store extent return empty feature collection
	if maxExtent := service.DataStore.Extent; len(maxExtent) > 0 {
		minX := geo.LongitudeToTile(maxExtent[0], tile.Z)
		minY := geo.LatitudeToTile(maxExtent[3], tile.Z) // flip y
		maxX := geo.LongitudeToTile(maxExtent[2], tile.Z)
		maxY := geo.LatitudeToTile(maxExtent[1], tile.Z) // flip y
		if tile.X < minX || tile.X > maxX || tile.Y < minY || tile.Y > maxY {
			tileRequest.OutsideExtent = true
			return emptyFeatureCollection, nil
		}
	}

	buffer, err := qs.FirstInt("buffer")
	if err != nil {
		switch errors.Cause(err).(type) {
		case *request.ErrQueryStringParameterMissing:
			buffer = 1
		default:
			return nil, err
		}
	}

	tileRequest.Bbox = tile.Bbox()

	bufferedBoundingBox := []float64{
		geo.TileToLongitude(tile.X-buffer, tile.Z),
		geo.TileToLatitude(tile.Y+1+buffer, tile.Z),
		geo.TileToLongitude(tile.X+1+buffer, tile.Z),
		geo.TileToLatitude(tile.Y-buffer, tile.Z),
	}

	variables := h.AggregateMaps(
		h.GetServiceVariables(h.Cache, serviceName),
		service.Defaults,
		tile.Map())
	variables["bbox"] = bufferedBoundingBox
	//variables["limit"] = limit

	_, inputUri, err := dfl.EvaluateString(service.DataStore.Uri, variables, map[string]interface{}{}, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
	if err != nil {
		return emptyFeatureCollection, nil
	}

	inputScheme, inputPath := grw.SplitUri(inputUri)

	tileRequest.Source = inputUri
	cacheRequest.Key = inputUri

	p := pipeline.New().FilterBoundingBox().Next(service.Process.Node)

	cacheKeyDataStore := ""
	var inputReader grw.ByteReadCloser
	var inputObject interface{}

	var s3Client *s3.S3
	if inputScheme == "s3" {
		s3Client, err = h.GetAWSS3Client()
		if err != nil {
			return nil, errors.Wrap(err, "error connecting to AWS")
		}

		i := strings.Index(inputPath, "/")
		if i == -1 {
			return nil, errors.New("path missing bucket")
		}

		bucket := inputPath[0:i]
		key := inputPath[i+1:]
		headObjectOutput, err := s3Client.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return nil, errors.Wrap(err, "error heading S3 object")
		}

		cacheKeyDataStore = h.BuildCacheKeyDataStore(
			service.DataStore.Name,
			inputUri,
			*headObjectOutput.LastModified)

		if object, found := h.Cache.Get(cacheKeyDataStore); found {
			inputObject = object
			inputReader = nil
			cacheRequest.Hit = true
		} else {
			inputObject = nil
			inputReader = nil
			cacheRequest.Hit = false
		}

	} else if inputScheme == "file" || inputScheme == "" {
		inputFile, inputFileInfo, err := grw.ExpandOpenAndStat(inputPath)
		if err != nil {
			return nil, err
		}

		cacheKeyDataStore = h.BuildCacheKeyDataStore(
			service.DataStore.Name,
			inputUri,
			inputFileInfo.ModTime())

		if object, found := h.Cache.Get(cacheKeyDataStore); found {
			inputObject = object
			inputReader = nil
			cacheRequest.Hit = true
		} else {
			inputObject = nil
			cacheRequest.Hit = false
			r, err := grw.ReadFromFile(inputFile, service.DataStore.Compression, false, 4096)
			if err != nil {
				return nil, errors.Wrap(err, "error creating grw.ByteReadCloser for file at path \""+inputPath+"\"")
			}
			inputReader = r
		}
	}

	if inputObject == nil {
		inputBytes := make([]byte, 0)
		if inputReader != nil {
			b, err := inputReader.ReadAllAndClose()
			if err != nil {
				return nil, errors.Wrap(err, "error reading from resource at uri "+inputUri)
			}
			inputBytes = b
		} else {
			b, err := grw.ReadAllAndClose(inputUri, service.DataStore.Compression, s3Client)
			if err != nil {
				return nil, errors.Wrap(err, "error reading from resource at uri "+inputUri)
			}
			inputBytes = b
		}

		object, err := h.DeserializeBytes(inputBytes, service.DataStore.Format)
		if err != nil {
			return nil, errors.Wrap(err, "error deserializing input")
		}
		inputObject = object
	}

	h.Cache.Set(cacheKeyDataStore, inputObject, gocache.DefaultExpiration)

	variables, outputObject, err := p.Evaluate(
		variables,
		inputObject)
	if err != nil {
		return nil, errors.Wrap(err, "error processing features")
	}

	tileRequest.Features = gtg.TryGetInt(outputObject, "numberOfFeatures", 0)

	h.SetServiceVariables(h.Cache, serviceName, variables)

	return gss.StringifyMapKeys(outputObject), nil

}
