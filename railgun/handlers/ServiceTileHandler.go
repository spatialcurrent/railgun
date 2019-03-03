// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
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
	"github.com/spatialcurrent/railgun/railgun/middleware"
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

	ctx := r.Context()

	vars := mux.Vars(r)

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)

	if v := ctx.Value("request"); v != nil {
		if req, ok := v.(middleware.Request); ok {
			req.Vars = vars
			req.Handler = reflect.TypeOf(h).Elem().Name()
			ctx = context.WithValue(ctx, "request", req)
		}
	}

	switch r.Method {
	case "GET":
		once := &sync.Once{}
		h.Catalog.ReadLock()
		defer once.Do(func() { h.Catalog.ReadUnlock() })
		h.SendDebug("read locked for " + r.URL.String())
		obj, err := h.Get(w, r.WithContext(ctx), format, vars)
		once.Do(func() { h.Catalog.ReadUnlock() })
		h.SendDebug("read unlocked for " + r.URL.String())
		if err != nil {
			h.SendError(err)
			err = h.RespondWithError(w, err, format)
			if err != nil {
				h.SendError(err)
				panic(err)
			}
		} else {
			err = h.RespondWithObject(w, http.StatusOK, obj, format, "")
			if err != nil {
				h.SendInfo(err)
				err = h.RespondWithError(w, err, format)
				if err != nil {
					h.SendError(err)
					panic(err)
				}
			}
		}
	default:
		err := h.RespondWithNotImplemented(w, format)
		if err != nil {
			h.SendError(err)
			panic(err)
		}
	}

}

