package valueobject

import (
	"fmt"
	sharedvo "linke/internal/shared/valueobject"
)

// DiscountRule encapsulates discount calculation logic
type DiscountRule struct {
	discountValue DiscountValue
	minOrderAmount *sharedvo.Money
	maxDiscountAmount *sharedvo.Money
}

// NewDiscountRule creates a new discount rule
func NewDiscountRule(
	discountValue DiscountValue, 
	minOrderAmount *sharedvo.Money, 
	maxDiscountAmount *sharedvo.Money,
) (DiscountRule, error) {
	// Validate minimum order amount
	if minOrderAmount != nil && !minOrderAmount.IsPositive() {
		return DiscountRule{}, fmt.Errorf("minimum order amount must be positive")
	}
	
	// Validate maximum discount amount
	if maxDiscountAmount != nil && !maxDiscountAmount.IsPositive() {
		return DiscountRule{}, fmt.Errorf("maximum discount amount must be positive")
	}
	
	// For percentage discounts, max discount amount makes sense
	// For fixed amount discounts, max discount amount should be >= discount value
	if discountValue.IsFixedAmount() && maxDiscountAmount != nil {
		// Convert discount value to Money for comparison
		currency := minOrderAmount.Currency()
		if minOrderAmount == nil {
			currency = maxDiscountAmount.Currency()
		}
		
		discountMoney, err := sharedvo.NewMoney(discountValue.Value(), currency)
		if err != nil {
			return DiscountRule{}, fmt.Errorf("failed to create discount money: %w", err)
		}
		
		isGreater, err := discountMoney.GreaterThan(*maxDiscountAmount)
		if err != nil {
			return DiscountRule{}, fmt.Errorf("failed to compare discount amounts: %w", err)
		}
		
		if isGreater {
			return DiscountRule{}, fmt.Errorf("fixed discount amount cannot exceed maximum discount amount")
		}
	}
	
	return DiscountRule{
		discountValue:     discountValue,
		minOrderAmount:    minOrderAmount,
		maxDiscountAmount: maxDiscountAmount,
	}, nil
}

// CalculateDiscount calculates the discount amount for a given order
func (dr DiscountRule) CalculateDiscount(orderAmount sharedvo.Money) (sharedvo.Money, error) {
	// Check minimum order amount
	if dr.minOrderAmount != nil {
		isGreaterOrEqual, err := orderAmount.GreaterThanOrEqual(*dr.minOrderAmount)
		if err != nil {
			return sharedvo.Money{}, fmt.Errorf("failed to compare order amount with minimum: %w", err)
		}
		if !isGreaterOrEqual {
			return sharedvo.Money{}, fmt.Errorf("order amount %.2f does not meet minimum requirement %.2f", 
				orderAmount.Amount(), dr.minOrderAmount.Amount())
		}
	}
	
	var discount sharedvo.Money
	var err error
	
	// Calculate base discount
	if dr.discountValue.IsPercentage() {
		discount, err = orderAmount.MultiplyByPercentage(dr.discountValue.Value())
		if err != nil {
			return sharedvo.Money{}, fmt.Errorf("failed to calculate percentage discount: %w", err)
		}
	} else {
		// Fixed amount discount
		discount, err = sharedvo.NewMoney(dr.discountValue.Value(), orderAmount.Currency())
		if err != nil {
			return sharedvo.Money{}, fmt.Errorf("failed to create fixed discount amount: %w", err)
		}
	}
	
	// Apply maximum discount limit
	if dr.maxDiscountAmount != nil {
		isGreater, err := discount.GreaterThan(*dr.maxDiscountAmount)
		if err != nil {
			return sharedvo.Money{}, fmt.Errorf("failed to compare with maximum discount: %w", err)
		}
		if isGreater {
			discount = *dr.maxDiscountAmount
		}
	}
	
	// Ensure discount doesn't exceed order amount
	isGreater, err := discount.GreaterThan(orderAmount)
	if err != nil {
		return sharedvo.Money{}, fmt.Errorf("failed to compare discount with order amount: %w", err)
	}
	if isGreater {
		discount = orderAmount
	}
	
	return discount, nil
}

// CanApplyTo checks if this discount rule can be applied to an order
func (dr DiscountRule) CanApplyTo(orderAmount sharedvo.Money) bool {
	if dr.minOrderAmount != nil {
		isGreaterOrEqual, err := orderAmount.GreaterThanOrEqual(*dr.minOrderAmount)
		if err != nil || !isGreaterOrEqual {
			return false
		}
	}
	return true
}

// DiscountValue returns the discount value
func (dr DiscountRule) DiscountValue() DiscountValue {
	return dr.discountValue
}

// MinOrderAmount returns the minimum order amount
func (dr DiscountRule) MinOrderAmount() *sharedvo.Money {
	return dr.minOrderAmount
}

// MaxDiscountAmount returns the maximum discount amount
func (dr DiscountRule) MaxDiscountAmount() *sharedvo.Money {
	return dr.maxDiscountAmount
}

// String returns string representation
func (dr DiscountRule) String() string {
	result := fmt.Sprintf("Discount: %s", dr.discountValue.String())
	
	if dr.minOrderAmount != nil {
		result += fmt.Sprintf(", Min Order: %s", dr.minOrderAmount.String())
	}
	
	if dr.maxDiscountAmount != nil {
		result += fmt.Sprintf(", Max Discount: %s", dr.maxDiscountAmount.String())
	}
	
	return result
}