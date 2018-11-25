package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type AWS struct {
	DefaultRegion   string `viper:"aws-default-region"`
	AccessKeyId     string `viper:"aws-access-key-id"`
	SecretAccessKey string `viper:"aws-secret-access-key"`
	SessionToken    string `viper:"aws-session-token"`
}

func (a AWS) Config() aws.Config {
	// https://docs.aws.amazon.com/sdk-for-go/api/aws/#Config
	c := aws.Config{
		MaxRetries: aws.Int(3),
		Region:     aws.String(a.DefaultRegion),
	}

	// The credentials object to use when signing requests. Defaults to a
	// chain of credential providers to search for credentials in environment
	// variables, shared credential file, and EC2 Instance Roles.
	if len(a.AccessKeyId) > 0 && len(a.SecretAccessKey) > 0 {
		c.Credentials = credentials.NewStaticCredentials(
			a.AccessKeyId,
			a.SecretAccessKey,
			a.SessionToken)
	}

	return c
}

func (a AWS) SessionOptions() session.Options {
	return session.Options{Config: a.Config()}
}

func (a AWS) Map() map[string]interface{} {
	return map[string]interface{}{
		"DefaultRegion":   a.DefaultRegion,
		"AccessKeyId":     a.AccessKeyId,
		"SecretAccessKey": a.SecretAccessKey,
		"SessionToken":    a.SessionToken,
	}
}
