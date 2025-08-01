package valueobject

import "fmt"

// DiscountValue represents a discount value with validation based on type
type DiscountValue struct {
	value       float64
	couponType  CouponType
}

// NewDiscountValue creates a new discount value with validation
func NewDiscountValue(value float64, couponType CouponType) (DiscountValue, error) {
	if value < 0 {
		return DiscountValue{}, fmt.Errorf("discount value cannot be negative: %.2f", value)
	}
	
	// Validate based on coupon type
	switch couponType {
	case CouponTypePercentage:
		if value > 100 {
			return DiscountValue{}, fmt.Errorf("percentage discount cannot exceed 100%%: %.2f", value)
		}
	case CouponTypeFixedAmount:
		// No upper limit for fixed amount, but must be positive
		if value == 0 {
			return DiscountValue{}, fmt.Errorf("fixed amount discount must be greater than 0: %.2f", value)
		}
	default:
		return DiscountValue{}, fmt.Errorf("invalid coupon type: %s", couponType)
	}
	
	return DiscountValue{
		value:      value,
		couponType: couponType,
	}, nil
}

// MustNewDiscountValue creates a new discount value and panics on error
func MustNewDiscountValue(value float64, couponType CouponType) DiscountValue {
	dv, err := NewDiscountValue(value, couponType)
	if err != nil {
		panic(err)
	}
	return dv
}

// Value returns the discount value
func (dv DiscountValue) Value() float64 {
	return dv.value
}

// Type returns the coupon type
func (dv DiscountValue) Type() CouponType {
	return dv.couponType
}

// String returns string representation
func (dv DiscountValue) String() string {
	switch dv.couponType {
	case CouponTypePercentage:
		return fmt.Sprintf("%.1f%%", dv.value)
	case CouponTypeFixedAmount:
		return fmt.Sprintf("%.2f", dv.value)
	default:
		return fmt.Sprintf("%.2f", dv.value)
	}
}

// IsPercentage checks if this is a percentage discount
func (dv DiscountValue) IsPercentage() bool {
	return dv.couponType.IsPercentage()
}

// IsFixedAmount checks if this is a fixed amount discount
func (dv DiscountValue) IsFixedAmount() bool {
	return dv.couponType.IsFixedAmount()
}

// CalculateDiscount calculates the discount amount for a given order amount
func (dv DiscountValue) CalculateDiscount(orderAmount Money) (Money, error) {
	switch dv.couponType {
	case CouponTypePercentage:
		return orderAmount.MultiplyByPercentage(dv.value)
	case CouponTypeFixedAmount:
		// Create fixed discount amount with same currency as order
		fixedDiscount, err := NewMoney(dv.value, orderAmount.Currency().String())
		if err != nil {
			return Money{}, err
		}
		
		// Discount cannot exceed order amount
		return orderAmount.Min(fixedDiscount)
	default:
		return Money{}, fmt.Errorf("invalid coupon type: %s", dv.couponType)
	}
}

// Equals checks if two discount values are equal
func (dv DiscountValue) Equals(other DiscountValue) bool {
	return dv.value == other.value && dv.couponType.Equals(other.couponType)
}

// MarshalJSON implements json.Marshaler
func (dv DiscountValue) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"value":%.2f,"type":"%s"}`, dv.value, dv.couponType)), nil
}