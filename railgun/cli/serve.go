// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	//"image"
	//"image/color"
	//"image/draw"
	//"image/png"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

import (
	//"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	gorilla_handlers "github.com/gorilla/handlers"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-adaptive-functions/af"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-try-get/gtg"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spatialcurrent/railgun/railgun/handlers"
	//"github.com/spatialcurrent/railgun/railgun/img"
	"github.com/spatialcurrent/railgun/railgun/railgunerrors"
)

var serveViper = viper.New()

var emptyFeatureCollection = []byte("{\"type\":\"FeatureCollection\",\"features\":[]}")

func NewRouter(v *viper.Viper, funcs dfl.FunctionMap, datastores []string, errorWriter grw.ByteWriteCloser, logWriter grw.ByteWriteCloser, logFormat string, verbose bool) (*mux.Router, error) {

	collectionsList := make([]railgun.Collection, 0, len(datastores))
	collectionsByName := map[string]railgun.Collection{}
	for _, ds := range datastores {
		c := railgun.Collection{}
		if strings.HasPrefix(ds, "{") {
			_, m, err := dfl.ParseCompileEvaluateMap(ds, map[string]interface{}{}, map[string]interface{}{}, funcs, dfl.DefaultQuotes)
			if err != nil {
				return nil, errors.Wrap(err, "error parsing datastore argument "+ds)
			}
			uri, err := dfl.ParseCompile(gtg.TryGetString(m, "uri", ""))
			if err != nil {
				return nil, errors.Wrap(err, "error parsing datastore uri")
			}
			uriSuffix := ""
			switch concat := uri.(type) {
			case dfl.Concat:
				uriSuffix = concat.Suffix()
			}
			_, path := grw.SplitUri(uriSuffix)
			name, format, compression := railgun.SplitNameFormatCompression(filepath.Base(path))
			if str := gtg.TryGetString(m, "name", ""); len(str) > 0 {
				name = str
			}
			if str := gtg.TryGetString(m, "format", ""); len(str) > 0 {
				format = str
			}
			if str := gtg.TryGetString(m, "compression", ""); len(str) > 0 {
				compression = str
			}
			datastore := railgun.DataStore{
				Format:      format,
				Compression: compression,
				Uri:         uri,
			}
			if arr := gtg.TryGet(m, "extent", nil); arr != nil {
				extent, err := af.ToFloat64Array.ValidateRun([]interface{}{arr})
				if err != nil {
					return nil, errors.Wrap(err, "error parsing datastore extent")
				}
				if len(extent.([]float64)) > 0 {
					datastore.Extent = extent.([]float64)
				}
			}
			if verbose {
				fmt.Println("Loading Data store:", datastore)
			}
			c = railgun.Collection{
				Name:        railgun.Slugify(name),
				Title:       name,
				Description: uri.Dfl(dfl.DefaultQuotes, false, 0),
				DataStore:   datastore,
				Cache:       railgun.NewCache(),
			}
		} else {
			_, path := grw.SplitUri(ds)
			basename, format, compression := railgun.SplitNameFormatCompression(filepath.Base(path))
			datastore := railgun.DataStore{
				Format:      format,
				Compression: compression,
				Uri:         &dfl.Literal{Value: ds},
			}
			if verbose {
				fmt.Println("Loading Data store:", datastore)
			}
			c = railgun.Collection{
				Name:        railgun.Slugify(basename),
				Title:       basename,
				Description: ds,
				DataStore:   datastore,
				Cache:       railgun.NewCache(),
			}
		}
		collectionsByName[c.Name] = c
		collectionsList = append(collectionsList, c)
	}

	r := mux.NewRouter()

	errorsChannel := make(chan error, 10000)
	requests := make(chan railgun.Request, 10000)
	messages := make(chan interface{}, 10000)

	if logFormat == "text" {
		go func(requests chan railgun.Request, logRequestsTile bool, logRequestsCache bool) {
			for r := range requests {
				switch r.(type) {
				case *railgun.TileRequest:
					if logRequestsTile {
						messages <- r.String()
					}
				case *railgun.CacheRequest:
					if logRequestsCache {
						messages <- r.String()
					}
				}
			}
		}(requests, v.GetBool("log-requests-tile"), v.GetBool("log-requests-cache"))
	} else {
		go func(requests chan railgun.Request, format string, errorsChannel chan error, logRequestsTile bool, logRequestsCache bool) {
			for r := range requests {
				switch r.(type) {
				case *railgun.TileRequest:
					if logRequestsTile {
						msg, err := r.Serialize(format)
						if err != nil {
							errorsChannel <- err
						} else {
							messages <- msg
						}
					}
				case *railgun.CacheRequest:
					if logRequestsCache {
						msg, err := r.Serialize(format)
						if err != nil {
							errorsChannel <- err
						} else {
							messages <- msg
						}
					}
				}
			}
		}(requests, logFormat, errorsChannel, v.GetBool("log-requests-tile"), v.GetBool("log-requests-cache"))
	}

	go func(messages chan interface{}) {
		for message := range messages {
			logWriter.WriteString(fmt.Sprint(message) + "\n")
			logWriter.Flush()
		}
	}(messages)

	errorDestination := v.GetString("error-destination")
	logDestination := v.GetString("log-destination")

	if errorDestination == logDestination {
		go func(errorsChannel chan error) {
			for err := range errorsChannel {
				messages <- err.Error()
			}
		}(errorsChannel)
	} else {
		go func(errorsChannel chan error) {
			for err := range errorsChannel {
				switch rerr := err.(type) {
				case *railgunerrors.ErrInvalidParameter:
					errorWriter.WriteString(rerr.Error())
				case *railgunerrors.ErrMissingCollection:
					errorWriter.WriteString(rerr.Error())
				default:
					errorWriter.WriteString(rerr.Error())
				}
			}
		}(errorsChannel)
	}

	awsSessionCache := gocache.New(5*time.Minute, 10*time.Minute)
	dflFuncs := dfl.NewFuntionMapWithDefaults()

	r.Methods("Get").Name("home").Path("/").Handler(&handlers.HomeHandler{
		CollectionsList:   collectionsList,
		CollectionsByName: collectionsByName,
		BaseHandler: &handlers.BaseHandler{
			Viper:    v,
			Requests: requests,
			Messages: messages,
			Errors:   errorsChannel,
		},
	})

	r.Methods("Get").Name("swagger").Path("/swagger.{ext}").Handler(&handlers.SwaggerHandler{&handlers.BaseHandler{
		Viper:    v,
		Requests: requests,
		Messages: messages,
		Errors:   errorsChannel,
	}})

	r.Methods("Get").Name("formats").Path("/gss/formats.{ext}").Handler(&handlers.FormatsHandler{&handlers.BaseHandler{
		Viper:    v,
		Requests: requests,
		Messages: messages,
		Errors:   errorsChannel,
	}})

	r.Methods("Get").Name("functions").Path("/dfl/functions.{ext}").Handler(&handlers.FunctionsHandler{&handlers.BaseHandler{
		Viper:    v,
		Requests: requests,
		Messages: messages,
		Errors:   errorsChannel,
	}})

	r.Methods("Get").Name("collections").Path("/collections.{ext}").Handler(&handlers.CollectionsHandler{
		CollectionsList:   collectionsList,
		CollectionsByName: collectionsByName,
		BaseHandler: &handlers.BaseHandler{
			Viper:    v,
			Requests: requests,
			Messages: messages,
			Errors:   errorsChannel,
		},
	})

	r.Methods("Get").Name("items").Path("/collections/{name}/items.{ext}").Handler(&handlers.ItemsHandler{
		CollectionsList:   collectionsList,
		CollectionsByName: collectionsByName,
		AwsSessionCache:   awsSessionCache,
		DflFuncs:          dflFuncs,
		BaseHandler: &handlers.BaseHandler{
			Viper:    v,
			Requests: requests,
			Messages: messages,
			Errors:   errorsChannel,
		},
	})

	r.Methods("Get").Name("tile").Path("/collections/{name}/data/tiles/{z}/{x}/{y}.{ext}").Handler(&handlers.TileHandler{
		CollectionsList:   collectionsList,
		CollectionsByName: collectionsByName,
		AwsSessionCache:   awsSessionCache,
		DflFuncs:          dflFuncs,
		BaseHandler: &handlers.BaseHandler{
			Viper:    v,
			Requests: requests,
			Messages: messages,
			Errors:   errorsChannel,
		},
	})

	r.Methods("Get").Name("mask").Path("/collections/{name}/mask/tiles/{z}/{x}/{y}.{ext}").Handler(&handlers.MaskHandler{
		CollectionsList:   collectionsList,
		CollectionsByName: collectionsByName,
		AwsSessionCache:   awsSessionCache,
		DflFuncs:          dflFuncs,
		BaseHandler: &handlers.BaseHandler{
			Viper:    v,
			Requests: requests,
			Messages: messages,
			Errors:   errorsChannel,
		},
	})

	return r, nil
}

