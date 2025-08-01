package valueobject

import (
	"fmt"
	"strconv"
)

// CouponID represents a unique identifier for a coupon
type CouponID struct {
	value uint64
}

// NewCouponID creates a new coupon ID
func NewCouponID(value uint64) CouponID {
	if value == 0 {
		panic("coupon ID cannot be zero")
	}
	return CouponID{value: value}
}

// Value returns the underlying value
func (id CouponID) Value() uint64 {
	return id.value
}

// String returns string representation
func (id CouponID) String() string {
	return strconv.FormatUint(id.value, 10)
}

// Equals checks if two coupon IDs are equal
func (id CouponID) Equals(other CouponID) bool {
	return id.value == other.value
}

// IsEmpty checks if the ID is empty (zero value)
func (id CouponID) IsEmpty() bool {
	return id.value == 0
}

// MarshalJSON implements json.Marshaler
func (id CouponID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%d\"", id.value)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (id *CouponID) UnmarshalJSON(data []byte) error {
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	
	value, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid coupon ID: %w", err)
	}
	
	id.value = value
	return nil
}