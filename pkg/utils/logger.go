package utils

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Logger = slog.New(handler)
}

// NewLogger creates a new logger instance
func NewLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}

func SetLogLevel(level slog.Level) {
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
