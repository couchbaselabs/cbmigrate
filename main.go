/*
Copyright © 2023 Couchbase Inc.
*/
package main

import (
	"github.com/couchbaselabs/cbmigrate/cmd"
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/internal/pkg/logger"
)

func init() {
	logger.Init()
}

var (
	Version = "1.0.0" // This will be overridden by -ldflags
)

func main() {
	common.SetVersion(Version)
	cmd.Execute()
}
