package feature

import (
	"os"
	"strconv"
)

type Name string

const (
	CbmigrateMongoHostOptsConfig Name = "MONGO_HOST_OPTS_CONFIG"
)

var Features = map[Name]bool{
	CbmigrateMongoHostOptsConfig: false,
}

func IsFeatureEnabled(feature Name) bool {
	if value, _ := strconv.ParseBool(os.Getenv("CBMIGRATE_" + string(feature))); value {
		return true
	}
	return Features[feature]
}
