// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"os"
	"strings"

	"github.com/spatialcurrent/cobra"

	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/railgun/pkg/cli/algorithms"
	"github.com/spatialcurrent/railgun/pkg/cli/catalog"
	"github.com/spatialcurrent/railgun/pkg/cli/client"
	"github.com/spatialcurrent/railgun/pkg/cli/formats"
	"github.com/spatialcurrent/railgun/pkg/cli/process"
	"github.com/spatialcurrent/railgun/pkg/cli/serve"
	"github.com/spatialcurrent/railgun/pkg/cli/version"
)

// Execute handles command line calls to railgun.
func Execute(gitBranch string, gitCommit string) error {

	//
	// Root Command
	//

	var rootCmd = &cobra.Command{
		Use:   "railgun",
		Short: "a simple and fast data processing tool",
		Long: `Railgun is a simple and fast data processing tool.
Through go-reader-writer, supports the follow compression algorithms: ` + strings.Join(grw.Algorithms, ", ") + `
Through go-simple-serializer, supports the follow file formats: ` + strings.Join(gss.Formats, ", "),
	}
	InitRootFlags(rootCmd.PersistentFlags())

	//
	// Completion Command
	//

	completionCommandLong := ""
	if _, err := os.Stat("/etc/bash_completion.d/"); !os.IsNotExist(err) {
		completionCommandLong = "To install completion scripts run:\nrailgun completion > /etc/bash_completion.d/railgun"
	} else {
		if _, err := os.Stat("/usr/local/etc/bash_completion.d/"); !os.IsNotExist(err) {
			completionCommandLong = "To install completion scripts run:\nrailgun completion > /usr/local/etc/bash_completion.d/railgun"
		} else {
			completionCommandLong = "To install completion scripts run:\nrailgun completion > .../bash_completion.d/railgun"
		}
	}

	rootCmd.AddCommand(func() *cobra.Command {
		return &cobra.Command{
			Use:   "completion",
			Short: "Generates bash completion scripts",
			Long:  completionCommandLong,
			RunE: func(cmd *cobra.Command, args []string) error {
				return rootCmd.GenBashCompletion(os.Stdout)
			},
		}
	}())

	rootCmd.AddCommand(version.NewCommand(&version.NewCommandInput{
		GitBranch: gitBranch,
		GitCommit: gitCommit,
	}))

	//
	// Process Command
	//

	rootCmd.AddCommand(process.NewCommand())

	//
	// Serve Command
	//

	rootCmd.AddCommand(serve.NewCommand(&serve.NewCommandInput{
		GitBranch: gitBranch,
		GitCommit: gitCommit,
	}))

	//
	// Catalog Command
	//

	rootCmd.AddCommand(catalog.NewCommand())

	//
	// Client Command
	//

	rootCmd.AddCommand(client.NewCommand())

	//
	// Algorithms Command
	//

	rootCmd.AddCommand(algorithms.NewCommand())

	//
	// Formats Command
	//

	rootCmd.AddCommand(formats.NewCommand())

	return rootCmd.Execute()
}
