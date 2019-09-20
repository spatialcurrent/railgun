// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package serve

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
	"github.com/spatialcurrent/railgun/pkg/catalog"
	"github.com/spatialcurrent/railgun/pkg/cli/input"
	"github.com/spatialcurrent/railgun/pkg/cli/logging"
	"github.com/spatialcurrent/railgun/pkg/config"
	"github.com/spatialcurrent/railgun/pkg/jwt"
	"github.com/spatialcurrent/railgun/pkg/request"
	"github.com/spatialcurrent/railgun/pkg/router"
	"github.com/spatialcurrent/railgun/pkg/util"
	"github.com/spatialcurrent/viper"
)

const (
	FlagCacheDefaultExpiration = "cache-default-expiration"
	FlagCacheCleanupInterval   = "cache-cleanup-interval"
	FlagLogRequestsCache       = "log-requests-cache"
	FlagLogRequestsTile        = "log-requests-tile"
	FlagMaskMinZoom            = "mask-min-zoom"
	FlagMaskMaxZoom            = "mask-max-zoom"
	FlagCatalogUri             = "catalog-uri"
	FlagConfigSkipErrors       = "config-skip-errors"
	FlagTileRandomDelay        = "tile-random-delay"
	FlagRootPassword           = "root-password"

	FlagCoconutApiUrl       = "coconut-api-url"
	FlagCoconutBaselayerUrl = "coconut-baselayer-url"
	FlagCoconutBundleUrl    = "coconut-bundle-url"

	FlagInputPassphrase       = input.FlagInputPassphrase
	FlagInputSalt             = input.FlagInputSalt
	FlagInputReaderBufferSize = input.FlagInputReaderBufferSize

	DefaultCacheDefaultExpiration = time.Minute * 5
	DefaultCacheCleanupInterval   = time.Minute * 10

	DefaultMaskMinZoom = 14
	DefaultMaskMaxZoom = 18

	DefaultTileRandomDelay = 1000 // in milliseconds

	DefaultCoconutApiUrl       = "http://localhost:8080/"
	DefaultCoconutBaselayerUrl = "https://{a-c}.tile.openstreetmap.org/{z}/{x}/{y}.png"
	DefaultCoconutBundleUrl    = "https://coconut.spatialcurrent.io/bundle.js"

	DefaultInputReaderBufferSize = input.DefaultInputReaderBufferSize
)

var emptyFeatureCollection = []byte("{\"type\":\"FeatureCollection\",\"features\":[]}")

func newRouter(v *viper.Viper, railgunCatalog *catalog.RailgunCatalog, logger *gsl.Logger, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey, validMethods []string, errorsChannel chan interface{}, requests chan request.Request, messages chan interface{}, gitBranch string, gitCommit string, verbose bool) (*router.RailgunRouter, error) {

	go func(requests chan request.Request, logRequestsTile bool, logRequestsCache bool) {
		for r := range requests {
			switch r.(type) {
			case *request.TileRequest:
				if logRequestsTile {
					messages <- r
				}
			case *request.CacheRequest:
				if logRequestsCache {
					messages <- r
				}
			}
		}
	}(requests, v.GetBool("log-requests-tile"), v.GetBool("log-requests-cache"))

	errorDestination := v.GetString("error-destination")
	infoDestination := v.GetString("info-destination")

	if errorDestination == infoDestination {
		go func(errorsChannel chan interface{}) {
			for err := range errorsChannel {
				messages <- err
			}
		}(errorsChannel)
	} else {
		logger.ListenError(errorsChannel, nil)
	}

	awsSessionCache := gocache.New(5*time.Minute, 10*time.Minute)

	r := router.NewRailgunRouter(&router.NewRailgunRouterInput{
		Viper:           v,
		RailgunCatalog:  railgunCatalog,
		Requests:        requests,
		Messages:        messages,
		ErrorsChannel:   errorsChannel,
		AwsSessionCache: awsSessionCache,
		PublicKey:       publicKey,
		PrivateKey:      privateKey,
		ValidMethods:    validMethods,
		GitBranch:       gitBranch,
		GitCommit:       gitCommit,
		Logger:          logger,
	})

	return r, nil
}

