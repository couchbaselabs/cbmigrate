package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func Init() {
	// Create a configuration for the logger.
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
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

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	// Replace the global logger.
	zap.ReplaceGlobals(logger)
}

func EnableDebugLevel() {
	logger = logger.WithOptions(zap.IncreaseLevel(zap.DebugLevel))
	zap.ReplaceGlobals(logger)
}
