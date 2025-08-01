package valueobject

import (
	"fmt"
	"strings"
)

// CouponName represents a coupon's display name
type CouponName struct {
	value string
}

// NewCouponName creates a new coupon name with validation
func NewCouponName(value string) (CouponName, error) {
	value = strings.TrimSpace(value)
	
	if value == "" {
		return CouponName{}, fmt.Errorf("coupon name cannot be empty")
	}
	
	if len(value) > 100 {
		return CouponName{}, fmt.Errorf("coupon name cannot exceed 100 characters: %d", len(value))
	}
	
	if len(value) < 2 {
		return CouponName{}, fmt.Errorf("coupon name must be at least 2 characters: %d", len(value))
	}
	
	return CouponName{value: value}, nil
}

// MustNewCouponName creates a new coupon name and panics on error
func MustNewCouponName(value string) CouponName {
	name, err := NewCouponName(value)
	if err != nil {
		panic(fmt.Sprintf("failed to create coupon name: %v", err))
	}
	return name
}

// Value returns the underlying value
func (cn CouponName) Value() string {
	return cn.value
}

// String returns string representation
func (cn CouponName) String() string {
	return cn.value
}

// IsEmpty checks if the name is empty
func (cn CouponName) IsEmpty() bool {
	return strings.TrimSpace(cn.value) == ""
}

// Equals checks if two coupon names are equal
func (cn CouponName) Equals(other CouponName) bool {
	return cn.value == other.value
}

// Length returns the character count of the name
func (cn CouponName) Length() int {
	return len(cn.value)
}

// Contains checks if the name contains a substring (case-insensitive)
func (cn CouponName) Contains(substring string) bool {
	return strings.Contains(strings.ToLower(cn.value), strings.ToLower(substring))
}