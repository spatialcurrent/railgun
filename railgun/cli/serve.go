// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"context"
	"crypto/rsa"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	gocache "github.com/patrickmn/go-cache"
	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/viper"
	//"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-sync-logger/gsl"
)

import (
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spatialcurrent/railgun/railgun/catalog"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/request"
	"github.com/spatialcurrent/railgun/railgun/router"
	"github.com/spatialcurrent/railgun/railgun/util"
)

var emptyFeatureCollection = []byte("{\"type\":\"FeatureCollection\",\"features\":[]}")

func NewRouter(v *viper.Viper, railgunCatalog *catalog.RailgunCatalog, logger *gsl.Logger, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey, validMethods []string, errorsChannel chan interface{}, requests chan request.Request, messages chan interface{}, version string, gitBranch string, gitCommit string, verbose bool) (*router.RailgunRouter, error) {

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

	r := router.NewRailgunRouter(
		v,
		railgunCatalog,
		requests,
		messages,
		errorsChannel,
		awsSessionCache,
		publicKey,
		privateKey,
		validMethods,
		version,
		gitBranch,
		gitCommit)

	return r, nil
}

func initPublicKey(publicKeyString string, publicKeyUri string, s3_client *s3.S3) (*rsa.PublicKey, error) {

	if len(publicKeyString) > 0 {
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyString))
		if err != nil {
			return nil, errors.Wrap(err, "error parsing RSA public key from jwt-public-key")
		}
		return publicKey, nil
	}

	publicKeyReader, _, err := grw.ReadFromResource(publicKeyUri, "", 4096, false, s3_client)
	if err != nil {
		return nil, errors.Wrap(err, "error opening public key at uri "+publicKeyUri)
	}

	publicKeyBytes, err := publicKeyReader.ReadAllAndClose()
	if err != nil {
		return nil, errors.Wrap(err, "error reading public key at uri "+publicKeyUri)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing RSA public key from "+publicKeyUri)
	}

	return publicKey, nil
}

func initPrivateKey(privateKeyString string, privateKeyUri string, s3_client *s3.S3) (*rsa.PrivateKey, error) {

	if len(privateKeyString) > 0 {
		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyString))
		if err != nil {
			return nil, errors.Wrap(err, "error parsing RSA public key from jwt-private-key")
		}
		return privateKey, nil
	}

	privateKeyReader, _, err := grw.ReadFromResource(privateKeyUri, "", 4096, false, s3_client)
	if err != nil {
		return nil, errors.Wrap(err, "error opening private key at uri "+privateKeyUri)
	}

	privateKeyBytes, err := privateKeyReader.ReadAllAndClose()
	if err != nil {
		return nil, errors.Wrap(err, "error reading private key at uri "+privateKeyUri)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing RSA private key from "+privateKeyUri)
	}

	return privateKey, nil
}

