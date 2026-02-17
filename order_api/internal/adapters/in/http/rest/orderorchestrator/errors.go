package orderorchestrator

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorResp struct {
	Error string `json:"error"`
}

func failHttp(w http.ResponseWriter, status int, outMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := ErrorResp{Error: outMsg}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode error response body", "error", err)
	}
}
