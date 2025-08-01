package valueobject

import (
	"errors"
	"strconv"
)

type UserID struct {
	value uint
}

func NewUserID(value uint) (*UserID, error) {
	if value == 0 {
		return nil, errors.New("user ID cannot be zero")
	}
	return &UserID{value: value}, nil
}

func (u UserID) Value() uint {
	return u.value
}

func (u UserID) String() string {
	return strconv.FormatUint(uint64(u.value), 10)
}

func (u UserID) Equals(other UserID) bool {
	return u.value == other.value
}