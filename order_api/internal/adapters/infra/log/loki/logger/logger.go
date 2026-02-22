package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
)

var BaseLogger ports.Logger

type Logger struct {
	slogLogger *slog.Logger
}

func NewLogger(endpoint string, labels map[string]string) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	lokiHandler := NewLokiHandler(handler, endpoint, labels)
	slogLogger := slog.New(lokiHandler)
	return &Logger{slogLogger: slogLogger}
}

func (l *Logger) With(fields ...ports.Field) ports.Logger {
	var attrs []any
	for _, f := range fields {
		attrs = append(attrs, slog.Any(f.Key, f.Value))
	}
	return &Logger{slogLogger: l.slogLogger.With(attrs...)}
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...ports.Field) {
	l.log(ctx, slog.LevelInfo, msg, fields)
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...ports.Field) {
	l.log(ctx, slog.LevelDebug, msg, fields)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields ...ports.Field) {
	l.log(ctx, slog.LevelWarn, msg, fields)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...ports.Field) {
	l.log(ctx, slog.LevelError, msg, fields)
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, fields []ports.Field) {
	var attrs []slog.Attr
	for _, f := range fields {
		attrs = append(attrs, slog.Any(f.Key, f.Value))
	}
	l.slogLogger.LogAttrs(ctx, level, msg, attrs...)
}
