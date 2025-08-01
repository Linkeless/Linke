package entity

import (
	"fmt"
	"time"
	
	"linke/internal/coupon/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// CouponUsageID represents a unique identifier for coupon usage
type CouponUsageID struct {
	value uint64
}

// NewCouponUsageID creates a new coupon usage ID
func NewCouponUsageID(value uint64) CouponUsageID {
	if value == 0 {
		panic("coupon usage ID cannot be zero")
	}
	return CouponUsageID{value: value}
}

// Value returns the underlying value
func (id CouponUsageID) Value() uint64 {
	return id.value
}

// String returns string representation
func (id CouponUsageID) String() string {
	return string(rune(id.value))
}

// CouponUsage represents a record of coupon usage (Entity within Coupon aggregate)
type CouponUsage struct {
	id             CouponUsageID
	couponID       valueobject.CouponID
	userID         sharedvo.UserID
	orderID        uint64
	discountAmount valueobject.Money
	orderAmount    valueobject.Money
	createdAt      time.Time
	updatedAt      time.Time
	deletedAt      *time.Time
}

// NewCouponUsage creates a new coupon usage entity
func NewCouponUsage(
	id CouponUsageID,
	couponID valueobject.CouponID,
	userID sharedvo.UserID,
	orderID uint64,
	discountAmount valueobject.Money,
	orderAmount valueobject.Money,
) (*CouponUsage, error) {
	// Validation
	if id.value == 0 {
		return nil, fmt.Errorf("coupon usage ID cannot be zero")
	}
	
	if couponID.IsEmpty() {
		return nil, fmt.Errorf("coupon ID cannot be empty")
	}
	
	if userID.IsEmpty() {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	
	if orderID == 0 {
		return nil, fmt.Errorf("order ID cannot be zero")
	}
	
	if !discountAmount.IsPositive() {
		return nil, fmt.Errorf("discount amount must be positive")
	}
	
	if !orderAmount.IsPositive() {
		return nil, fmt.Errorf("order amount must be positive")
	}
	
	// Discount amount should not exceed order amount
	if isGreater, err := discountAmount.GreaterThanOrEqual(orderAmount); err != nil {
		return nil, fmt.Errorf("failed to compare amounts: %w", err)
	} else if isGreater {
		return nil, fmt.Errorf("discount amount cannot exceed order amount")
	}
	
	// Currencies must match
	if discountAmount.Currency() != orderAmount.Currency() {
		return nil, fmt.Errorf("discount and order amounts must have the same currency")
	}
	
	now := time.Now()
	return &CouponUsage{
		id:             id,
		couponID:       couponID,
		userID:         userID,
		orderID:        orderID,
		discountAmount: discountAmount,
		orderAmount:    orderAmount,
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

// ID returns the usage ID
func (cu *CouponUsage) ID() CouponUsageID {
	return cu.id
}

// CouponID returns the coupon ID
func (cu *CouponUsage) CouponID() valueobject.CouponID {
	return cu.couponID
}

// UserID returns the user ID
func (cu *CouponUsage) UserID() sharedvo.UserID {
	return cu.userID
}

// OrderID returns the order ID
func (cu *CouponUsage) OrderID() uint64 {
	return cu.orderID
}

// DiscountAmount returns the discount amount
func (cu *CouponUsage) DiscountAmount() valueobject.Money {
	return cu.discountAmount
}

// OrderAmount returns the order amount
func (cu *CouponUsage) OrderAmount() valueobject.Money {
	return cu.orderAmount
}

// CreatedAt returns the creation time
func (cu *CouponUsage) CreatedAt() time.Time {
	return cu.createdAt
}

// UpdatedAt returns the last update time
func (cu *CouponUsage) UpdatedAt() time.Time {
	return cu.updatedAt
}

// DeletedAt returns the deletion time
func (cu *CouponUsage) DeletedAt() *time.Time {
	return cu.deletedAt
}

// IsDeleted checks if the usage record is soft deleted
func (cu *CouponUsage) IsDeleted() bool {
	return cu.deletedAt != nil
}

// SoftDelete marks the usage record as deleted
func (cu *CouponUsage) SoftDelete() {
	if cu.deletedAt == nil {
		now := time.Now()
		cu.deletedAt = &now
		cu.updatedAt = now
	}
}

// Equals checks if two usage records are equal (based on ID)
func (cu *CouponUsage) Equals(other *CouponUsage) bool {
	if other == nil {
		return false
	}
	return cu.id.value == other.id.value
}

// String returns string representation
func (cu *CouponUsage) String() string {
	return fmt.Sprintf("CouponUsage{ID:%s, CouponID:%s, UserID:%s, OrderID:%d, DiscountAmount:%s}",
		cu.id.String(), cu.couponID.String(), cu.userID.String(), cu.orderID, cu.discountAmount.String())
}