/*
Copyright © 2023 Couchbase Inc.
*/
package main

import (
	"github.com/couchbaselabs/cbmigrate/cmd"
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/internal/pkg/logger"
	"go.uber.org/zap/zapcore"
)

func init() {
	logger.Init(zapcore.InfoLevel)
}

var (
	Version = "" // - Ldflags will override this
)

func main() {
	common.SetVersion(Version)
	cmd.Execute()
}
