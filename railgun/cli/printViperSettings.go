package cli

import (
	"fmt"
	"os"
)

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/viper"
)

import (
	"github.com/spatialcurrent/go-simple-serializer/gss"
)

func printViperSettings(v *viper.Viper) {
	fmt.Println("=================================================")
	fmt.Println("Viper Settings:")
	fmt.Println("-------------------------------------------------")
	str, err := gss.SerializeString(&gss.SerializeInput{
		Object: v.AllSettings(),
		Format: "properties",
		Header: gss.NoHeader,
		Limit:  gss.NoLimit,
		Pretty: false,
	})
	if err != nil {
		fmt.Println(errors.Wrap(err, "error serializing viper settings").Error())
		os.Exit(1)
	}
	fmt.Println(str)
	fmt.Println("=================================================")
}
