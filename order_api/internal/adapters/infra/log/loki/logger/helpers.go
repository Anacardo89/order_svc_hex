package logger

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
)

type CtxKey string

const (
	CtxKeyReqID  CtxKey = "request_id"
	CtxKeyLogger CtxKey = "logger"
)

func LogFromCtx(ctx context.Context, defaultLogger ports.Logger) ports.Logger {
	l, ok := ctx.Value(CtxKeyLogger).(ports.Logger)
	if !ok {
		return defaultLogger
	}
	return l
}
