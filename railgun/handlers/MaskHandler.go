package handlers

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spatialcurrent/railgun/railgun/geo"
	"github.com/spatialcurrent/railgun/railgun/img"
	"github.com/spatialcurrent/railgun/railgun/named"
	"github.com/spatialcurrent/railgun/railgun/railgunerrors"
	"image/color"
	"math"
	"net/http"
	"reflect"
	"strings"
)

type MaskHandler struct {
	*BaseHandler
	AwsSessionCache *gocache.Cache
	DflFuncs        dfl.FunctionMap
}

func (h *MaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	qs := railgun.NewQueryString(r)
	err := h.Run(w, r, vars, qs)
	if err != nil {
		h.Errors <- err
		w.WriteHeader(http.StatusInternalServerError)
		img.RespondWithImage(vars["ext"], w, img.CreateImage(color.RGBA{255, 0, 0, 220}))
	}
}

func (h *MaskHandler) Run(w http.ResponseWriter, r *http.Request, vars map[string]string, qs railgun.QueryString) error {

	ext := vars["ext"]

	tileRequest := &railgun.TileRequest{Layer: vars["name"], Header: r.Header}
	cacheRequest := &railgun.CacheRequest{}
	// Defer putting tile request into requests channel, so it can pick up more metadata during execution
	defer func() {
		h.Requests <- tileRequest
		if len(cacheRequest.Key) > 0 {
			h.Requests <- cacheRequest
		}
	}()

	layer, ok := h.Config.GetLayer(vars["name"])
	if !ok {
		return &railgunerrors.ErrMissing{Type: "layer", Name: vars["name"]}
	}

	tile, err := railgun.NewTileFromRequestVars(vars)
	if err != nil {
		return err
	}
	tileRequest.Tile = tile

	if maxExtent := layer.DataStore.Extent; len(maxExtent) > 0 {
		minX := geo.LongitudeToTile(maxExtent[0], tile.Z)
		minY := geo.LatitudeToTile(maxExtent[3], tile.Z) // flip y
		//minY := geo.FlipY(geo.LatitudeToTile(maxExtent[1], tile.Z), tile.Z, 256, geo.WebMercatorExtent, geo.WebMercatorResolutions)
		maxX := geo.LongitudeToTile(maxExtent[2], tile.Z)
		maxY := geo.LatitudeToTile(maxExtent[1], tile.Z) // flip y
		//maxY := geo.FlipY(geo.LatitudeToTile(maxExtent[1], tile.Z), tile.Z, 256, geo.WebMercatorExtent, geo.WebMercatorResolutions)
		fmt.Println(minX, minY, maxX, maxY)
		if tile.X < minX || tile.X > maxX || tile.Y < minY || tile.Y > maxY {
			tileRequest.OutsideExtent = true
			return img.RespondWithImage(ext, w, img.BlankImage)
		}
	}

	ctx := tile.Map()
	_, inputUri, err := layer.DataStore.Uri.Evaluate(map[string]interface{}{}, ctx, dfl.NewFuntionMapWithDefaults(), dfl.DefaultQuotes)
	if err != nil {
		return errors.Wrap(err, "error evaluating datastore uri with context "+fmt.Sprint(ctx))
	}

	inputUriString := ""
	switch inputUri.(type) {
	case string:
		inputUriString = inputUri.(string)
	default:
		return img.RespondWithImage(ext, w, img.BlankImage)
	}

	tileRequest.Source = inputUriString
	cacheRequest.Key = inputUriString

	bbox := tile.Bbox()
	tileRequest.Bbox = bbox

	pipeline := []dfl.Node{named.GeometryFilter}

	threshold, err := qs.FirstInt("threshold")
	if err != nil {
		return err
	}

	maskAlpha, err := qs.FirstInt("alpha")
	if err != nil {
		return err
	}

	maskZoom, err := qs.FirstInt("zoom")
	if err != nil {
		return err
	}

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

	pipeline = append(pipeline, named.GroupByTile)
	//pipeline = append(pipeline, named.Length)

	// AWS Flags
	awsDefaultRegion := h.Config.GetString("aws-default-region")
	awsAccessKeyId := h.Config.GetString("aws-access-key-id")
	awsSecretAccessKey := h.Config.GetString("aws-secret-access-key")
	awsSessionToken := h.Config.GetString("aws-session-token")

	// Input Flags
	inputReaderBufferSize := h.Config.GetInt("input-reader-buffer-size")
	inputPassphrase := h.Config.GetString("input-passphrase")
	inputSalt := h.Config.GetString("input-salt")

	var awsSession *session.Session
	var s3_client *s3.S3

	verbose := h.Config.GetBool("verbose")

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

	pow_diff := int(math.Pow(2.0, float64(maskZoom-tile.Z)))

	//maskBoundingBox := geo.TileToBoundingBox(maskZoom, tile.X*pow_diff, tile.Y*pow_diff)
	//fmt.Println("Mask BBOX:", maskBoundingBox)

	_, outputObject, err := dfl.EvaluateMap(
		dfl.Pipeline{Nodes: pipeline},
		map[string]interface{}{
			"bbox": bbox,
			"z":    maskZoom},
		inputObject,
		h.DflFuncs,
		dfl.DefaultQuotes)
	if err != nil {
		return errors.Wrap(err, "error processing features")
	}

	groups, ok := outputObject.(map[string]map[string][]interface{})
	if !ok {
		return &railgunerrors.ErrInvalidType{
			Type:  reflect.TypeOf(map[string]map[string][]interface{}{}),
			Value: groups}
	}

	grid := make([]uint8, 256*256)
	pixels_per_step := 256.0 / pow_diff
	for py := 0; py < 256; py++ {
		ty := (tile.Y * pow_diff) + int(py/pixels_per_step)
		for px := 0; px < 256; px++ {
			tx := (tile.X * pow_diff) + int(px/pixels_per_step)
			if _, ok := groups[fmt.Sprint(ty)]; ok {
				if features, ok := groups[fmt.Sprint(ty)][fmt.Sprint(tx)]; ok {
					if len(features) >= threshold {
						grid[(py*256)+px] = 1
					}
				}
			}
		}
	}

	return img.RespondWithGrid(ext, w, grid, 256, 256, color.RGBA{0, 0, 128, uint8(maskAlpha)}, color.RGBA{0, 0, 0, 0})

}