func (h *ServiceTileHandler) Get(w http.ResponseWriter, r *http.Request, format string, vars map[string]string) (object interface{}, err error) {

	ctx := struct{ Context context.Context }{Context: r.Context()}

	// Randomly sleep to increase cache performance.
	inside := true
	hit := false
	if tileRandomDelay := h.Viper.GetInt("tile-random-delay"); tileRandomDelay > 0 {
		delay := time.Duration(rand.Intn(tileRandomDelay)) * time.Millisecond
		time.Sleep(delay)
	}

	defer func() {
		ctx.Context.Value("log").(*sync.Once).Do(func() {
			if v := ctx.Context.Value("request"); v != nil {
				if req, ok := v.(middleware.Request); ok {
					req.Error = err
					end := time.Now()
					req.End = &end
					profile := map[string]interface{}{}
					if d := ctx.Context.Value("profile_head"); d != nil {
						profile["head"] = d
					}
					if d := ctx.Context.Value("profile_read"); d != nil {
						profile["read"] = d
					}
					if d := ctx.Context.Value("profile_deserialize"); d != nil {
						profile["deserialize"] = d
					}
					m := map[string]interface{}{
						"request": req.Map(),
						"service": ctx.Context.Value("service"),
						"process": ctx.Context.Value("process"),
						"profile": profile,
						"cache": map[string]interface{}{
							"hit": hit,
						},
						"inside": inside,
					}
					datastore := map[string]interface{}{
						"name": ctx.Context.Value("datastore"),
					}
					if uri := ctx.Context.Value("uri"); uri != nil {
						datastore["uri"] = uri
					}
					m["datastore"] = datastore
					s3 := map[string]interface{}{}
					if bucket := ctx.Context.Value("bucket"); bucket != nil {
						s3["bucket"] = bucket
					}
					if key := ctx.Context.Value("key"); key != nil {
						s3["key"] = key
					}
					if len(s3) > 0 {
						m["s3"] = s3
					}
					h.SendInfo(m)
				}
			}
		})
	}()

	qs := request.NewQueryString(r)

	serviceName, ok := vars["name"]
	if !ok {
		err := &rerrors.ErrMissingRequiredParameter{Name: "name"}
		ctx.Context = context.WithValue(ctx.Context, "error", err)
		return nil, err
	}

	service, ok := h.Catalog.GetService(serviceName)
	if !ok {
		err := &rerrors.ErrMissingObject{Type: "service", Name: serviceName}
		ctx.Context = context.WithValue(ctx.Context, "error", err)
		return nil, err
	}

	ctx.Context = context.WithValue(ctx.Context, "service", service.Name)
	ctx.Context = context.WithValue(ctx.Context, "datastore", service.DataStore.Name)
	ctx.Context = context.WithValue(ctx.Context, "process", service.Process.Name)

	tileRequest := &request.TileRequest{Layer: serviceName, Header: r.Header}
	cacheRequest := &request.CacheRequest{}
	// Defer putting tile request into requests channel, so it can pick up more metadata during execution
	defer func() {
		h.Requests <- tileRequest
		if len(cacheRequest.Key) > 0 {
			h.Requests <- cacheRequest
		}
	}()

	tileRequest.Expression = service.Process.Node.Dfl(dfl.DefaultQuotes, true, 0)

	tile, err := core.NewTileFromRequestVars(vars)
	if err != nil {
		ctx.Context = context.WithValue(ctx.Context, "error", err)
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
			inside = false
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
		ctx.Context = context.WithValue(ctx.Context, "error", err)
		inside = false
		return emptyFeatureCollection, nil
	}

	ctx.Context = context.WithValue(ctx.Context, "uri", inputUri)

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

		ctx.Context = context.WithValue(ctx.Context, "bucket", bucket)
		ctx.Context = context.WithValue(ctx.Context, "key", key)

		start := time.Now()
		var headObjectOutput *s3.HeadObjectOutput
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for i := 0; i < 3; i++ {
				headObjectOutput, err = s3Client.HeadObjectWithContext(ctx.Context, &s3.HeadObjectInput{
					Bucket: aws.String(bucket),
					Key:    aws.String(key),
				})
				if err != nil {
					h.SendError(err)
				} else if headObjectOutput != nil {
					break
				}
				time.Sleep(1 * time.Second)
			}
			wg.Done()
		}()
		wg.Wait()

		ctx.Context = context.WithValue(ctx.Context, "profile_head", time.Now().Sub(start).String())
		if err != nil {
			h.SendError(errors.Wrap(err, fmt.Sprintf("error heading S3 object at bucket %s and key %s", bucket, key)))
			ctx.Context = context.WithValue(ctx.Context, "error", err)
			return emptyFeatureCollection, nil
		}

		cacheKeyDataStore = h.BuildCacheKeyDataStore(
			service.DataStore.Name,
			inputUri,
			*headObjectOutput.LastModified)

		if object, found := h.Cache.Get(cacheKeyDataStore); found {
			hit = true
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
			ctx.Context = context.WithValue(ctx.Context, "error", err)
			return nil, err
		}

		cacheKeyDataStore = h.BuildCacheKeyDataStore(
			service.DataStore.Name,
			inputUri,
			inputFileInfo.ModTime())

		if object, found := h.Cache.Get(cacheKeyDataStore); found {
			hit = true
			inputObject = object
			inputReader = nil
			cacheRequest.Hit = true
		} else {
			inputObject = nil
			cacheRequest.Hit = false
			r, err := grw.ReadFromFile(inputFile, service.DataStore.Compression, false, 4096)
			if err != nil {
				err := errors.Wrap(err, "error creating grw.ByteReadCloser for file at path \""+inputPath+"\"")
				ctx.Context = context.WithValue(ctx.Context, "error", err)
				return nil, err
			}
			inputReader = r
		}
	}

	if inputObject == nil {
		start := time.Now()
		if inputReader == nil {
			inputReader, _, err = grw.ReadFromResource(inputUri, service.DataStore.Compression, 4096, false, s3Client)
			if err != nil {
				return nil, err
			}
		}
		if service.DataStore.Format == "jsonl" {
			inputSlice := make([]interface{}, 0)
			decoder := json.NewDecoder(inputReader)
			for {
				object := map[string]interface{}{}
				err := decoder.Decode(&object)
				if err != nil {
					if err == io.EOF {
						break
					} else if err != nil {
						ctx.Context = context.WithValue(ctx.Context, "error", err)
						return nil, err
					}
				}
				inputSlice = append(inputSlice, object)
			}
			inputObject = inputSlice
		} else {
			inputBytes, err := inputReader.ReadAllAndClose()
			if err != nil {
				err := errors.Wrap(err, "error reading from resource at uri "+inputUri)
				ctx.Context = context.WithValue(ctx.Context, "error", err)
				return nil, err
			}
			inputSlice, err := h.DeserializeBytes(inputBytes, service.DataStore.Format)
			if err != nil {
				err := errors.Wrap(err, "error deserializing input")
				ctx.Context = context.WithValue(ctx.Context, "error", err)
				return nil, err
			}
			inputObject = inputSlice
		}
		ctx.Context = context.WithValue(ctx.Context, "profile_read", time.Now().Sub(start).String())
	}

	go h.Cache.Set(cacheKeyDataStore, inputObject, gocache.DefaultExpiration) // save variables to cache outside of request/response thread

	variables, outputObject, err := p.Evaluate(variables, inputObject)
	if err != nil {
		err := errors.Wrap(err, "error processing features")
		ctx.Context = context.WithValue(ctx.Context, "error", err)
		return nil, err
	}

	tileRequest.Features = gtg.TryGetInt(outputObject, "numberOfFeatures", 0)

	go h.SetServiceVariables(h.Cache, serviceName, variables) // save variables to cache outside of request/response thread

	return gss.StringifyMapKeys(outputObject), nil

}
