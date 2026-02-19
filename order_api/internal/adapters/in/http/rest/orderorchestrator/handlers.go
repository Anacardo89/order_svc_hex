package orderorchestrator

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

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

type GetOrderResp struct {
	Order *core.Order `json:"order"`
}

// GET /orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	// Setup
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	// Execution
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		slog.Error("failed to parse id from URL", "error", err)
		failHttp(w, http.StatusBadRequest, "invalid path")
		return
	}
	order, err := h.svc.GetOrder(ctx, id)
	if err != nil {
		slog.Error("failed to get order from order_svc", "error", err)
		failHttp(w, http.StatusNotFound, "internal error")
		return
	}
	resp := GetOrderResp{
		Order: order,
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		slog.Error("failed to encode response body", "error", err)
		failHttp(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(buf.Bytes()); err != nil {
		slog.Error("failed to send response to client", "error", err)
	}
}

type GetOrdersResp struct {
	Orders []*core.Order `json:"orders"`
}

// GET /orders
func (h *OrderHandler) ListOrdersByStatus(w http.ResponseWriter, r *http.Request) {
	// Setup
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	// Execution
	statusStr := r.URL.Query().Get("status")
	if statusStr == "" {
		slog.Error("request with empty query")
		failHttp(w, http.StatusBadRequest, "request must contain 'status' in query")
		return
	}
	status, err := core.MapStrToStatus(statusStr)
	if err != nil {
		slog.Error("invalid status", "error", err)
		failHttp(w, http.StatusBadRequest, "status must be either 'pending', 'confirmed' or 'failed'")
		return
	}
	orders, err := h.svc.ListOrdersByStatus(ctx, status)
	if err != nil {
		slog.Error("failed to get order from order_svc", "error", err)
		failHttp(w, http.StatusInternalServerError, "internal error")
		return
	}
	resp := GetOrdersResp{
		Orders: orders,
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		slog.Error("failed to encode response body", "error", err)
		failHttp(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(buf.Bytes()); err != nil {
		slog.Error("failed to send response to client", "error", err)
	}
}

// POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Setup
	ctx := r.Context()

	// Execution
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", "error", err)
		failHttp(w, http.StatusBadRequest, "invalid request body")
		return
	}
	var reqBody core.CreateOrder
	if err := validator.ParseAndValidate(raw, &reqBody); err != nil {
		if strings.Contains(err.Error(), "missing fields") {
			slog.Error("missing required fields", "error", err)
			failHttp(w, http.StatusBadRequest, err.Error())
		} else {
			slog.Error("failed to parse JSON from body", "error", err)
			failHttp(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}
	if err := h.svc.CreateOrder(ctx, &reqBody); err != nil {
		slog.Error("failed to create order", "error", err)
		failHttp(w, http.StatusInternalServerError, "internal error")
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

	// Execution
	vars := mux.Vars(r)
	id := vars["id"]
	_, err := uuid.Parse(id)
	if err != nil {
		slog.Error("id provided not valid", "error", err)
		failHttp(w, http.StatusBadRequest, "invalid path")
		return
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", "error", err)
		failHttp(w, http.StatusBadRequest, "invalid request body")
		return
	}
	var reqBody UpdateOrderStatusReq
	if err := validator.ParseAndValidate(raw, &reqBody); err != nil {
		if strings.Contains(err.Error(), "missing fields") {
			slog.Error("missing required fields", "error", err)
			failHttp(w, http.StatusBadRequest, err.Error())
		} else {
			slog.Error("failed to parse JSON from body", "error", err)
			failHttp(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}
	status, err := core.MapStrToStatus(reqBody.Status)
	if err != nil {
		slog.Error("invalid status", "error", err)
		failHttp(w, http.StatusBadRequest, "status must be either 'pending', 'confirmed' or 'failed'")
		return
	}
	req := core.UpdateOrderStatus{
		ID:     id,
		Status: *status,
	}
	if err := h.svc.UpdateOrderStatus(ctx, &req); err != nil {
		slog.Error("failed to update order", "error", err)
		failHttp(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
