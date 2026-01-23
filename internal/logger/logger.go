package logger

import (
	"log/slog"
	"os"
)

// L is the global logger instance
var L *slog.Logger

func init() {
	// Default logger - will be replaced by Wails logger via SetLogger()
	L = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// SetLogger replaces the default logger with the provided one.
// Call this with app.Logger after creating the Wails application
// to unify log formats.
func SetLogger(l *slog.Logger) {
	if l != nil {
		L = l
	}
}

// Convenience functions for structured logging

func Info(msg string, args ...any)  { L.Info(msg, args...) }
func Error(msg string, args ...any) { L.Error(msg, args...) }
func Debug(msg string, args ...any) { L.Debug(msg, args...) }
func Warn(msg string, args ...any)  { L.Warn(msg, args...) }
