// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package util

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

func ConnectToAWS(awsAccessKeyId string, awsSecretAccessKey string, awsSessionToken string, awsRegion string) (*session.Session, error) {

	// https://docs.aws.amazon.com/sdk-for-go/api/aws/#Config
	config := aws.Config{
		MaxRetries: aws.Int(3),
		Region:     aws.String(awsRegion),
	}

	// The credentials object to use when signing requests. Defaults to a
	// chain of credential providers to search for credentials in environment
	// variables, shared credential file, and EC2 Instance Roles.
	if len(awsAccessKeyId) > 0 && len(awsSecretAccessKey) > 0 {
		config.Credentials = credentials.NewStaticCredentials(
			awsAccessKeyId,
			awsSecretAccessKey,
			awsSessionToken)
	}

	return session.NewSessionWithOptions(session.Options{Config: config})
}
