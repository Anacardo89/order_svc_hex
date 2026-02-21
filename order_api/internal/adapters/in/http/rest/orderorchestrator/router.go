package orderorchestrator

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

func NewRouter(h *OrderHandler) http.Handler {
	r := mux.NewRouter()
	r.Use(otelmux.Middleware("order_api.rest"))
	r.Use(ReqID)
	// Health check
	r.Handle("/", http.HandlerFunc(HealthCheck)).Methods("GET")
	// Orders
	r.Handle("/orders", http.HandlerFunc(h.CreateOrder)).Methods("POST")
	r.Handle("/orders", http.HandlerFunc(h.ListOrdersByStatus)).Methods("GET")
	r.Handle("/orders/{id}", http.HandlerFunc(h.GetOrder)).Methods("GET")
	r.Handle("/orders/{id}/status", http.HandlerFunc(h.UpdateOrderStatus)).Methods("PUT")
	// Catch-all 404
	r.NotFoundHandler = http.HandlerFunc(CatchAll)
	return r
}
