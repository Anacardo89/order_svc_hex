package orderorchestrator

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type CtxKey string

const (
	CtxKeyReqID CtxKey = "request_id"
)

func ReqID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), CtxKeyReqID, reqID)
		span := trace.SpanFromContext(ctx)
		if span != nil && span.SpanContext().IsValid() {
			span.SetAttributes(attribute.String("request.id", reqID))
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
