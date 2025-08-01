package valueobject

import "errors"

type SubscriptionStatus string

const (
	StatusActive    SubscriptionStatus = "active"
	StatusPaused    SubscriptionStatus = "paused"
	StatusCancelled SubscriptionStatus = "cancelled"
	StatusExpired   SubscriptionStatus = "expired"
)

func NewSubscriptionStatus(value string) (*SubscriptionStatus, error) {
	status := SubscriptionStatus(value)
	if !status.IsValid() {
		return nil, errors.New("invalid subscription status")
	}
	return &status, nil
}

func (s SubscriptionStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusPaused, StatusCancelled, StatusExpired:
		return true
	default:
		return false
	}
}

func (s SubscriptionStatus) String() string {
	return string(s)
}

func (s SubscriptionStatus) IsActive() bool {
	return s == StatusActive
}

func (s SubscriptionStatus) IsPaused() bool {
	return s == StatusPaused
}

func (s SubscriptionStatus) IsCancelled() bool {
	return s == StatusCancelled
}

func (s SubscriptionStatus) IsExpired() bool {
	return s == StatusExpired
}