package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger
var atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)

func Init() {
	// Create a configuration for the logger.
	config := zap.Config{
		Level:            atomicLevel,
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	config.DisableCaller = true
	config.DisableStacktrace = true
	// Enable color
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// Disable timestamp
	config.EncoderConfig.TimeKey = ""
	var err error
	logger, err = config.Build()
	if err != nil {
		panic(err)
	}
	// Replace the global logger.
	zap.ReplaceGlobals(logger)
}

func EnableDebugLevel() {
	atomicLevel.SetLevel(zap.DebugLevel)
}

func EnableErrorLevel() {
	atomicLevel.SetLevel(zap.ErrorLevel)
}
