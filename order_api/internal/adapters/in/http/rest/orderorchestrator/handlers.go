package orderorchestrator

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/trace"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
	"github.com/Anacardo89/order_svc_hex/order_api/pkg/log"
	"github.com/Anacardo89/order_svc_hex/order_api/pkg/observability"
	"github.com/Anacardo89/order_svc_hex/order_api/pkg/validator"
)

type OrderHandler struct {
	svc core.OrderOrchestrator
}

func NewOrderHandler(reader ports.OrderReader, writer ports.OrderWriter) *OrderHandler {
	svc := NewOrderService(reader, writer)
	return &OrderHandler{
		svc: svc,
	}
}

type HealthCheckResp struct {
	Status string `json:"status"`
}

// /
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body := HealthCheckResp{
		Status: "OK",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(body)
}

// 404
func CatchAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body := ErrorResp{
		Error: "endpoint not found",
	}
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(body)
}

type GetOrderResp struct {
	Order *core.Order `json:"order"`
}

// GET /orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	// Setup
	ctx := r.Context()
	reqID := ctx.Value(CtxKeyReqID)
	span := trace.SpanFromContext(r.Context())
	traceID, spanID := observability.GetTraceSpan(span)
	w.Header().Set("Content-Type", "application/json")

	// Execution
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		log.Log.Error("failed to parse id from URL", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
		failHttp(w, ctx, http.StatusBadRequest, "invalid path", err)
		return
	}
	order, err := h.svc.GetOrder(ctx, id)
	if err != nil {
		log.Log.Error("failed to get order from order_svc", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
		failHttp(w, ctx, http.StatusNotFound, "invalid path", err)
		return
	}
	resp := GetOrderResp{
		Order: order,
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		log.Log.Error("failed to encode response body", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
		failHttp(w, ctx, http.StatusInternalServerError, "internal error", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Log.Error("failed to send response to client", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
	}
}

type GetOrdersResp struct {
	Orders []*core.Order `json:"orders"`
}

// GET /orders
func (h *OrderHandler) ListOrdersByStatus(w http.ResponseWriter, r *http.Request) {
	// Setup
	ctx := r.Context()
	reqID := ctx.Value(CtxKeyReqID)
	span := trace.SpanFromContext(r.Context())
	traceID, spanID := observability.GetTraceSpan(span)
	w.Header().Set("Content-Type", "application/json")

	// Execution
	statusStr := r.URL.Query().Get("status")
	if statusStr == "" {
		log.Log.Error("request with empty query", "request_id", reqID, "trace_id", traceID, "span_id", spanID)
		failHttp(w, ctx, http.StatusBadRequest, "request must contain 'status' in query", errors.New("request with empty query"))
		return
	}
	status, err := core.MapStrToStatus(statusStr)
	if err != nil {
		log.Log.Error("invalid status", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
		failHttp(w, ctx, http.StatusBadRequest, "status must be either 'pending', 'confirmed' or 'failed'", err)
		return
	}
	orders, err := h.svc.ListOrdersByStatus(ctx, status)
	if err != nil {
		log.Log.Error("failed to get order from order_svc", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
		failHttp(w, ctx, http.StatusInternalServerError, "internal error", err)
		return
	}
	resp := GetOrdersResp{
		Orders: orders,
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		log.Log.Error("failed to encode response body", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
		failHttp(w, ctx, http.StatusInternalServerError, "internal error", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Log.Error("failed to send response to client", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
	}
}

// POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Setup
	ctx := r.Context()
	reqID := ctx.Value(CtxKeyReqID)
	span := trace.SpanFromContext(r.Context())
	traceID, spanID := observability.GetTraceSpan(span)
	w.Header().Set("Content-Type", "application/json")

	// Execution
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		log.Log.Error("failed to read request body", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
		failHttp(w, ctx, http.StatusBadRequest, "invalid request body", err)
		return
	}
	var reqBody core.CreateOrder
	if err := validator.ParseAndValidate(raw, &reqBody); err != nil {
		if strings.Contains(err.Error(), "missing fields") {
			log.Log.Error("missing required fields", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
			failHttp(w, ctx, http.StatusBadRequest, err.Error(), err)
		} else {
			log.Log.Error("failed to parse JSON from body", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
			failHttp(w, ctx, http.StatusBadRequest, "invalid request body", err)
		}
		return
	}
	if err := h.svc.CreateOrder(ctx, &reqBody); err != nil {
		log.Log.Error("failed to create order", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
		failHttp(w, ctx, http.StatusInternalServerError, "internal error", err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

type UpdateOrderStatusReq struct {
	Status string `json:"status" validate:"required"`
}

// PUT /orders/{id}/status
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	// Setup
	ctx := r.Context()
	reqID := ctx.Value(CtxKeyReqID)
	span := trace.SpanFromContext(r.Context())
	traceID, spanID := observability.GetTraceSpan(span)
	w.Header().Set("Content-Type", "application/json")

	// Execution
	vars := mux.Vars(r)
	id := vars["id"]
	_, err := uuid.Parse(id)
	if err != nil {
		log.Log.Error("id provided not valid", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
		failHttp(w, ctx, http.StatusBadRequest, "invalid path", err)
		return
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		log.Log.Error("failed to read request body", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
		failHttp(w, ctx, http.StatusBadRequest, "invalid request body", err)
		return
	}
	var reqBody UpdateOrderStatusReq
	if err := validator.ParseAndValidate(raw, &reqBody); err != nil {
		if strings.Contains(err.Error(), "missing fields") {
			log.Log.Error("missing required fields", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
			failHttp(w, ctx, http.StatusBadRequest, err.Error(), err)
		} else {
			log.Log.Error("failed to parse JSON from body", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
			failHttp(w, ctx, http.StatusBadRequest, "invalid request body", err)
		}
		return
	}
	status, err := core.MapStrToStatus(reqBody.Status)
	if err != nil {
		log.Log.Error("invalid status", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
		failHttp(w, ctx, http.StatusBadRequest, "status must be either 'pending', 'confirmed' or 'failed'", err)
		return
	}
	req := core.UpdateOrderStatus{
		ID:     id,
		Status: *status,
	}
	if err := h.svc.UpdateOrderStatus(ctx, &req); err != nil {
		log.Log.Error("failed to update order", "request_id", reqID, "trace_id", traceID, "span_id", spanID, "error", err)
		failHttp(w, ctx, http.StatusInternalServerError, "internal error", err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
