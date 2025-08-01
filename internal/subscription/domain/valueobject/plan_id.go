package valueobject

import (
	"errors"
	"strconv"
)

type PlanID struct {
	value uint
}

func NewPlanID(value uint) (*PlanID, error) {
	if value == 0 {
		return nil, errors.New("plan ID cannot be zero")
	}
	return &PlanID{value: value}, nil
}

func (p PlanID) Value() uint {
	return p.value
}

func (p PlanID) String() string {
	return strconv.FormatUint(uint64(p.value), 10)
}

func (p PlanID) Equals(other PlanID) bool {
	return p.value == other.value
}