package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger is a wrapper around slog.Logger to enforce structure.
type Logger struct {
	*slog.Logger
}

// New creates a new JSON logger.
func New(serviceName string) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler).With(
		slog.String("service", serviceName),
	)

	return &Logger{logger}
}

// Info logs an info message.
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// WithContext adds tracing info from context (if available).
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Placeholder: Need to extract TraceID from OTEL context here later
	return l
}
