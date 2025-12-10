package log

import (
	"context"
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
