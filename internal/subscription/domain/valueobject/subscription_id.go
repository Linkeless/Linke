package valueobject

import (
	"errors"
	"strconv"
)

type SubscriptionID struct {
	value uint
}

func NewSubscriptionID(value uint) (*SubscriptionID, error) {
	if value == 0 {
		return nil, errors.New("subscription ID cannot be zero")
	}
	return &SubscriptionID{value: value}, nil
}

func (s SubscriptionID) Value() uint {
	return s.value
}

func (s SubscriptionID) String() string {
	return strconv.FormatUint(uint64(s.value), 10)
}

func (s SubscriptionID) Equals(other SubscriptionID) bool {
	return s.value == other.value
}