package serve

import (
	"crypto/rsa"
	"time"

	gocache "github.com/patrickmn/go-cache"

	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
	"github.com/spatialcurrent/railgun/pkg/catalog"
	"github.com/spatialcurrent/railgun/pkg/request"
	"github.com/spatialcurrent/railgun/pkg/router"
	"github.com/spatialcurrent/viper"
)

type NewRouterInput struct {
	Viper          *viper.Viper
	RailgunCatalog *catalog.RailgunCatalog
	Logger         *gsl.Logger
	PublicKey      *rsa.PublicKey
	PrivateKey     *rsa.PrivateKey
	ValidMethods   []string
	ErrorsChannel  chan interface{}
	Requests       chan request.Request
	Messages       chan interface{}
	GitBranch      string
	GitCommit      string
	Verbose        bool
}

func NewRouter(input *NewRouterInput) (*router.RailgunRouter, error) {

	go func(requests chan request.Request, logRequestsTile bool, logRequestsCache bool) {
		for r := range requests {
			switch r.(type) {
			case *request.TileRequest:
				if logRequestsTile {
					input.Messages <- r
				}
			case *request.CacheRequest:
				if logRequestsCache {
					input.Messages <- r
				}
			}
		}
	}(
		input.Requests,
		input.Viper.GetBool("log-requests-tile"),
		input.Viper.GetBool("log-requests-cache"),
	)

	errorDestination := input.Viper.GetString("error-destination")
	infoDestination := input.Viper.GetString("info-destination")

	if errorDestination == infoDestination {
		go func(errorsChannel chan interface{}) {
			for err := range errorsChannel {
				input.Messages <- err
			}
		}(input.ErrorsChannel)
	} else {
		input.Logger.ListenError(input.ErrorsChannel, nil)
	}

	awsSessionCache := gocache.New(5*time.Minute, 10*time.Minute)

	r := router.NewRailgunRouter(&router.NewRailgunRouterInput{
		Viper:           input.Viper,
		RailgunCatalog:  input.RailgunCatalog,
		Requests:        input.Requests,
		Messages:        input.Messages,
		ErrorsChannel:   input.ErrorsChannel,
		AwsSessionCache: awsSessionCache,
		PublicKey:       input.PublicKey,
		PrivateKey:      input.PrivateKey,
		ValidMethods:    input.ValidMethods,
		GitBranch:       input.GitBranch,
		GitCommit:       input.GitCommit,
		Logger:          input.Logger,
	})

	return r, nil
}
