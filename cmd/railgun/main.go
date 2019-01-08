// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package main

import (
	"github.com/spatialcurrent/railgun/railgun/cli"
)

// GitCommit & Branch are empty unless set as a build flag
// See https://blog.alexellis.io/inject-build-time-vars-golang/
var gitBranch string
var gitCommit string

func main() {
	cli.Execute(gitBranch, gitCommit)
}
