package orderorchestrator

import (
	"context"
	"net/http"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/infra/log/loki/logger"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
	"github.com/Anacardo89/order_svc_hex/order_api/pkg/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func ReqID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), logger.CtxKeyReqID, reqID)
		span := trace.SpanFromContext(ctx)
		if span != nil && span.SpanContext().IsValid() {
			span.SetAttributes(attribute.String("request.id", reqID))
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Log(baseLogger ports.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			reqID := ctx.Value(logger.CtxKeyReqID)
			span := trace.SpanFromContext(ctx)
			traceID, spanID := observability.GetTraceSpan(span)
			enrichedLogger := baseLogger.With(
				ports.Field{Key: "request_id", Value: reqID},
				ports.Field{Key: "trace_id", Value: traceID},
				ports.Field{Key: "span_id", Value: spanID},
			)
			ctx = context.WithValue(ctx, logger.CtxKeyLogger, enrichedLogger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
