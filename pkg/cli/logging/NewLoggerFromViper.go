package logging

import (
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
	"github.com/spatialcurrent/viper"
)

// NewLoggerFromViper returns a new logger from the viper configuration.
// Panics if the logging configuration is invalid.
func NewLoggerFromViper(v *viper.Viper) *gsl.Logger {
	return gsl.CreateApplicationLogger(&gsl.CreateApplicationLoggerInput{
		ErrorDestination: v.GetString(FlagErrorDestination),
		ErrorCompression: v.GetString(FlagErrorCompression),
		ErrorFormat:      v.GetString(FlagErrorFormat),
		InfoDestination:  v.GetString(FlagInfoDestination),
		InfoCompression:  v.GetString(FlagInfoCompression),
		InfoFormat:       v.GetString(FlagInfoFormat),
		Verbose:          v.GetBool(FlagVerbose),
	})
}