func NewServer(address string, timeoutIdle time.Duration, timeoutRead time.Duration, timeoutWrite time.Duration, handler http.Handler) *http.Server {

	return &http.Server{
		Addr:         address,
		IdleTimeout:  timeoutIdle,
		ReadTimeout:  timeoutRead,
		WriteTimeout: timeoutWrite,
		Handler:      handler,
	}
}

func serveFunction(cmd *cobra.Command, args []string) {

	v := serveViper

	v.BindPFlags(cmd.PersistentFlags())
	v.BindPFlags(cmd.Flags())
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	railgun.MergeConfigs(v, v.GetStringArray("config-uri"))

	verbose := v.GetBool("verbose")

	if verbose {
		fmt.Println("=================================================")
		fmt.Println("Configuration:")
		fmt.Println("-------------------------------------------------")
		str, err := gss.SerializeString(v.AllSettings(), "properties", []string{}, -1)
		if err != nil {
			fmt.Println("error getting all settings")
			os.Exit(1)
		}
		fmt.Println(str)
		fmt.Println("=================================================")
	}

	// HTTP Flags
	address := v.GetString("http-address")
	httpTimeoutIdle := v.GetDuration("http-timeout-idle")
	httpTimeoutRead := v.GetDuration("http-timeout-read")
	httpTimeoutWrite := v.GetDuration("http-timeout-write")

	// Error Flags
	errorDestination := v.GetString("error-destination")
	errorCompression := v.GetString("error-compression")

	// Logging Flags
	logDestination := v.GetString("log-destination")
	logCompression := v.GetString("log-compression")
	logFormat := v.GetString("log-format")
	//logRequestsTile := v.GetBool("log-requests-tile")
	//logRequestsCache := v.GetBool("log-requests-cache")

	// AWS Flags
	awsDefaultRegion := v.GetString("aws-default-region")
	awsAccessKeyId := v.GetString("aws-access-key-id")
	awsSecretAccessKey := v.GetString("aws-secret-access-key")
	awsSessionToken := v.GetString("aws-session-token")

	// use StringArray since we don't want to split on comma
	datastores := v.GetStringArray("datastore")
	wait := v.GetDuration("wait")

	var aws_session *session.Session
	var s3_client *s3.S3

	if strings.HasPrefix(errorDestination, "s3://") || strings.HasPrefix(logDestination, "s3://") {
		if verbose {
			fmt.Println("Connecting to AWS with AWS_ACCESS_KEY_ID " + awsAccessKeyId)
		}
		aws_session = railgun.ConnectToAWS(awsAccessKeyId, awsSecretAccessKey, awsSessionToken, awsDefaultRegion)
		s3_client = s3.New(aws_session)
	}

	funcs := dfl.NewFuntionMapWithDefaults()

	errorWriter, err := grw.WriteToResource(errorDestination, errorCompression, true, s3_client)
	if err != nil {
		fmt.Println(errors.Wrap(err, "error creating error writer"))
		os.Exit(1)
	}

	logWriter, err := grw.WriteToResource(logDestination, logCompression, true, s3_client)
	if err != nil {
		errorWriter.WriteString(errors.Wrap(err, "error creating log writer").Error())
		errorWriter.Close()
		os.Exit(1)
	}

	router, err := NewRouter(v, funcs, datastores, errorWriter, logWriter, logFormat, verbose)
	if err != nil {
		errorWriter.WriteString(errors.Wrap(err, "error creating new router").Error())
		errorWriter.Close()
		os.Exit(1)
	}

	corsOrigin := v.GetString("cors-origin")
	corsCredentials := v.GetString("cors-credentials")

	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
			w.Header().Set("Access-Control-Allow-Credentials", corsCredentials)
			h.ServeHTTP(w, r)
		})
	})

	handler := gorilla_handlers.CompressHandler(router)

	srv := NewServer(
		address,
		httpTimeoutIdle,
		httpTimeoutRead,
		httpTimeoutWrite,
		handler)

	go func() {
		if verbose {
			fmt.Println("starting up server...")
			fmt.Println("listening on " + srv.Addr)
			fmt.Println("Using datastores:\n" + strings.Join(datastores, "\n"))
		}
		if err := srv.ListenAndServe(); err != nil {
			fmt.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-c
	errorWriter.Close()
	logWriter.Close()
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	if verbose {
		fmt.Println("received signal to attemping graceful shutdown of server")
	}
	os.Exit(0)
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "run railgun server",
	Long:  `run railgun server`,
	Run:   serveFunction,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringArrayP("datastore", "d", []string{}, "the input datastores")
	serveCmd.Flags().DurationP("duration", "", time.Second*15, "the duration to wait for graceful shutdown")

	// HTTP Flags
	serveCmd.Flags().StringSliceP("http-schemes", "", []string{"http"}, "the \"public\" schemes")
	serveCmd.Flags().StringP("http-location", "", "http://localhost/", "the \"public\" location")
	serveCmd.Flags().StringP("http-address", "a", ":8080", "http bind address")
	serveCmd.Flags().DurationP("http-timeout-idle", "", time.Second*60, "the idle timeout for the http server")
	serveCmd.Flags().DurationP("http-timeout-read", "", time.Second*15, "the read timeout for the http server")
	serveCmd.Flags().DurationP("http-timeout-write", "", time.Second*15, "the write timeout for the http server")

	// Cache Flags
	serveCmd.Flags().DurationP("cache-default-expiration", "", time.Minute*5, "the default exipration for items in the cache")
	serveCmd.Flags().DurationP("cache-cleanup-interval", "", time.Minute*10, "the cleanup interval for the cache")

	// Input Flags
	serveCmd.Flags().StringP("input-passphrase", "", "", "input passphrase for AES-256 encryption")
	serveCmd.Flags().StringP("input-salt", "", "", "input salt for AES-256 encryption")
	serveCmd.Flags().IntP("input-reader-buffer-size", "", 4096, "the buffer size for the input reader")

	// Logging Flags
	serveCmd.PersistentFlags().BoolP("log-requests-tile", "", false, "log tile requests")
	serveCmd.PersistentFlags().BoolP("log-requests-cache", "", false, "log cache hit/miss")

	// Mask Flags
	serveCmd.PersistentFlags().IntP("mask-max-zoom", "", 18, "maximum mask zoom level")
	serveCmd.PersistentFlags().IntP("mask-min-zoom", "", 14, "minimum mask zoom leel")

	// Swagger Flags
	serveCmd.PersistentFlags().StringP("swagger-contact-email", "", "", "contact email for swapper document")

	// CORS Flags
	serveCmd.PersistentFlags().StringP("cors-origin", "", "*", "value for Access-Control-Allow-Origin header")
	serveCmd.PersistentFlags().StringP("cors-credentials", "", "false", "value for Access-Control-Allow-Credentials header")

}
