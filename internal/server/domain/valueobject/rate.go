package valueobject

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidRate = errors.New("rate must be greater than 0")
)

// Rate represents a server rate multiplier
type Rate struct {
	value float64
}

// NewRate creates a new Rate
func NewRate(value float64) (Rate, error) {
	if value <= 0 {
		return Rate{}, ErrInvalidRate
	}
	
	return Rate{value: value}, nil
}

// Value returns the underlying value
func (r Rate) Value() float64 {
	return r.value
}

// String returns string representation
func (r Rate) String() string {
	return fmt.Sprintf("%.2f", r.value)
}

// IsStandard checks if the rate is standard (1.0)
func (r Rate) IsStandard() bool {
	return r.value == 1.0
}

// IsPremium checks if the rate is premium (> 1.0)
func (r Rate) IsPremium() bool {
	return r.value > 1.0
}

// IsDiscounted checks if the rate is discounted (< 1.0)
func (r Rate) IsDiscounted() bool {
	return r.value < 1.0
}

// Equals checks equality with another Rate
func (r Rate) Equals(other Rate) bool {
	const epsilon = 1e-9
	return abs(r.value-other.value) < epsilon
}

// abs returns the absolute value of x
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}