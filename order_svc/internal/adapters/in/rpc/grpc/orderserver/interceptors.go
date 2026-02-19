package orderserver

import (
	"context"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
		defer span.End()
		return handler(ctx, req)
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
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = ctx
		return handler(srv, wrapped)
	}
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
