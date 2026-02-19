package orderreader

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	tracer := otel.Tracer("order_api.grpc.unary")
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx, span := tracer.Start(ctx, method,
			trace.WithSpanKind(trace.SpanKindClient),
		)
		defer span.End()
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			span.RecordError(err)
		}
		return err
	}
}

// Needed for client streams or for per message observability on server streams (needs to wrap stream.Recv)
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	tracer := otel.Tracer("order_api.grpc.stream")
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		ctx, span := tracer.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient))
		cs, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			span.RecordError(err)
			span.End()
			return cs, err
		}
		return &clientStreamWrapper{ClientStream: cs, span: span}, nil
	}
}

type clientStreamWrapper struct {
	grpc.ClientStream
	span trace.Span
}

func (w *clientStreamWrapper) CloseSend() error {
	err := w.ClientStream.CloseSend()
	w.span.End()
	return err
}
