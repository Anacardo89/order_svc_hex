package core

import (
	"errors"

	"github.com/Anacardo89/order_svc_hex/order_api/pkg/ptr"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusFailed    Status = "failed"
)

func MapStrToStatus(s string) (*Status, error) {
	switch s {
	case string(StatusPending):
		return ptr.Ptr(StatusPending), nil
	case string(StatusConfirmed):
		return ptr.Ptr(StatusConfirmed), nil
	case string(StatusFailed):
		return ptr.Ptr(StatusFailed), nil
	default:
		return nil, errors.New("unknown status")
	}
}
