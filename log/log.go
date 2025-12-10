package log

import (
	"context"
	"log/slog"
)

type Logger interface {
	DebugContext(ctx context.Context, msg string, data ...interface{})
	InfoContext(ctx context.Context, msg string, data ...interface{})
	WarnContext(ctx context.Context, msg string, args ...interface{})
	ErrorContext(ctx context.Context, msg string, args ...interface{})

	Debug(msg string, data ...interface{})
	Info(msg string, data ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

var DefaultLogger Logger = slog.Default()

func DebugContext(ctx context.Context, msg string, data ...interface{}) {
	DefaultLogger.DebugContext(ctx, msg, data...)
}

func InfoContext(ctx context.Context, msg string, data ...interface{}) {
	DefaultLogger.InfoContext(ctx, msg, data...)
}

func WarnContext(ctx context.Context, msg string, args ...interface{}) {
	DefaultLogger.WarnContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	DefaultLogger.ErrorContext(ctx, msg, args...)
}

func Debug(msg string, data ...interface{}) {
	DefaultLogger.Debug(msg, data...)
}

func Info(msg string, data ...interface{}) {
	DefaultLogger.Info(msg, data...)
}

func Warn(msg string, args ...interface{}) {
	DefaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	DefaultLogger.Error(msg, args...)
}
