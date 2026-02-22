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
)

type LokiHandler struct {
	handler      slog.Handler
	handlerAttrs []slog.Attr
	endpoint     string
	client       *http.Client
	labels       map[string]string
}

func NewLokiHandler(handler slog.Handler, endpoint string, labels map[string]string) *LokiHandler {
	return &LokiHandler{
		handler:      handler,
		handlerAttrs: []slog.Attr{},
		endpoint:     endpoint,
		client:       &http.Client{},
		labels:       labels,
	}
}

func (h *LokiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *LokiHandler) Handle(ctx context.Context, r slog.Record) error {
	if err := h.handler.Handle(ctx, r); err != nil {
		fmt.Println("failed to print to stdout: ", err)
	}
	entry := make(map[string]any)
	for _, a := range h.handlerAttrs {
		entry[a.Key] = a.Value.Any()
	}
	for _, a := range h.handlerAttrs {
		v := a.Value.Any()
		if err, ok := v.(error); ok {
			entry[a.Key] = err.Error()
		} else if v == nil {
			entry[a.Key] = "<nil>"
		} else {
			entry[a.Key] = v
		}
	}
	r.Attrs(func(a slog.Attr) bool {
		v := a.Value.Any()
		if err, ok := v.(error); ok {
			entry[a.Key] = err.Error()
		} else if v == nil {
			entry[a.Key] = "<nil>"
		} else {
			entry[a.Key] = v
		}
		return true
	})
	entry["level"] = r.Level.String()
	entry["msg"] = r.Message
	entry["time"] = r.Time.UTC().Format(time.RFC3339Nano)
	line, err := json.Marshal(entry)
	if err != nil {
		fmt.Println("failed to marshal line: ", err)
	}
	payload := map[string][]map[string]any{
		"streams": {{
			"stream": h.labels,
			"values": [][2]string{
				{strconv.FormatInt(time.Now().UnixNano(), 10), string(line)},
			},
		}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("failed to marshal payload: ", err)
	}
	req, err := http.NewRequest("POST", h.endpoint, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("failed to construct request: ", err)
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = h.client.Do(req)
	if err != nil {
		fmt.Println("failed to send logs to Loki:", err)
		return err
	}
	return nil
}

func (h *LokiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LokiHandler{
		handler:      h.handler.WithAttrs(attrs),
		handlerAttrs: append(h.handlerAttrs, attrs...),
		endpoint:     h.endpoint,
		client:       h.client,
		labels:       h.labels,
	}
}

func (h *LokiHandler) WithGroup(name string) slog.Handler {
	return &LokiHandler{
		handler:      h.handler.WithGroup(name),
		handlerAttrs: h.handlerAttrs,
		endpoint:     h.endpoint,
		client:       h.client,
		labels:       h.labels,
	}
}
