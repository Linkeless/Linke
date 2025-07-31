package model

import (
	"time"

	"gorm.io/gorm"
)

// Order represents an order (purchase intent and service configuration)
type Order struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	UserID uint `json:"user_id" gorm:"not null;index"`
	PlanID uint `json:"plan_id" gorm:"not null;index"`

	// Order Information
	OrderNumber string `json:"order_number" gorm:"uniqueIndex;size:32;not null"`
	OrderType   string `json:"order_type" gorm:"size:20;not null;default:'new';index"` // new, upgrade, downgrade, renewal
	Status      string `json:"status" gorm:"size:20;not null;default:'draft';index"`   // draft, confirmed, cancelled, fulfilled

	// Service Configuration
	BillingCycle    string `json:"billing_cycle" gorm:"size:20;not null"`     // monthly, yearly, lifetime
	BillingInterval int    `json:"billing_interval" gorm:"not null;default:1"` // 1 month, 3 months, etc.
	ServicePeriod   int    `json:"service_period" gorm:"not null"`             // Service period length in months

	// Pricing Information (locked at order creation)
	BaseAmount     float64 `json:"base_amount" gorm:"type:decimal(10,2);not null"`
	DiscountAmount float64 `json:"discount_amount" gorm:"type:decimal(10,2);default:0"`
	SetupFee       float64 `json:"setup_fee" gorm:"type:decimal(10,2);default:0"`
	TotalAmount    float64 `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	Currency       string  `json:"currency" gorm:"size:3;not null;default:'USD'"`

	// Coupon Information
	CouponCode     string  `json:"coupon_code,omitempty" gorm:"size:50;index"`
	CouponDiscount float64 `json:"coupon_discount" gorm:"type:decimal(10,2);default:0"`

	// Service Time
	ServiceStartDate *time.Time `json:"service_start_date,omitempty" gorm:"index"`
	ServiceEndDate   *time.Time `json:"service_end_date,omitempty" gorm:"index"`

	// Status Timestamps
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty" gorm:"index"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty" gorm:"index"`
	FulfilledAt *time.Time `json:"fulfilled_at,omitempty" gorm:"index"`

	// Business Fields
	Notes    string `json:"notes,omitempty" gorm:"type:text"`
	Metadata string `json:"metadata,omitempty" gorm:"type:json"`

	// Relationships (no foreign key constraints for performance)
	User *User             `json:"user,omitempty" gorm:"-"`
	Plan *SubscriptionPlan `json:"plan,omitempty" gorm:"-"`

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for Order model
func (Order) TableName() string {
	return "orders"
}

// Order type constants
const (
	NewOrderTypeNew       = "new"
	NewOrderTypeUpgrade   = "upgrade"
	NewOrderTypeDowngrade = "downgrade"
	NewOrderTypeRenewal   = "renewal"
)

// Order status constants
const (
	NewOrderStatusDraft     = "draft"
	NewOrderStatusConfirmed = "confirmed"
	NewOrderStatusCancelled = "cancelled"
	NewOrderStatusFulfilled = "fulfilled"
)

// Business logic methods

// IsDraft checks if the order is in draft status
func (o *Order) IsDraft() bool {
	return o.Status == NewOrderStatusDraft
}

// IsConfirmed checks if the order is confirmed
func (o *Order) IsConfirmed() bool {
	return o.Status == NewOrderStatusConfirmed
}

// IsCancelled checks if the order is cancelled
func (o *Order) IsCancelled() bool {
	return o.Status == NewOrderStatusCancelled
}

// IsFulfilled checks if the order is fulfilled
func (o *Order) IsFulfilled() bool {
	return o.Status == NewOrderStatusFulfilled
}

// CanBeConfirmed checks if the order can be confirmed
func (o *Order) CanBeConfirmed() bool {
	return o.IsDraft()
}

// CanBeCancelled checks if the order can be cancelled
func (o *Order) CanBeCancelled() bool {
	return o.IsDraft() || o.IsConfirmed()
}

// CanBeFulfilled checks if the order can be fulfilled
func (o *Order) CanBeFulfilled() bool {
	return o.IsConfirmed()
}

// IsDeleted checks if the order is soft deleted
func (o *Order) IsDeleted() bool {
	return o.DeletedAt.Valid
}

