package logger

import (
	"fmt"
	"log/slog"
	"os"
)

var (
	logger *slog.Logger
)

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo, // Changed from LevelDebug to LevelInfo to hide debug messages
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger = slog.New(handler)

}

func Info(format string, v ...interface{}) {
	logger.Info(fmt.Sprintf(format, v...))
}

func Debug(format string, v ...interface{}) {
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
