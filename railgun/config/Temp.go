package config

import (
	"strings"
)

type Temp struct {
	Uri string `viper:"temp-uri"`
}

func (t Temp) IsAthenaStoredQuery() bool {
	return strings.HasPrefix(t.Uri, "athena://")
}

func (t Temp) IsS3Bucket() bool {
	return strings.HasPrefix(t.Uri, "s3://")
}

func (t Temp) Map() map[string]interface{} {
	return map[string]interface{}{
		"Uri": t.Uri,
	}
}