func serveFunction(gitBranch string, gitCommit string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {

		v := viper.New()

		err := v.BindPFlags(cmd.Flags())
		if err != nil {
			return errors.Wrap(err, "error binding flags")
		}
		v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		v.AutomaticEnv() // set environment variables to overwrite config
		util.MergeConfigs(v, v.GetStringArray("config-uri"))

		verbose := v.GetBool("verbose")

		if verbose {
			config.PrintViperSettings(v)
		}

		err = CheckServeConfig(v, args)
		if err != nil {
			return errors.Wrap(err, "error with configuration")
		}

		// Runtime Flags
		runtimeMaxProcs := v.GetInt("runtime-max-procs")
		if runtimeMaxProcs == 0 {
			runtimeMaxProcs = runtime.NumCPU()
		} else if runtimeMaxProcs < 0 {
			panic(errors.New("runtime-max-procs cannot be less than 1"))
		}
		fmt.Println(fmt.Sprintf("Maximum number of parallel procsses set to %d", runtimeMaxProcs))
		runtime.GOMAXPROCS(runtimeMaxProcs)

		// HTTP Flags
		address := v.GetString("http-address")
		httpTimeoutIdle := v.GetDuration("http-timeout-idle")
		httpTimeoutRead := v.GetDuration("http-timeout-read")
		httpTimeoutWrite := v.GetDuration("http-timeout-write")

		//logRequestsTile := v.GetBool("log-requests-tile")
		//logRequestsCache := v.GetBool("log-requests-cache")

		// AWS Flags
		awsDefaultRegion := v.GetString("aws-default-region")
		awsAccessKeyId := v.GetString("aws-access-key-id")
		awsSecretAccessKey := v.GetString("aws-secret-access-key")
		awsSessionToken := v.GetString("aws-session-token")
		//awsContainerCredentialsRelativeUri := v.GetString("aws-container-credentials-relative-uri")

		// Catalog Flags
		catalogUri := v.GetString("catalog-uri")

		// Security Flags
		publicKeyUri := v.GetString("jwt-public-key-uri")
		privateKeyUri := v.GetString("jwt-private-key-uri")

		var s3Client *s3.S3

		if strings.HasPrefix(catalogUri, "s3://") || strings.HasPrefix(publicKeyUri, "s3://") || strings.HasPrefix(privateKeyUri, "s3://") {
			aws_session, err := util.ConnectToAWS(awsAccessKeyId, awsSecretAccessKey, awsSessionToken, awsDefaultRegion)
			if err != nil {
				fmt.Println(errors.Wrap(err, "error connecting to AWS"))
				os.Exit(1)
			}
			s3Client = s3.New(aws_session)
		}

		logger := logging.NewLoggerFromViper(v)

		railgunCatalog := catalog.NewRailgunCatalog()

		err = railgunCatalog.LoadFromViper(v)
		if err != nil {
			logger.Fatal(err)
		}

		messages := make(chan interface{}, 10000)
		logger.ListenInfo(messages, nil)

		if len(catalogUri) > 0 {
			err := railgunCatalog.LoadFromUri(catalogUri, logger, s3Client, messages)
			if err != nil {
				logger.Fatal(err)
			}
		}

		publicKey, err := jwt.LoadPublicKey(v.GetString("jwt-public-key"), publicKeyUri, s3Client)
		if err != nil {
			logger.Fatal(errors.Wrap(err, "error initializing public key"))
		}

		privateKey, err := jwt.LoadPrivateKey(v.GetString("jwt-private-key"), privateKeyUri, s3Client)
		if err != nil {
			logger.Fatal(errors.Wrap(err, "error initializing private key"))
		}

		errorsChannel := make(chan interface{}, 10000)
		requests := make(chan request.Request, 10000)

		handler, err := newRouter(
			v,
			railgunCatalog,
			logger,
			publicKey,
			privateKey,
			v.GetStringSlice("jwt-valid-methods"),
			errorsChannel,
			requests,
			messages,
			gitBranch,
			gitCommit,
			verbose)
		if err != nil {
			logger.Fatal(errors.Wrap(err, "error creating new router"))
		}

		gracefulShutdown := v.GetBool("http-graceful-shutdown")
		gracefulShutdownWait := v.GetDuration("http-graceful-shutdown-wait")

		logger.Info(map[string]interface{}{
			"msg":                  "configuring server",
			"address":              address,
			"httpTimeoutIdle":      httpTimeoutIdle,
			"httpTimeoutRead":      httpTimeoutRead,
			"httpTimeoutWrite":     httpTimeoutWrite,
			"gracefulShutdown":     gracefulShutdown,
			"gracefulShutdownWait": gracefulShutdownWait,
		})

		srv := &http.Server{
			Addr:         address,
			IdleTimeout:  httpTimeoutIdle,
			ReadTimeout:  httpTimeoutRead,
			WriteTimeout: httpTimeoutWrite,
			Handler:      handler,
		}

		logger.Flush()

		if gracefulShutdown {
			go func() {
				logger.Info("starting server with graceful shutdown")
				logger.InfoF("listening on %s", srv.Addr)
				logger.Flush()
				if err := srv.ListenAndServe(); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}()

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			<-c
			logger.Close()
			ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownWait)
			defer cancel()
			err := srv.Shutdown(ctx)
			logger.Info("received signal for graceful shutdown of server")
			logger.Flush()
			if err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		}

		logger.Info("starting server without graceful shutdown")
		logger.InfoF("listening on %s", srv.Addr)
		logger.Flush()
		logger.Fatal(srv.ListenAndServe())

		return nil
	}
}
