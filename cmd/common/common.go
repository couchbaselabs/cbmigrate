package common

import "github.com/couchbaselabs/cbmigrate/cmd/flag"

const (
	Debug = "debug"
)

func GetDebugFlag() flag.Flag {
	return &flag.BoolFlag{
		Name:  Debug,
		Usage: "Enable debug output.",
	}
}
