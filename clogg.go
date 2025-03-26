package clogg

import (
	"context"
	"log/slog"
	"os"
)

// NewClogg initializes the logger
func NewClogg() {
	handler := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(handler)
}

// Debug logs a debug message
func Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(
		ctx,
		slog.LevelDebug,
		msg,
		attrs...,
	)
}

// Info logs an info message
func Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		msg,
		attrs...,
	)
}

// Warn logs a warning message
func Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(
		ctx,
		slog.LevelWarn,
		msg,
		attrs...,
	)
}

// Error logs an error message
func Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(
		ctx,
		slog.LevelError,
		msg,
		attrs...,
	)
}
