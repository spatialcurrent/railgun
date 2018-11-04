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
	//"github.com/spf13/viper"
	//"image"
	//"image/color"
	//"image/draw"
	//"image/png"
	"net/http"
	"os"
	"os/signal"
	//"path/filepath"
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
	//"github.com/spatialcurrent/go-adaptive-functions/af"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	//"github.com/spatialcurrent/go-simple-serializer/gss"
	//"github.com/spatialcurrent/go-try-get/gtg"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spatialcurrent/railgun/railgun/handlers"
	//"github.com/spatialcurrent/railgun/railgun/img"
	"github.com/spatialcurrent/railgun/railgun/railgunerrors"
)

var emptyFeatureCollection = []byte("{\"type\":\"FeatureCollection\",\"features\":[]}")

func NewRouter(c *railgun.Config, funcs dfl.FunctionMap, datastores []string, errorWriter grw.ByteWriteCloser, logWriter grw.ByteWriteCloser, logFormat string, verbose bool) (*mux.Router, error) {

	r := mux.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Method:", r.Method)
			next.ServeHTTP(w, r)
		})
	})

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
		}(requests, c.GetBool("log-requests-tile"), c.GetBool("log-requests-cache"))
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
		}(requests, logFormat, errorsChannel, c.GetBool("log-requests-tile"), c.GetBool("log-requests-cache"))
	}

	go func(messages chan interface{}) {
		for message := range messages {
			logWriter.WriteString(fmt.Sprint(message) + "\n")
			logWriter.Flush()
		}
	}(messages)

	errorDestination := c.GetString("error-destination")
	logDestination := c.GetString("log-destination")

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
				case *railgunerrors.ErrMissing:
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
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("Get").Name("swagger").Path("/swagger.{ext}").Handler(&handlers.SwaggerHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("Get").Name("formats").Path("/gss/formats.{ext}").Handler(&handlers.FormatsHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("Get").Name("functions").Path("/dfl/functions.{ext}").Handler(&handlers.FunctionsHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("GET", "POST", "PUT", "OPTIONS").Name("workspaces").Path("/workspaces.{ext}").Handler(&handlers.WorkspacesHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("GET", "DELETE").Name("workspace").Path("/workspaces/{name}.{ext}").Handler(&handlers.WorkspaceHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("GET", "POST", "PUT", "OPTIONS").Name("datastores").Path("/datastores.{ext}").Handler(&handlers.DataStoresHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("GET", "DELETE").Name("datastore").Path("/datastores/{name}.{ext}").Handler(&handlers.DataStoreHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("Get").Name("layers").Path("/layers.{ext}").Handler(&handlers.LayersHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("GET", "POST", "PUT", "OPTIONS").Name("processes").Path("/processes.{ext}").Handler(&handlers.ProcessesHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("GET", "DELETE").Name("process").Path("/processes/{name}.{ext}").Handler(&handlers.ProcessHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("GET", "POST", "PUT", "OPTIONS").Name("services").Path("/services.{ext}").Handler(&handlers.ServicesHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("GET", "DELETE").Name("service").Path("/services/{name}.{ext}").Handler(&handlers.ServiceHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("POST", "PUT", "OPTIONS").Name("services").Path("/services/exec.{ext}").Handler(&handlers.ServicesExecHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("GET", "POST", "PUT", "OPTIONS").Name("jobs").Path("/jobs.{ext}").Handler(&handlers.JobsHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("GET", "DELETE").Name("service").Path("/jobs/{name}.{ext}").Handler(&handlers.JobHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("Get").Name("items").Path("/layers/{name}/items.{ext}").Handler(&handlers.ItemsHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("Get").Name("tile").Path("/layers/{name}/data/tiles/{z}/{x}/{y}.{ext}").Handler(&handlers.TileHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
		},
	})

	r.Methods("Get").Name("mask").Path("/layers/{name}/mask/tiles/{z}/{x}/{y}.{ext}").Handler(&handlers.MaskHandler{
		BaseHandler: &handlers.BaseHandler{
			Config:          c,
			Requests:        requests,
			Messages:        messages,
			Errors:          errorsChannel,
			AwsSessionCache: awsSessionCache,
			DflFuncs:        dflFuncs,
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

	config := railgun.NewConfig(cmd)
	err := config.Reload()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	verbose := config.GetBool("verbose")

	if verbose {
		err := config.Print()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// HTTP Flags
	address := config.GetString("http-address")
	httpTimeoutIdle := config.GetDuration("http-timeout-idle")
	httpTimeoutRead := config.GetDuration("http-timeout-read")
	httpTimeoutWrite := config.GetDuration("http-timeout-write")

	// Error Flags
	errorDestination := config.GetString("error-destination")
	errorCompression := config.GetString("error-compression")

	// Logging Flags
	logDestination := config.GetString("log-destination")
	logCompression := config.GetString("log-compression")
	logFormat := config.GetString("log-format")
	//logRequestsTile := v.GetBool("log-requests-tile")
	//logRequestsCache := v.GetBool("log-requests-cache")

	// AWS Flags
	awsDefaultRegion := config.GetString("aws-default-region")
	awsAccessKeyId := config.GetString("aws-access-key-id")
	awsSecretAccessKey := config.GetString("aws-secret-access-key")
	awsSessionToken := config.GetString("aws-session-token")

	// use StringArray since we don't want to split on comma
	datastores := config.GetStringArray("datastore")
	wait := config.GetDuration("wait")

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

	router, err := NewRouter(config, funcs, datastores, errorWriter, logWriter, logFormat, verbose)
	if err != nil {
		errorWriter.WriteString(errors.Wrap(err, "error creating new router").Error())
		errorWriter.Close()
		os.Exit(1)
	}

	corsOrigin := config.GetString("cors-origin")
	corsCredentials := config.GetString("cors-credentials")

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

	serveCmd.Flags().StringArrayP("datastore", "d", []string{}, "datastore")
	serveCmd.Flags().StringArrayP("workspace", "w", []string{}, "workspace")
	serveCmd.Flags().StringArrayP("layer", "l", []string{}, "layer")
	serveCmd.Flags().StringArrayP("process", "p", []string{}, "process")
	serveCmd.Flags().StringArrayP("service", "s", []string{}, "service")
	serveCmd.Flags().StringArrayP("jobs", "j", []string{}, "jobs")
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
