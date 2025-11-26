package internal

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// InitLogger initializes the global logger with JSON formatting
func InitLogger(env string) error {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		// Disable color in development for JSON consistency
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	// Always use JSON encoding for easy parsing
	config.Encoding = "json"

	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return err
	}

	Log = logger
	return nil
}

// SyncLogger flushes any buffered log entries
func SyncLogger() {
	if Log != nil {
		_ = Log.Sync()
	}
}
