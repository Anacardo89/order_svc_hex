package orderorchestrator

import (
	"encoding/json"
	"net/http"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
)

type OrderHandler struct {
	orderReader ports.OrderReader
	orderWriter ports.OrderWriter
}

func NewOrderHandler(reader ports.OrderReader, writer ports.OrderWriter) *OrderHandler {
	return &OrderHandler{
		orderReader: reader,
		orderWriter: writer,
	}
}

// GET /orders/{id}
func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	order, err := h.orderReader.GetOrderByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req core.CreateOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.orderWriter.PublishCreate(r.Context(), &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// PATCH /orders/{id}/status
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	var req core.UpdateOrderStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.orderWriter.PublishStatusUpdate(r.Context(), &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
