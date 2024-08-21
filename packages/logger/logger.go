package logger

import (
	"fmt"
	"log/slog"
	"os"
)

var (
	log *slog.Logger
)

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	log = slog.New(handler)

}

func Info(format string, v ...interface{}) {
	log.Info(fmt.Sprintf(format, v...))
}

func Warn(format string, v ...interface{}) {
	log.Warn(fmt.Sprintf(format, v...))
}

func Error(format string, v ...interface{}) {
	log.Error(fmt.Sprintf(format, v...))
}

func Fatal(format string, v ...interface{}) {
	log.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}
