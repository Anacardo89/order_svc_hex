package orderrepo

import (
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func failExec(span trace.Span, reason string, err error) error {
	span.RecordError(err)
	span.SetStatus(codes.Error, reason)
	return err
}

func failQueryRow[T any](span trace.Span, reason string, err error) (*T, error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, reason)
	return nil, err
}

func failQuery[T any](span trace.Span, reason string, err error) ([]*T, error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, reason)
	return nil, err
}
