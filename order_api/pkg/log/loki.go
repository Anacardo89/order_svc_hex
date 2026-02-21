package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LokiHook struct {
	client *http.Client
	url    string
	labels string
}

func NewLokiHook(url string, labels map[string]string) *LokiHook {
	labelStr := "{"
	for k, v := range labels {
		labelStr += fmt.Sprintf(`%s="%s",`, k, v)
	}
	labelStr = strings.TrimRight(labelStr, ",") + "}"
	return &LokiHook{
		client: http.DefaultClient,
		url:    url + "/loki/api/v1/push",
		labels: labelStr,
	}
}

func (h *LokiHook) WriteLog(entry map[string]any) error {
	now := time.Now().UnixNano()
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"streams": []map[string]any{
			{
				"stream": map[string]string{
					"service": "order_api",
				},
				"values": [][]string{
					{fmt.Sprintf("%d", now), string(line)},
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
