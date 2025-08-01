package valueobject

import (
	"fmt"
	"regexp"
	"strings"
)

// CouponCode represents a unique coupon code
type CouponCode struct {
	value string
}

var (
	// CouponCode validation regex: alphanumeric, dash, underscore, 3-50 chars
	couponCodeRegex = regexp.MustCompile(`^[A-Z0-9_-]{3,50}$`)
)

// NewCouponCode creates a new coupon code with validation
func NewCouponCode(code string) (CouponCode, error) {
	// Normalize: trim whitespace and convert to uppercase
	normalized := strings.TrimSpace(strings.ToUpper(code))
	
	if normalized == "" {
		return CouponCode{}, fmt.Errorf("coupon code cannot be empty")
	}
	
	if !couponCodeRegex.MatchString(normalized) {
		return CouponCode{}, fmt.Errorf("invalid coupon code format: must be 3-50 characters, alphanumeric, dash or underscore only")
	}
	
	return CouponCode{value: normalized}, nil
}

// MustNewCouponCode creates a new coupon code and panics on error
func MustNewCouponCode(code string) CouponCode {
	cc, err := NewCouponCode(code)
	if err != nil {
		panic(err)
	}
	return cc
}

// Value returns the underlying string value
func (cc CouponCode) Value() string {
	return cc.value
}

// String returns string representation
func (cc CouponCode) String() string {
	return cc.value
}

// Equals checks if two coupon codes are equal
func (cc CouponCode) Equals(other CouponCode) bool {
	return cc.value == other.value
}

// IsEmpty checks if the code is empty
func (cc CouponCode) IsEmpty() bool {
	return cc.value == ""
}

// MarshalJSON implements json.Marshaler
func (cc CouponCode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", cc.value)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (cc *CouponCode) UnmarshalJSON(data []byte) error {
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	
	code, err := NewCouponCode(str)
	if err != nil {
		return err
	}
	
	*cc = code
	return nil
}