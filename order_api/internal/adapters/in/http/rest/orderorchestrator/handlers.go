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

	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/infra/log/loki/logger"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
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

// GET /orders/{id}
type GetOrderResp struct {
	Order *core.Order `json:"order"`
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	// Setup
	ctx := r.Context()
	log := logger.LogFromCtx(ctx, logger.BaseLogger)
	w.Header().Set("Content-Type", "application/json")

	// Execution
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		log.Error(ctx, "failed to parse id from URL", ports.Field{Key: "error", Value: err})
		h.failHttp(w, ctx, http.StatusBadRequest, "invalid path", err)
		return
	}
	order, err := h.svc.GetOrder(ctx, id)
	if err != nil {
		log.Error(ctx, "failed to get order from order_svc", ports.Field{Key: "error", Value: err})
		h.failHttp(w, ctx, http.StatusNotFound, "invalid path", err)
		return
	}
	resp := GetOrderResp{
		Order: order,
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		log.Error(ctx, "failed to encode response body", ports.Field{Key: "error", Value: err})
		h.failHttp(w, ctx, http.StatusInternalServerError, "internal error", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Error(ctx, "failed to send response to client", ports.Field{Key: "error", Value: err})
	}
}

// GET /orders
type GetOrdersResp struct {
	Orders []*core.Order `json:"orders"`
}

func (h *OrderHandler) ListOrdersByStatus(w http.ResponseWriter, r *http.Request) {
	// Setup
	ctx := r.Context()
	log := logger.LogFromCtx(ctx, logger.BaseLogger)
	w.Header().Set("Content-Type", "application/json")

	// Execution
	statusStr := r.URL.Query().Get("status")
	if statusStr == "" {
		log.Error(ctx, "request with empty query")
		h.failHttp(w, ctx, http.StatusBadRequest, "request must contain 'status' in query", errors.New("request with empty query"))
		return
	}
	status, err := core.MapStrToStatus(statusStr)
	if err != nil {
		log.Error(ctx, "invalid status", ports.Field{Key: "error", Value: err})
		h.failHttp(w, ctx, http.StatusBadRequest, "status must be either 'pending', 'confirmed' or 'failed'", err)
		return
	}
	orders, err := h.svc.ListOrdersByStatus(ctx, status)
	if err != nil {
		log.Error(ctx, "failed to get order from order_svc", ports.Field{Key: "error", Value: err})
		h.failHttp(w, ctx, http.StatusInternalServerError, "internal error", err)
		return
	}
	resp := GetOrdersResp{
		Orders: orders,
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		log.Error(ctx, "failed to encode response body", ports.Field{Key: "error", Value: err})
		h.failHttp(w, ctx, http.StatusInternalServerError, "internal error", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Error(ctx, "failed to send response to client", ports.Field{Key: "error", Value: err})
	}
}

// POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Setup
	ctx := r.Context()
	log := logger.LogFromCtx(ctx, logger.BaseLogger)
	w.Header().Set("Content-Type", "application/json")

	// Execution
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(ctx, "failed to read request body", ports.Field{Key: "error", Value: err})
		h.failHttp(w, ctx, http.StatusBadRequest, "invalid request body", err)
		return
	}
	var reqBody core.CreateOrder
	if err := validator.ParseAndValidate(raw, &reqBody); err != nil {
		if strings.Contains(err.Error(), "missing fields") {
			log.Error(ctx, "missing required fields", ports.Field{Key: "error", Value: err})
			h.failHttp(w, ctx, http.StatusBadRequest, err.Error(), err)
		} else {
			log.Error(ctx, "failed to parse JSON from body", ports.Field{Key: "error", Value: err})
			h.failHttp(w, ctx, http.StatusBadRequest, "invalid request body", err)
		}
		return
	}
	if err := h.svc.CreateOrder(ctx, &reqBody); err != nil {
		log.Error(ctx, "failed to create order", ports.Field{Key: "error", Value: err})
		h.failHttp(w, ctx, http.StatusInternalServerError, "internal error", err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// PUT /orders/{id}/status
type UpdateOrderStatusReq struct {
	Status string `json:"status" validate:"required"`
}

func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	// Setup
	ctx := r.Context()
	log := logger.LogFromCtx(ctx, logger.BaseLogger)
	w.Header().Set("Content-Type", "application/json")

	// Execution
	vars := mux.Vars(r)
	id := vars["id"]
	_, err := uuid.Parse(id)
	if err != nil {
		log.Error(ctx, "id provided not valid", ports.Field{Key: "error", Value: err})
		h.failHttp(w, ctx, http.StatusBadRequest, "invalid path", err)
		return
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(ctx, "failed to read request body", ports.Field{Key: "error", Value: err})
		h.failHttp(w, ctx, http.StatusBadRequest, "invalid request body", err)
		return
	}
	var reqBody UpdateOrderStatusReq
	if err := validator.ParseAndValidate(raw, &reqBody); err != nil {
		if strings.Contains(err.Error(), "missing fields") {
			log.Error(ctx, "missing required fields", ports.Field{Key: "error", Value: err})
			h.failHttp(w, ctx, http.StatusBadRequest, err.Error(), err)
		} else {
			log.Error(ctx, "failed to parse JSON from body", ports.Field{Key: "error", Value: err})
			h.failHttp(w, ctx, http.StatusBadRequest, "invalid request body", err)
		}
		return
	}
	status, err := core.MapStrToStatus(reqBody.Status)
	if err != nil {
		log.Error(ctx, "invalid status", ports.Field{Key: "error", Value: err})
		h.failHttp(w, ctx, http.StatusBadRequest, "status must be either 'pending', 'confirmed' or 'failed'", err)
		return
	}
	req := core.UpdateOrderStatus{
		ID:     id,
		Status: *status,
	}
	if err := h.svc.UpdateOrderStatus(ctx, &req); err != nil {
		log.Error(ctx, "failed to update order", ports.Field{Key: "error", Value: err})
		h.failHttp(w, ctx, http.StatusInternalServerError, "internal error", err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
