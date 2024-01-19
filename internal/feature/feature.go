package feature

import (
	"os"
	"strconv"
)

type Name string

const (
	CbmigrateMongoHostOptsConfig Name = "CBMIGRATE_MONGO_HOST_OPTS_CONFIG"
)

var Features = map[Name]bool{
	CbmigrateMongoHostOptsConfig: false,
}

func IsFeatureEnabled(feature Name) bool {
	if value, _ := strconv.ParseBool(os.Getenv(string(CbmigrateMongoHostOptsConfig))); value {
		return true
	}
	return Features[CbmigrateMongoHostOptsConfig]
}
