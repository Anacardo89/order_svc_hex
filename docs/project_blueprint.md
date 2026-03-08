# Overview  

This project is a blueprint for Hexagonal architecture and distributed observability.  
It consists of two services: `order_api` and `order_svc`, each has it's own `go.mod` to be independent from the other, they are kept in the same Github repository for testing simplicity using Docker compose.  

## Architecture  

### order_api  

The entrypoint of the system, it exposes a REST API with the endpoints provided in the following router implementation:  
```go
func NewRouter(h *OrderHandler) http.Handler {
	r := mux.NewRouter()
	r.Use(otelmux.Middleware("order_api.rest"))
	r.Use(ReqID)
	r.Use(Log(logger.BaseLogger))
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
```

When a request is received, `order_api` sends the necessary data to fulfill the request to `order_svc`.  
`order_api` is CQRS compliant, so write operations (commands) are handled via Kafka and read operations (queries) are handled via gRPC.  

The core of `order_api`:  
```go
type Order struct {
	ID        uuid.UUID      `json:"id"`
	Items     map[string]int `json:"items"`
	Status    Status         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Commands
type CreateOrderCmd struct {
	Items  map[string]int `json:"items" validate:"required"`
	Status Status         `json:"status"`
}

type UpdateOrderStatusCmd struct {
	ID     string `json:"id" validate:"required"`
	Status Status `json:"status" validate:"required"`
}

// Queries
type GetOrderQry struct {
	ID uuid.UUID
}

type ListOrdersByStatusQry struct {
	Status Status
}
```

The orchestrator port of `order_api`:  
```go
type OrderOrchestrator interface {
	GetOrder(ctx context.Context, qry *core.GetOrderQry) (*core.Order, error)
	ListOrdersByStatus(ctx context.Context, qry *core.ListOrdersByStatusQry) ([]*core.Order, error)
	CreateOrder(ctx context.Context, req *core.CreateOrderCmd) error
	UpdateOrderStatus(ctx context.Context, req *core.UpdateOrderStatusCmd) error
}
```

### order_svc  

The service that handles DB access for persistence.  

`order_svc` receives Kafka messages and gRPC requests from `order_api` and execute the requested operation.  

The core of `order_svc`:  
```go
type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusFailed    Status = "failed"
)

type Order struct {
	ID        uuid.UUID
	Items     map[string]int
	Status    *Status
	CreatedAt time.Time
	UpdatedAt time.Time
}
```

The repository port of `order_svc`:  
```go
type OrderRepo interface {
	Create(ctx context.Context, order *core.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*core.Order, error)
	ListByStatus(ctx context.Context, status core.Status) ([]*core.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status core.Status) error
}
```

## Observability  

Observability uses the LGTM stack:  
- Traces are instrumented and propagated through Kafka headers and gRPC interceptors across service boundaries. They are exported to Tempo.  
- Logs are handled using a wrapper for `log/slog`. Request, trace and span IDs are added to `slog.Record` using `With()` to avoid codebase pollution. Logs are exported to Loki.  
- Tempo and Loki are provisioned in Grafana via yaml and Trace IDs are linked from Loki to Tempo.  