package logger

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
	"github.com/Anacardo89/order_svc_hex/order_api/pkg/observability"
	"go.opentelemetry.io/otel/trace"
)

type CtxKey string

const (
	CtxKeyReqID  CtxKey = "request_id"
	CtxKeyLogger CtxKey = "logger"
)

func LogFromCtx(ctx context.Context, defaultLogger ports.Logger) ports.Logger {
	l, ok := ctx.Value(CtxKeyLogger).(ports.Logger)
	if !ok {
		return defaultLogger
	}
	return l
}

func LogFromSpan(span trace.Span, defaultLogger ports.Logger) ports.Logger {
	traceID, spanID := observability.GetTraceSpan(span)
	return defaultLogger.With(
		ports.Field{Key: "trace_id", Value: traceID},
		ports.Field{Key: "span_id", Value: spanID},
	)
}