// GetFinalAmount returns the final amount after all discounts
func (o *Order) GetFinalAmount() float64 {
	return o.BaseAmount + o.SetupFee - o.DiscountAmount - o.CouponDiscount
}

// OrderResponse represents the order data structure for API responses
type OrderResponse struct {
	ID          uint       `json:"id" example:"1"`
	UserID      uint       `json:"user_id" example:"1"`
	PlanID      uint       `json:"plan_id" example:"1"`
	OrderNumber string     `json:"order_number" example:"ORD20240101001"`
	OrderType   string     `json:"order_type" example:"new"`
	Status      string     `json:"status" example:"draft"`
	
	// Service Configuration
	BillingCycle    string `json:"billing_cycle" example:"monthly"`
	BillingInterval int    `json:"billing_interval" example:"1"`
	ServicePeriod   int    `json:"service_period" example:"1"`
	
	// Pricing Information
	BaseAmount     float64 `json:"base_amount" example:"29.99"`
	DiscountAmount float64 `json:"discount_amount" example:"5.00"`
	SetupFee       float64 `json:"setup_fee" example:"0"`
	TotalAmount    float64 `json:"total_amount" example:"24.99"`
	Currency       string  `json:"currency" example:"USD"`
	
	// Coupon Information
	CouponCode     string  `json:"coupon_code,omitempty" example:"SAVE20"`
	CouponDiscount float64 `json:"coupon_discount" example:"0"`
	
	// Service Time
	ServiceStartDate *time.Time `json:"service_start_date,omitempty" example:"2024-01-01T00:00:00Z"`
	ServiceEndDate   *time.Time `json:"service_end_date,omitempty" example:"2024-02-01T00:00:00Z"`
	
	// Status Timestamps
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty" example:"2024-01-01T10:00:00Z"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
	FulfilledAt *time.Time `json:"fulfilled_at,omitempty" example:"2024-01-01T10:30:00Z"`
	
	// Business Fields
	Notes string `json:"notes,omitempty" example:"Customer requested early start"`
	
	// Timestamp Fields
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	
	// Related data
	User *UserResponse             `json:"user,omitempty"`
	Plan *SubscriptionPlanResponse `json:"plan,omitempty"`
	
	// Computed fields
	FinalAmount float64 `json:"final_amount" example:"24.99"`
	CanConfirm  bool    `json:"can_confirm" example:"true"`
	CanCancel   bool    `json:"can_cancel" example:"true"`
	CanFulfill  bool    `json:"can_fulfill" example:"false"`
}

// ToResponse converts Order to OrderResponse
func (o *Order) ToResponse() *OrderResponse {
	resp := &OrderResponse{
		ID:               o.ID,
		UserID:           o.UserID,
		PlanID:           o.PlanID,
		OrderNumber:      o.OrderNumber,
		OrderType:        o.OrderType,
		Status:           o.Status,
		BillingCycle:     o.BillingCycle,
		BillingInterval:  o.BillingInterval,
		ServicePeriod:    o.ServicePeriod,
		BaseAmount:       o.BaseAmount,
		DiscountAmount:   o.DiscountAmount,
		SetupFee:         o.SetupFee,
		TotalAmount:      o.TotalAmount,
		Currency:         o.Currency,
		CouponCode:       o.CouponCode,
		CouponDiscount:   o.CouponDiscount,
		ServiceStartDate: o.ServiceStartDate,
		ServiceEndDate:   o.ServiceEndDate,
		ConfirmedAt:      o.ConfirmedAt,
		CancelledAt:      o.CancelledAt,
		FulfilledAt:      o.FulfilledAt,
		Notes:            o.Notes,
		CreatedAt:        o.CreatedAt,
		UpdatedAt:        o.UpdatedAt,
		
		// Computed fields
		FinalAmount: o.GetFinalAmount(),
		CanConfirm:  o.CanBeConfirmed(),
		CanCancel:   o.CanBeCancelled(),
		CanFulfill:  o.CanBeFulfilled(),
	}
	
	// Include related data if loaded
	if o.User != nil {
		resp.User = o.User.ToResponse()
	}
	if o.Plan != nil {
		resp.Plan = o.Plan.ToResponse()
	}
	
	return resp
}