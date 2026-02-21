package orderserver

import (
	"context"
	"io"
	"strings"

	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/log"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
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
		traceID, spanID := observability.GetTraceSpan(span)
		defer span.End()
		resp, err := handler(ctx, req)
		if err != nil {
			log.Log.Error("unary handler error", "trace_id", traceID, "span_id", spanID, "error", err)
			span.RecordError(err)
			grpcStatus, _ := status.FromError(err)
			span.SetStatus(codes.Error, grpcStatus.Message())
		}
		return resp, err
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
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
		traceID, spanID := observability.GetTraceSpan(w.span)
		log.Log.Error("stream RecvMsg error", "trace_id", traceID, "span_id", spanID, "error", err)
		w.span.RecordError(err)
		w.span.SetStatus(codes.Error, err.Error())
		w.span.End()
	}
	return err
}

func (w *serverStreamWrapper) SendMsg(m any) error {
	err := w.ServerStream.SendMsg(m)
	if err != nil {
		traceID, spanID := observability.GetTraceSpan(w.span)
		log.Log.Error("stream SendMsg error", "trace_id", traceID, "span_id", spanID, "error", err)
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
