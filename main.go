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
	Version = "" // - Ldflags will override this
)

func main() {
	common.SetVersion(Version)
	cmd.Execute()
}
