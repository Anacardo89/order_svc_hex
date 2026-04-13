package orderserver

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/infra/log/loki/logger"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Tracing
func UnaryTraceInterceptor() grpc.UnaryServerInterceptor {
	tracer := otel.Tracer("order_svc.grpc.unary")
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		carrier := metadataCarrier{md}
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
		ctx, span := tracer.Start(
			ctx,
			info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		log := logger.LogFromSpan(span, logger.BaseLogger)
		defer span.End()

		resp, err := handler(ctx, req)
		if err != nil {
			log.Error(ctx, "unary handler error", ports.Field{Key: "error", Value: err})
			span.RecordError(err)
			grpcStatus, _ := status.FromError(err)
			span.SetStatus(codes.Error, grpcStatus.Message())
		}

		return resp, err
	}
}

func StreamTraceInterceptor() grpc.StreamServerInterceptor {
	tracer := otel.Tracer("order_svc.grpc.stream")
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		md, _ := metadata.FromIncomingContext(ss.Context())
		carrier := metadataCarrier{md}
		ctx := otel.GetTextMapPropagator().Extract(ss.Context(), carrier)
		ctx, span := tracer.Start(
			ctx,
			info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()
		wrapped := &serverStreamWrapper{ServerStream: ss, ctx: ctx, span: span}
		return handler(srv, wrapped)
	}
}

// For context propagation
type serverStreamWrapper struct {
	grpc.ServerStream
	ctx  context.Context
	span trace.Span
}

func (w *serverStreamWrapper) Context() context.Context {
	return w.ctx
}

func (w *serverStreamWrapper) RecvMsg(m any) error {
	err := w.ServerStream.RecvMsg(m)
	if err == io.EOF {
		w.span.End()
	}
	if err != nil {
		log := logger.LogFromSpan(w.span, logger.BaseLogger)
		log.Error(w.ctx, "stream RecvMsg error", ports.Field{Key: "error", Value: err})
		w.span.RecordError(err)
		w.span.SetStatus(codes.Error, err.Error())
		w.span.End()
	}
	return err
}

func (w *serverStreamWrapper) SendMsg(m any) error {
	err := w.ServerStream.SendMsg(m)
	if err != nil {
		log := logger.LogFromSpan(w.span, logger.BaseLogger)
		log.Error(w.ctx, "stream SendMsg error", ports.Field{Key: "error", Value: err})
		w.span.RecordError(err)
		w.span.SetStatus(codes.Error, err.Error())
		w.span.End()
	}
	return err
}

// Needed to ensure all lowercase
type metadataCarrier struct {
	metadata.MD
}

func (c metadataCarrier) Get(key string) string {
	values := c.MD[strings.ToLower(key)]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c metadataCarrier) Set(key, value string) {
	key = strings.ToLower(key)
	c.MD[key] = []string{value}
}

func (c metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(c.MD))
	for k := range c.MD {
		keys = append(keys, k)
	}
	return keys
}

// Metrics
func UnaryMetricsInterceptor(m *grpcMetrics) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		m.active.Add(ctx, 1)
		defer m.active.Add(ctx, -1)

		resp, err := handler(ctx, req)

		st, _ := status.FromError(err)
		metricAttrs := metric.WithAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", info.FullMethod),
			attribute.String("rpc.grpc.status_code", st.Code().String()),
		)
		m.duration.Record(ctx, time.Since(start).Seconds(), metricAttrs)
		m.requests.Add(ctx, 1, metricAttrs)
		return resp, err
	}
}

func StreamMetricsInterceptor(m *grpcMetrics) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		ctx := ss.Context()
		m.active.Add(ctx, 1)
		defer m.active.Add(ctx, -1)

		err := handler(srv, ss)

		st, _ := status.FromError(err)
		metricAttrs := metric.WithAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", info.FullMethod),
			attribute.String("rpc.grpc.status_code", st.Code().String()),
			attribute.Bool("rpc.is_streaming", true),
		)
		m.duration.Record(ctx, time.Since(start).Seconds(), metricAttrs)
		m.requests.Add(ctx, 1, metricAttrs)
		return err
	}
}
