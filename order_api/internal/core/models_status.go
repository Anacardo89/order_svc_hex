package core

type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusFailed    Status = "failed"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusPending,
		StatusConfirmed,
		StatusFailed:
		return true
	default:
		return false
	}
}
