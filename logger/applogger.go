package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var AppLogger *zap.Logger

func init() {
	logLevel := getLogLevelFromEnv()

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(logLevel)

	var err error
	AppLogger, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	AppLogger.Info("Logger initialized", zap.String("level", logLevel.String()))
}

func getLogLevelFromEnv() zapcore.Level {
	levelStr := os.Getenv("LOG_LEVEL")
	levelStr = strings.ToUpper(levelStr)

	switch levelStr {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "DPANIC":
		return zapcore.DPanicLevel
	case "PANIC":
		return zapcore.PanicLevel
	case "FATAL":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel // Default to Info if not specified or invalid
	}
}

// UpdateLogLevel allows changing the log level at runtime
func UpdateLogLevel(level zapcore.Level) {
	AppLogger.Core().Enabled(level)
	AppLogger.Info("Log level updated", zap.String("newLevel", level.String()))
}
