package ports

import "context"

type Field struct {
	Key   string
	Value any
}

type Logger interface {
	With(fields ...Field) Logger
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
}
