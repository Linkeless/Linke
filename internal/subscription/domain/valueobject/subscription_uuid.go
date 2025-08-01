package valueobject

import (
	"errors"

	"github.com/google/uuid"
)

type SubscriptionUUID struct {
	value string
}

func NewSubscriptionUUID() *SubscriptionUUID {
	return &SubscriptionUUID{value: uuid.New().String()}
}

func NewSubscriptionUUIDFromString(value string) (*SubscriptionUUID, error) {
	if value == "" {
		return nil, errors.New("UUID cannot be empty")
	}
	if _, err := uuid.Parse(value); err != nil {
		return nil, errors.New("invalid UUID format")
	}
	return &SubscriptionUUID{value: value}, nil
}

func (s SubscriptionUUID) Value() string {
	return s.value
}

func (s SubscriptionUUID) String() string {
	return s.value
}

func (s SubscriptionUUID) Equals(other SubscriptionUUID) bool {
	return s.value == other.value
}

func (s SubscriptionUUID) IsEmpty() bool {
	return s.value == ""
}