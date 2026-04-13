package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type lokiEntry struct {
	ctx    context.Context
	record slog.Record
	attrs  []slog.Attr
}

type LokiHandler struct {
	ch           chan lokiEntry
	handlerAttrs []slog.Attr
	endpoint     string
	client       *http.Client
	labels       map[string]string
}

func NewLokiHandler(endpoint string, labels map[string]string) *LokiHandler {
	h := &LokiHandler{
		ch:           make(chan lokiEntry, 1024),
		handlerAttrs: []slog.Attr{},
		endpoint:     endpoint,
		client:       &http.Client{Timeout: time.Second * 5},
		labels:       labels,
	}

	go h.logWorker()

	return h
}

// Implements slog.Handler
func (h *LokiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *LokiHandler) Handle(ctx context.Context, r slog.Record) error {
	select {
	case h.ch <- lokiEntry{ctx: ctx, record: r, attrs: h.handlerAttrs}:
	default:
	}
	return nil
}

func (h *LokiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LokiHandler{
		ch:           h.ch,
		handlerAttrs: append(h.handlerAttrs, attrs...),
		endpoint:     h.endpoint,
		client:       h.client,
		labels:       h.labels,
	}
}

func (h *LokiHandler) WithGroup(name string) slog.Handler {
	return h
}

// background worker
func (h *LokiHandler) logWorker() {
	var batch []lokiEntry
	ticker := time.NewTicker(2 * time.Second)
	maxBatchSize := 100
	for {
		select {
		case entry, ok := <-h.ch:
			if !ok {
				h.flush(batch)
				return
			}
			batch = append(batch, entry)
			if len(batch) >= maxBatchSize {
				h.flush(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				h.flush(batch)
				batch = nil
			}
		}
	}
}

func (h *LokiHandler) flush(entries []lokiEntry) {
	values := make([][2]string, 0)
	for _, e := range entries {
		lineMap := make(map[string]any)
		for _, a := range e.attrs {
			lineMap[a.Key] = a.Value.Resolve().Any()
		}
		e.record.Attrs(func(a slog.Attr) bool {
			lineMap[a.Key] = a.Value.Resolve().Any()
			return true
		})
		lineMap["msg"] = e.record.Message
		lineMap["level"] = e.record.Level.String()
		if span := trace.SpanFromContext(e.ctx); span.SpanContext().IsValid() {
			lineMap["trace_id"] = span.SpanContext().TraceID().String()
			lineMap["span_id"] = span.SpanContext().SpanID().String()
		}
		lineBytes, err := json.Marshal(lineMap)
		if err != nil {
			fmt.Println("failed to marshal line: ", err)
			continue
		}
		values = append(values, [2]string{
			strconv.FormatInt(e.record.Time.UnixNano(), 10),
			string(lineBytes),
		})
	}
	payload := map[string]any{
		"streams": []map[string]any{
			{
				"stream": h.labels,
				"values": values,
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("failed to marshal payload: ", err)
	}
	req, err := http.NewRequest(http.MethodPost, h.endpoint, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("failed to build Loki request: ", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		fmt.Printf("Loki flush error: %v\n", err)
		return
	}
	defer resp.Body.Close()
}
