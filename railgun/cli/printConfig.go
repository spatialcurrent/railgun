package cli

import (
	"fmt"
	"os"
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-simple-serializer/gss"
)

func printConfig(c mapper) {
	fmt.Println("=================================================")
	fmt.Println("Configuration:")
	fmt.Println("-------------------------------------------------")
	str, err := gss.SerializeString(&gss.SerializeInput{
		Object: c.Map(),
		Format: "yaml",
		Header: gss.NoHeader,
		Limit:  gss.NoLimit,
		Pretty: false,
	})
	if err != nil {
		fmt.Println(errors.Wrap(err, "error serializing process config").Error())
		os.Exit(1)
	}
	fmt.Println(str)
	fmt.Println("=================================================")
}
