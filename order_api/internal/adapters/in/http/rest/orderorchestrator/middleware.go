package orderorchestrator

import (
	"context"
	"net/http"
	"time"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/infra/log/loki/logger"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
	"github.com/Anacardo89/order_svc_hex/order_api/pkg/observability"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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

func Metrics(metrics *ReqMetrics) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			start := time.Now()
			metrics.active.Add(ctx, 1)
			defer metrics.active.Add(ctx, -1)
			rw := newRWInterceptor(w)
			next.ServeHTTP(rw, r)
			route := mux.CurrentRoute(r)
			path, _ := route.GetPathTemplate()
			if path == "" {
				path = "unknown"
			}
			attrs := metric.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", path),
				attribute.Int("http.status_code", rw.status),
			)
			metrics.counter.Add(ctx, 1, attrs)
			metrics.duration.Record(ctx, time.Since(start).Seconds(), attrs)
		})
	}
}

// To capture status code for metrics
type RWInterceptor struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func newRWInterceptor(w http.ResponseWriter) *RWInterceptor {
	return &RWInterceptor{
		ResponseWriter: w,
		status:         http.StatusOK,
		wroteHeader:    false,
	}
}

func (w *RWInterceptor) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *RWInterceptor) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}
