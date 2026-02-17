package orderwriter

type TopicKey string

const (
	TopicOrderCreated       TopicKey = "OrderCreated"
	TopicOrderStatusUpdated TopicKey = "OrderStatusUpdated"
)

type OrderCreatedEvent struct {
	Items  map[string]int `json:"items"`
	Status string         `json:"status"`
}

type OrderStatusUpdatedEvent struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
