package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type LokiHook struct {
	client *http.Client
	url    string
	labels map[string]string
}

func NewLokiHook(url string, labels map[string]string) *LokiHook {
	return &LokiHook{
		client: http.DefaultClient,
		url:    url + "/loki/api/v1/push",
		labels: labels,
	}
}

func (h *LokiHook) WriteLog(entry map[string]any) error {
	now := strconv.FormatInt(time.Now().UnixNano(), 10)
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"streams": []map[string]any{
			{
				"stream": h.labels,
				"values": [][]string{
					{now, string(line)},
				},
			},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", h.url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to push to Loki: %s", string(body))
	}
	return nil
}
