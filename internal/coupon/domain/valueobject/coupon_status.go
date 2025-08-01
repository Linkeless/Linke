package valueobject

import "fmt"

// CouponStatus represents the status of a coupon
type CouponStatus string

const (
	CouponStatusActive   CouponStatus = "active"
	CouponStatusInactive CouponStatus = "inactive"
	CouponStatusExpired  CouponStatus = "expired"
)

// ValidCouponStatuses contains all valid coupon statuses
var ValidCouponStatuses = []CouponStatus{
	CouponStatusActive,
	CouponStatusInactive,
	CouponStatusExpired,
}

// NewCouponStatus creates a new coupon status with validation
func NewCouponStatus(status string) (CouponStatus, error) {
	cs := CouponStatus(status)
	if !cs.IsValid() {
		return "", fmt.Errorf("invalid coupon status: %s, must be one of: active, inactive, expired", status)
	}
	return cs, nil
}

// MustNewCouponStatus creates a new coupon status and panics on error
func MustNewCouponStatus(status string) CouponStatus {
	cs, err := NewCouponStatus(status)
	if err != nil {
		panic(err)
	}
	return cs
}

// String returns string representation
func (cs CouponStatus) String() string {
	return string(cs)
}

// IsValid checks if the coupon status is valid
func (cs CouponStatus) IsValid() bool {
	for _, validStatus := range ValidCouponStatuses {
		if cs == validStatus {
			return true
		}
	}
	return false
}

// IsActive checks if the coupon is active
func (cs CouponStatus) IsActive() bool {
	return cs == CouponStatusActive
}

// IsInactive checks if the coupon is inactive
func (cs CouponStatus) IsInactive() bool {
	return cs == CouponStatusInactive
}

// IsExpired checks if the coupon is expired
func (cs CouponStatus) IsExpired() bool {
	return cs == CouponStatusExpired
}

// CanTransitionTo checks if status can transition to another status
func (cs CouponStatus) CanTransitionTo(newStatus CouponStatus) bool {
	// Business rules for status transitions
	switch cs {
	case CouponStatusActive:
		return newStatus == CouponStatusInactive || newStatus == CouponStatusExpired
	case CouponStatusInactive:
		return newStatus == CouponStatusActive || newStatus == CouponStatusExpired
	case CouponStatusExpired:
		return false // Cannot transition from expired
	default:
		return false
	}
}

// Equals checks if two coupon statuses are equal
func (cs CouponStatus) Equals(other CouponStatus) bool {
	return cs == other
}

// MarshalJSON implements json.Marshaler
func (cs CouponStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", cs.String())), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (cs *CouponStatus) UnmarshalJSON(data []byte) error {
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	
	status, err := NewCouponStatus(str)
	if err != nil {
		return err
	}
	
	*cs = status
	return nil
}