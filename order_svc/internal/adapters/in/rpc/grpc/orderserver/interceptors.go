package orderserver

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	tracer := otel.Tracer("order_svc.grpc.unary")
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()
		resp, err := handler(ctx, req)
		if err != nil {
			span.RecordError(err)
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
		ctx, span := tracer.Start(ss.Context(), info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = ctx

		return handler(srv, wrapped)
	}
}
