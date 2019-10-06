// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package main

import (
	"fmt"
	"os"

	"github.com/spatialcurrent/railgun/pkg/cli"
)

// GitCommit & Branch are empty unless set as a build flag
// See https://blog.alexellis.io/inject-build-time-vars-golang/
var gitBranch string
var gitCommit string

func main() {
	err := cli.Execute(gitBranch, gitCommit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "railgun: %s\n", err.Error())
		os.Exit(1)
	}
}
