package utils

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger
var currentLogLevel = slog.LevelInfo

func init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: currentLogLevel,
	})
	Logger = slog.New(handler)
}

// NewLogger creates a new logger instance using the current log level
func NewLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: currentLogLevel,
	})
	return slog.New(handler)
}

func SetLogLevel(level slog.Level) {
	currentLogLevel = level
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	Logger = slog.New(handler)
}

func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}
