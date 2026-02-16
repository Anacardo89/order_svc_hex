package orderdlq

import "time"

type DlqPayload struct {
	Timestamp     time.Time `json:"timestamp"`
	Reason        string    `json:"reason"`
	Error         string    `json:"error"`
	OriginalTopic string    `json:"original_topic"`
	OriginalKey   []byte    `json:"original_key,omitempty"`
	OriginalValue []byte    `json:"original_value"`
	Partition     int32     `json:"partition,omitempty"`
	Offset        int64     `json:"offset,omitempty"`
}
