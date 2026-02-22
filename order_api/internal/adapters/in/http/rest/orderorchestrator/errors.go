package orderorchestrator

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/infra/log/loki/logger"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ErrorResp struct {
	Error string `json:"error"`
}

func (h *OrderHandler) failHttp(w http.ResponseWriter, ctx context.Context, status int, outMsg string, err error) {
	log := logger.LogFromCtx(ctx, logger.BaseLogger)
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	w.WriteHeader(status)
	resp := ErrorResp{Error: outMsg}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error(ctx, "failed to encode error response body", ports.Field{Key: "error", Value: err})
	}
}
