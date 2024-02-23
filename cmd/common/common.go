package common

import "github.com/couchbaselabs/cbmigrate/cmd/flag"

const (
	Verbose = "verbose"
)

var verboseEnabled = false

func GetVerboseFlag() flag.Flag {
	return &flag.BoolFlag{
		Name:  Verbose,
		Usage: "enable verbose output.",
	}
}

func SetVerboseEnabled() {
	verboseEnabled = true
}

func IsVerboseEnabled() bool {
	return verboseEnabled
}
