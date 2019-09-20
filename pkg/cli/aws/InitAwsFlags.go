// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package aws

import (
	"github.com/spatialcurrent/pflag"
)

// InitAwsFlags initializes the AWS flags.
func InitAwsFlags(flag *pflag.FlagSet) {
	flag.StringP(FlagAwsProfile, "", "", "AWS Profile")
	flag.StringP(FlagAwsDefaultRegion, "", "", "AWS Default Region")
	flag.StringP(FlagAwsRegion, "", "", "AWS Region")
	flag.StringP(FlagAwsAccessKeyId, "", "", "AWS Access Key ID")
	flag.StringP(FlagAwsSecretAccessKey, "", "", "AWS Secret Access Key")
	flag.StringP(FlagAwsSessionToken, "", "", "AWS Session Token")
	flag.StringP(FlagAwsSecurityToken, "", "", "AWS Security Token")
	flag.StringP(FlagAwsContainerCredentialsRelativeUri, "", "", "AWS Container Credentials Relative URI")
}
