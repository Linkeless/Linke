package valueobject

import "fmt"

// CouponType represents the type of discount a coupon provides
type CouponType string

const (
	CouponTypePercentage  CouponType = "percentage"
	CouponTypeFixedAmount CouponType = "fixed_amount"
)

// ValidCouponTypes contains all valid coupon types
var ValidCouponTypes = []CouponType{
	CouponTypePercentage,
	CouponTypeFixedAmount,
}

// NewCouponType creates a new coupon type with validation
func NewCouponType(typeStr string) (CouponType, error) {
	ct := CouponType(typeStr)
	if !ct.IsValid() {
		return "", fmt.Errorf("invalid coupon type: %s, must be one of: percentage, fixed_amount", typeStr)
	}
	return ct, nil
}

// MustNewCouponType creates a new coupon type and panics on error
func MustNewCouponType(typeStr string) CouponType {
	ct, err := NewCouponType(typeStr)
	if err != nil {
		panic(err)
	}
	return ct
}

// String returns string representation
func (ct CouponType) String() string {
	return string(ct)
}

// IsValid checks if the coupon type is valid
func (ct CouponType) IsValid() bool {
	for _, validType := range ValidCouponTypes {
		if ct == validType {
			return true
		}
	}
	return false
}

// IsPercentage checks if this is a percentage discount
func (ct CouponType) IsPercentage() bool {
	return ct == CouponTypePercentage
}

// IsFixedAmount checks if this is a fixed amount discount
func (ct CouponType) IsFixedAmount() bool {
	return ct == CouponTypeFixedAmount
}

// Equals checks if two coupon types are equal
func (ct CouponType) Equals(other CouponType) bool {
	return ct == other
}

// MarshalJSON implements json.Marshaler
func (ct CouponType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", ct.String())), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (ct *CouponType) UnmarshalJSON(data []byte) error {
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	
	couponType, err := NewCouponType(str)
	if err != nil {
		return err
	}
	
	*ct = couponType
	return nil
}