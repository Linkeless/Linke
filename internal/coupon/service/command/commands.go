package command

import (
	"time"
)

// CreateCouponCommand represents a command to create a new coupon
type CreateCouponCommand struct {
	Code            string     `json:"code" validate:"required,min=3,max=50"`
	Name            string     `json:"name" validate:"required,min=1,max=100"`
	Description     string     `json:"description,omitempty" validate:"max=1000"`
	Type            string     `json:"type" validate:"required,oneof=percentage fixed_amount"`
	Value           float64    `json:"value" validate:"required,min=0"`
	MaxUses         int        `json:"max_uses,omitempty" validate:"min=0"`
	MaxUsesPerUser  int        `json:"max_uses_per_user,omitempty" validate:"min=1"`
	MinOrderAmount  float64    `json:"min_order_amount,omitempty" validate:"min=0"`
	Currency        string     `json:"currency,omitempty" validate:"omitempty,len=3"`
	ValidFrom       *time.Time `json:"valid_from,omitempty"`
	ValidUntil      *time.Time `json:"valid_until,omitempty"`
	ApplicablePlans []uint64   `json:"applicable_plans,omitempty"`
	IsPublic        bool       `json:"is_public,omitempty"`
	CreatedBy       uint64     `json:"created_by" validate:"required"`
}

// UpdateCouponCommand represents a command to update an existing coupon
type UpdateCouponCommand struct {
	CouponID        uint64     `json:"coupon_id" validate:"required"`
	Name            *string    `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description     *string    `json:"description,omitempty" validate:"omitempty,max=1000"`
	Type            *string    `json:"type,omitempty" validate:"omitempty,oneof=percentage fixed_amount"`
	Value           *float64   `json:"value,omitempty" validate:"omitempty,min=0"`
	MaxUses         *int       `json:"max_uses,omitempty" validate:"omitempty,min=0"`
	MaxUsesPerUser  *int       `json:"max_uses_per_user,omitempty" validate:"omitempty,min=1"`
	MinOrderAmount  *float64   `json:"min_order_amount,omitempty" validate:"omitempty,min=0"`
	ValidFrom       *time.Time `json:"valid_from,omitempty"`
	ValidUntil      *time.Time `json:"valid_until,omitempty"`
	ApplicablePlans []uint64   `json:"applicable_plans,omitempty"`
	Status          *string    `json:"status,omitempty" validate:"omitempty,oneof=active inactive expired"`
	IsPublic        *bool      `json:"is_public,omitempty"`
	UpdatedBy       uint64     `json:"updated_by" validate:"required"`
}

// DeleteCouponCommand represents a command to delete a coupon
type DeleteCouponCommand struct {
	CouponID  uint64 `json:"coupon_id" validate:"required"`
	DeletedBy uint64 `json:"deleted_by" validate:"required"`
}

// UseCouponCommand represents a command to use a coupon
type UseCouponCommand struct {
	CouponCode  string  `json:"coupon_code" validate:"required"`
	UserID      uint64  `json:"user_id" validate:"required"`
	OrderID     uint64  `json:"order_id" validate:"required"`
	OrderAmount float64 `json:"order_amount" validate:"required,min=0"`
	Currency    string  `json:"currency" validate:"required,len=3"`
	PlanID      uint64  `json:"plan_id" validate:"required"`
}

// ChangeCouponStatusCommand represents a command to change coupon status
type ChangeCouponStatusCommand struct {
	CouponID  uint64 `json:"coupon_id" validate:"required"`
	NewStatus string `json:"new_status" validate:"required,oneof=active inactive expired"`
	Reason    string `json:"reason,omitempty" validate:"max=500"`
	ChangedBy uint64 `json:"changed_by" validate:"required"`
}

// ExpireExpiredCouponsCommand represents a command to expire all expired coupons
type ExpireExpiredCouponsCommand struct {
	// This command doesn't need parameters as it processes all expired coupons
}