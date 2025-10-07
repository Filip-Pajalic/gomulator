package logger

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
)

var (
	logger       *slog.Logger
	debugEnabled bool
)

func init() {
	// Use higher log level for WASM to reduce console overhead
	logLevel := slog.LevelInfo
	if runtime.GOARCH == "wasm" {
		logLevel = slog.LevelWarn // Only show warnings and errors in WASM
	}

	debugEnabled = logLevel <= slog.LevelDebug

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger = slog.New(handler)

}

func Info(format string, v ...interface{}) {
	logger.Info(fmt.Sprintf(format, v...))
}

func Debug(format string, v ...interface{}) {
	// Skip formatting entirely if debug is disabled
	if !debugEnabled {
		return
	}
	logger.Debug(fmt.Sprintf(format, v...))
}

func Warn(format string, v ...interface{}) {
	logger.Warn(fmt.Sprintf(format, v...))
}

func Error(format string, v ...interface{}) {
	logger.Error(fmt.Sprintf(format, v...))
}

func Fatal(format string, v ...interface{}) {
	logger.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}
