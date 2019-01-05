// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"fmt"
	"os"
	"strings"
)

import (
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/viper"
)

import (
  "github.com/spatialcurrent/railgun/railgun"
)

var cfgFile string
var gitCommit string
var gitBranch string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "railgun",
	Short: "a simple and fast data processing tool",
	Long:  `Railgun is a simple and fast data processing tool built on go-reader, go-dfl, and go-simple-serializer.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(branch string, commit string) {
  // Set Global Variables
  gitBranch = branch
  gitCommit = commit
  
	rootCmd.AddCommand(&cobra.Command{
  	Use:   "version",
  	Short: "print version information to stdout",
  	Long:  "print version information to stdout",
  	Run: func(cmd *cobra.Command, args []string) {
  		fmt.Println("Version:"+railgun.Version)
  		fmt.Println("Branch: "+branch)
  		fmt.Println("Commit: "+commit)
  	},
  })
  
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		initConfig(viper.GetViper())
	})

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.railgun2.yaml)")

	// Config Flags
	rootCmd.PersistentFlags().StringArrayP("config-uri", "", []string{}, "the uri(s) to the config file")

	// AWS Flags
	rootCmd.PersistentFlags().StringP("aws-profile", "", "", "AWS Profile")
	rootCmd.PersistentFlags().StringP("aws-default-region", "", "", "AWS Default Region")
	rootCmd.PersistentFlags().StringP("aws-region", "", "", "AWS Region")
	rootCmd.PersistentFlags().StringP("aws-access-key-id", "", "", "AWS Access Key ID")
	rootCmd.PersistentFlags().StringP("aws-secret-access-key", "", "", "AWS Secret Access Key")
	rootCmd.PersistentFlags().StringP("aws-session-token", "", "", "AWS Session Token")
	rootCmd.PersistentFlags().StringP("aws-security-token", "", "", "AWS Security Token")
	rootCmd.PersistentFlags().StringP("aws-container-credentials-relative-uri", "", "", "AWS Container Credentials Relative URI")

	// HDFS Flags
	rootCmd.PersistentFlags().StringP("hdfs-name-node", "", "", "HDFS Name Node")

	// File Flags
	rootCmd.PersistentFlags().IntP("file-descriptor-limit", "", 4096, "limit to the number of open files")

	// Debub Flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "print verbose output to stdout")

	// Logging Flags
	rootCmd.PersistentFlags().String("info-destination", "stdout", "destination for info logs as a uri")
	rootCmd.PersistentFlags().String("info-format", "text", "info log format: text, properties, json, yaml, etc.")
	rootCmd.PersistentFlags().String("info-compression", "", "the compression algorithm for the info logs: none, gzip, or snappy")
	rootCmd.PersistentFlags().String("error-destination", "stderr", "destination for errors as a uri")
	rootCmd.PersistentFlags().String("error-compression", "", "the compression algorithm for the errors: none, gzip, or snappy")
	rootCmd.PersistentFlags().String("error-format", "text", "error log format: text, properties, json, yaml, etc.")

	// Runtime Flags
	serveCmd.PersistentFlags().Int("runtime-max-procs", 1, "Sets the maximum number of parallel processes.  If set to zero, then sets the maximum number of parallel processes to the number of CPUs on the machine. (https://godoc.org/runtime#GOMAXPROCS).")

}

// initConfig reads in config file and ENV variables if set.
func initConfig(v *viper.Viper) {
	if cfgFile != "" {
		// Use config file from the flag.
		v.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".railgun2" (without extension).
		v.AddConfigPath(home)
		v.SetConfigName(".railgun")
	}

	// Replace dashes with underscores for environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", v.ConfigFileUsed())
	}
}
