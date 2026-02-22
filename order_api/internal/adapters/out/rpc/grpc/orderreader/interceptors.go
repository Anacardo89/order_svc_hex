package orderreader

import (
	"context"
	"io"
	"strings"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/infra/log/loki/logger"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
		ctx, span := tracer.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient))
		log := logger.LogFromSpan(span, logger.BaseLogger)
		defer span.End()
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		carrier := metadataCarrier{md}
		otel.GetTextMapPropagator().Inject(ctx, carrier)
		ctx = metadata.NewOutgoingContext(ctx, md)
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			log.Error(ctx, "failed to invoke", ports.Field{Key: "error", Value: err})
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}
}

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
		log := logger.LogFromSpan(span, logger.BaseLogger)
		defer span.End()
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		carrier := metadataCarrier{md}
		otel.GetTextMapPropagator().Inject(ctx, carrier)
		ctx = metadata.NewOutgoingContext(ctx, md)
		cs, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			log.Error(ctx, "failed to stream", ports.Field{Key: "error", Value: err})
			span.RecordError(err)
			span.End()
			return cs, err
		}
		return &clientStreamWrapper{
			ClientStream:   cs,
			ctx:            ctx,
			span:           span,
			isServerStream: desc.ServerStreams,
		}, nil
	}
}

// For context propagation
type clientStreamWrapper struct {
	grpc.ClientStream
	ctx            context.Context
	span           trace.Span
	isServerStream bool
}

func (w *clientStreamWrapper) RecvMsg(m any) error {
	err := w.ClientStream.RecvMsg(m)
	if err == io.EOF {
		w.span.End()
		return err
	}
	if err != nil {
		log := logger.LogFromSpan(w.span, logger.BaseLogger)
		log.Error(w.ctx, "stream RecvMsg error", ports.Field{Key: "error", Value: err})
		w.span.RecordError(err)
		w.span.End()
		return err
	}
	return nil
}

func (w *clientStreamWrapper) CloseSend() error {
	err := w.ClientStream.CloseSend()
	if !w.isServerStream {
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