func serveFunction(cmd *cobra.Command, args []string) {

	v := viper.New()
	err := v.BindPFlags(cmd.PersistentFlags())
	if err != nil {
		panic(err)
	}
	err = v.BindPFlags(cmd.Flags())
	if err != nil {
		panic(err)
	}
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

	// Logging Flags
	infoDestination := v.GetString("info-destination")
	infoCompression := v.GetString("info-compression")
	infoFormat := v.GetString("info-format")
	errorDestination := v.GetString("error-destination")
	errorCompression := v.GetString("error-compression")
	errorFormat := v.GetString("error-format")

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
	publicKeyString := v.GetString("jwt-public-key")
	publicKeyUri := v.GetString("jwt-public-key-uri")
	privateKeyString := v.GetString("jwt-private-key")
	privateKeyUri := v.GetString("jwt-private-key-uri")

	var s3_client *s3.S3

	if strings.HasPrefix(errorDestination, "s3://") || strings.HasPrefix(infoDestination, "s3://") || strings.HasPrefix(catalogUri, "s3://") || strings.HasPrefix(publicKeyUri, "s3://") || strings.HasPrefix(privateKeyUri, "s3://") {
		aws_session, err := util.ConnectToAWS(awsAccessKeyId, awsSecretAccessKey, awsSessionToken, awsDefaultRegion)
		if err != nil {
			fmt.Println(errors.Wrap(err, "error connecting to AWS"))
			os.Exit(1)
		}
		s3_client = s3.New(aws_session)
	}

	errorWriter, err := grw.WriteToResource(errorDestination, errorCompression, true, s3_client)
	if err != nil {
		fmt.Println(errors.Wrap(err, "error creating error writer"))
		os.Exit(1)
	}

	infoWriter, err := grw.WriteToResource(infoDestination, infoCompression, true, s3_client)
	if err != nil {
		errorWriter.WriteError(errors.Wrap(err, "error creating log writer")) // #nosec
		errorWriter.Close()                                                   // #nosec
		os.Exit(1)
	}

	logger := gsl.NewLogger(
		map[string]int{"info": 0, "error": 1},
		[]grw.ByteWriteCloser{infoWriter, errorWriter},
		[]string{infoFormat, errorFormat},
	)

	railgunCatalog := catalog.NewRailgunCatalog()

	err = railgunCatalog.LoadFromViper(v)
	if err != nil {
		logger.Fatal(err)
	}

	messages := make(chan interface{}, 10000)
	logger.ListenInfo(messages, nil)

	if len(catalogUri) > 0 {
		err := railgunCatalog.LoadFromUri(catalogUri, logger, s3_client, messages)
		if err != nil {
			logger.Fatal(err)
		}
	}

	if len(publicKeyString) == 0 && len(publicKeyUri) == 0 {
		logger.Fatal(errors.New("jwt-public-key or jwt-public-key-uri is required"))
	}

	if len(privateKeyString) == 0 && len(privateKeyUri) == 0 {
		logger.Fatal(errors.New("jwt-private-key or jwt-private-key-uri is required"))
	}

	publicKey, err := initPublicKey(publicKeyString, publicKeyUri, s3_client)
	if err != nil {
		logger.Fatal(errors.Wrap(err, "error initializing public key"))
	}

	privateKey, err := initPrivateKey(privateKeyString, privateKeyUri, s3_client)
	if err != nil {
		logger.Fatal(errors.Wrap(err, "error initializing private key"))
	}

	validMethods := v.GetStringArray("jwt-valid-methods")
	if len(validMethods) == 0 {
		logger.Fatal(&rerrors.ErrMissingRequiredParameter{Name: "jwt-valid-methods"})
	}

	errorsChannel := make(chan interface{}, 10000)
	requests := make(chan request.Request, 10000)

	handler, err := NewRouter(
		v,
		railgunCatalog,
		logger,
		publicKey,
		privateKey,
		validMethods,
		errorsChannel,
		requests,
		messages,
		railgun.Version,
		gitBranch,
		gitCommit,
		verbose)
	if err != nil {
		logger.Fatal(errors.Wrap(err, "error creating new router"))
	}

	gracefulShutdown := v.GetBool("http-graceful-shutdown")
	gracefulShutdownWait := v.GetDuration("http-graceful-shutdown-wait")

	messages <- map[string]interface{}{"server": map[string]interface{}{
		"address":              address,
		"httpTimeoutIdle":      httpTimeoutIdle,
		"httpTimeoutRead":      httpTimeoutRead,
		"httpTimeoutWrite":     httpTimeoutWrite,
		"gracefulShutdown":     gracefulShutdown,
		"gracefulShutdownWait": gracefulShutdownWait,
	}}

	if httpTimeoutIdle.Seconds() < 5.0 {
		logger.Fatal("http-timeout-idle cannot be less than 5 seconds")
	}

	if httpTimeoutRead.Seconds() < 5.0 {
		logger.Fatal("http-timeout-read cannot be less than 5 seconds")
	}

	if httpTimeoutWrite.Seconds() < 5.0 {
		logger.Fatal("http-timeout-write cannot be less than 5 seconds")
	}

	if gracefulShutdown && gracefulShutdownWait.Seconds() < 5.0 {
		logger.Fatal("graceful-shutdown-wait cannot be less than 5 seconds")
	}

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
			messages <- map[string]interface{}{"message": "starting server with graceful shutdown"}
			messages <- map[string]interface{}{"message": "listening on " + srv.Addr}
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
		fmt.Println("received signal for graceful shutdown of server")
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	messages <- map[string]interface{}{"message": "starting server without graceful shutdown"}
	messages <- map[string]interface{}{"message": "listening on " + srv.Addr}
	logger.Fatal(srv.ListenAndServe())
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

	// HTTP Flags
	serveCmd.Flags().StringSlice("http-schemes", []string{"http"}, "the \"public\" schemes")
	serveCmd.Flags().StringP("http-location", "", "http://localhost:8080", "the \"public\" location")
	serveCmd.Flags().StringP("http-address", "a", ":8080", "http bind address")
	serveCmd.Flags().DurationP("http-timeout-idle", "", time.Second*60, "the idle timeout for the http server")
	serveCmd.Flags().DurationP("http-timeout-read", "", time.Second*15, "the read timeout for the http server")
	serveCmd.Flags().DurationP("http-timeout-write", "", time.Second*15, "the write timeout for the http server")
	serveCmd.Flags().Bool("http-middleware-debug", false, "enable debug middleware")
	serveCmd.Flags().Bool("http-middleware-recover", false, "enable recovery middleware")
	serveCmd.Flags().Bool("http-middleware-gzip", false, "enable gzip middleware")
	serveCmd.Flags().Bool("http-middleware-cors", false, "enable CORS middleware")
	serveCmd.Flags().Bool("http-graceful-shutdown", false, "enable graceful shutdown")
	serveCmd.Flags().Duration("http-graceful-shutdown-wait", time.Second*15, "the duration to wait for graceful shutdown")

	// Cache Flags
	serveCmd.Flags().DurationP("cache-default-expiration", "", time.Minute*5, "the default expiration for items in the cache")
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

	// Security
	serveCmd.Flags().String("root-password", "", "root user password")
	serveCmd.Flags().String("jwt-private-key", "", "Private RSA Key for JWT")
	serveCmd.Flags().String("jwt-private-key-uri", "", "URI to private RSA Key for JWT")
	serveCmd.Flags().String("jwt-public-key", "", "Public RSA Key for JWT")
	serveCmd.Flags().String("jwt-public-key-uri", "", "URI to public RSA Key for JWT")
	serveCmd.Flags().StringArray("jwt-valid-methods", []string{"RS512"}, "Valid methods for JWT")
	serveCmd.Flags().Duration("jwt-session-duration", 60*time.Minute, "duration of authenticated session")

	// Tile
	serveCmd.Flags().IntP("tile-random-delay", "", 1000, "random delay for processing tiles in milliseconds")

}
