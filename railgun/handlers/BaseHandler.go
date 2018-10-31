package handlers

import (
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spf13/viper"
)

type BaseHandler struct {
	Viper    *viper.Viper
	Requests chan railgun.Request
	Messages chan interface{}
	Errors   chan error
}
