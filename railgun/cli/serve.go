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
	gocache "github.com/patrickmn/go-cache"
	"github.com/spatialcurrent/viper"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	gorilla_handlers "github.com/gorilla/handlers"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun/catalog"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/request"
	"github.com/spatialcurrent/railgun/railgun/router"
	"github.com/spatialcurrent/railgun/railgun/util"
)

var emptyFeatureCollection = []byte("{\"type\":\"FeatureCollection\",\"features\":[]}")

func NewRouter(v *viper.Viper, railgunCatalog *catalog.RailgunCatalog, errorWriter grw.ByteWriteCloser, logWriter grw.ByteWriteCloser, logFormat string, verbose bool) (*router.RailgunRouter, error) {

	errorsChannel := make(chan error, 10000)
	requests := make(chan request.Request, 10000)
	messages := make(chan interface{}, 10000)

	if logFormat == "text" {
		go func(requests chan request.Request, logRequestsTile bool, logRequestsCache bool) {
			for r := range requests {
				switch r.(type) {
				case *request.TileRequest:
					if logRequestsTile {
						messages <- r.String()
					}
				case *request.CacheRequest:
					if logRequestsCache {
						messages <- r.String()
					}
				}
			}
		}(requests, v.GetBool("log-requests-tile"), v.GetBool("log-requests-cache"))
	} else {
		go func(requests chan request.Request, format string, errorsChannel chan error, logRequestsTile bool, logRequestsCache bool) {
			for r := range requests {
				switch r.(type) {
				case *request.TileRequest:
					if logRequestsTile {
						msg, err := r.Serialize(format)
						if err != nil {
							errorsChannel <- err
						} else {
							messages <- msg
						}
					}
				case *request.CacheRequest:
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
				case *rerrors.ErrInvalidParameter:
					errorWriter.WriteString(rerr.Error())
				case *rerrors.ErrMissing:
					errorWriter.WriteString(rerr.Error())
				default:
					errorWriter.WriteString(rerr.Error())
				}
			}
		}(errorsChannel)
	}

	awsSessionCache := gocache.New(5*time.Minute, 10*time.Minute)

	r := router.NewRailgunRouter(
		v,
		railgunCatalog,
		requests,
		messages,
		errorsChannel,
		awsSessionCache)

	return r, nil
}

func serveFunction(cmd *cobra.Command, args []string) {

	v := viper.New()
	v.BindPFlags(cmd.PersistentFlags())
	v.BindPFlags(cmd.Flags())
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	util.MergeConfigs(v, v.GetStringArray("config-uri"))

	verbose := v.GetBool("verbose")

	if verbose {
		str, err := gss.SerializeString(v.AllSettings(), "properties", gss.NoHeader, gss.NoLimit)
		if err != nil {
			fmt.Println(errors.Wrap(err, "error getting all settings for config"))
			os.Exit(1)
		}
		fmt.Println("=================================================")
		fmt.Println("Configuration:")
		fmt.Println("-------------------------------------------------")
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
	wait := v.GetDuration("wait")

	var aws_session *session.Session
	var s3_client *s3.S3

	if strings.HasPrefix(errorDestination, "s3://") || strings.HasPrefix(logDestination, "s3://") {
		if verbose {
			fmt.Println("Connecting to AWS with AWS_ACCESS_KEY_ID " + awsAccessKeyId)
		}
		aws_session = util.ConnectToAWS(awsAccessKeyId, awsSecretAccessKey, awsSessionToken, awsDefaultRegion)
		s3_client = s3.New(aws_session)
	}

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

	railgunCatalog := catalog.NewRailgunCatalog()

	err = railgunCatalog.LoadFromViper(v)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if uri := v.GetString("catalog-uri"); len(uri) > 0 {
		err := railgunCatalog.LoadFromFile(uri, logWriter, errorWriter)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	logWriter.Flush()
	errorWriter.Flush()

	router, err := NewRouter(v, railgunCatalog, errorWriter, logWriter, logFormat, verbose)
	if err != nil {
		errorWriter.WriteString(errors.Wrap(err, "error creating new router").Error())
		errorWriter.Close()
		os.Exit(1)
	}

	handler := gorilla_handlers.CompressHandler(router)

	srv := &http.Server{
		Addr:         address,
		IdleTimeout:  httpTimeoutIdle,
		ReadTimeout:  httpTimeoutRead,
		WriteTimeout: httpTimeoutWrite,
		Handler:      handler,
	}

	go func() {
		if verbose {
			fmt.Println("starting up server...")
			fmt.Println("listening on " + srv.Addr)
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

	serveCmd.Flags().StringArrayP("datastore", "d", []string{}, "datastore")
	serveCmd.Flags().StringArrayP("workspace", "w", []string{}, "workspace")
	serveCmd.Flags().StringArrayP("layer", "l", []string{}, "layer")
	serveCmd.Flags().StringArrayP("process", "p", []string{}, "process")
	serveCmd.Flags().StringArrayP("service", "s", []string{}, "service")
	serveCmd.Flags().StringArrayP("job", "j", []string{}, "job")
	serveCmd.Flags().DurationP("duration", "", time.Second*15, "the duration to wait for graceful shutdown")

	// HTTP Flags
	serveCmd.Flags().StringSliceP("http-schemes", "", []string{"http"}, "the \"public\" schemes")
	serveCmd.Flags().StringP("http-location", "", "http://localhost:8080/", "the \"public\" location")
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
	serveCmd.Flags().BoolP("log-requests-tile", "", false, "log tile requests")
	serveCmd.Flags().BoolP("log-requests-cache", "", false, "log cache hit/miss")

	// Mask Flags
	serveCmd.Flags().IntP("mask-max-zoom", "", 18, "maximum mask zoom level")
	serveCmd.Flags().IntP("mask-min-zoom", "", 14, "minimum mask zoom leel")

	// Swagger Flags
	serveCmd.Flags().StringP("swagger-contact-name", "", "", "contact name for swapper document")
	serveCmd.Flags().StringP("swagger-contact-email", "", "", "contact email for swapper document")
	serveCmd.Flags().StringP("swagger-contact-url", "", "", "contact url for swapper document")

	// CORS Flags
	serveCmd.Flags().StringP("cors-origin", "", "*", "value for Access-Control-Allow-Origin header")
	serveCmd.Flags().StringP("cors-credentials", "", "false", "value for Access-Control-Allow-Credentials header")

	// Catalog Skip Errors
	serveCmd.Flags().String("catalog-uri", "", "uri of the catalog backend")
	serveCmd.Flags().BoolP("config-skip-errors", "", false, "skip loading config with bad errors")

}