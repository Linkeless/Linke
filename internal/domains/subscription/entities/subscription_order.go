package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/shared/dto"
)

// SubscriptionOrder represents an order for a subscription
type SubscriptionOrder struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	UserID             uint  `json:"user_id" gorm:"not null;index"`
	SubscriptionPlanID uint  `json:"subscription_plan_id" gorm:"not null;index"`
	UserSubscriptionID *uint `json:"user_subscription_id,omitempty" gorm:"index"` // 关联的订阅(新订阅时为null)

	// Order Information
	OrderNumber string `json:"order_number" gorm:"uniqueIndex;size:50;not null"`       // 订单号
	OrderType   string `json:"order_type" gorm:"size:20;not null;index"`               // new, renewal, upgrade, downgrade
	Status      string `json:"status" gorm:"size:20;not null;default:'pending';index"` // pending, paid, failed, cancelled, refunded

	// Pricing Details
	Amount         float64 `json:"amount" gorm:"type:decimal(10,2);not null"`           // 订单金额
	Currency       string  `json:"currency" gorm:"size:3;not null;default:'USD'"`       // 货币
	SetupFee       float64 `json:"setup_fee" gorm:"type:decimal(10,2);default:0"`       // 初装费
	DiscountAmount float64 `json:"discount_amount" gorm:"type:decimal(10,2);default:0"` // 折扣金额
	TotalAmount    float64 `json:"total_amount" gorm:"type:decimal(10,2);not null"`     // 总金额

	// Billing Period
	BillingPeriodStart *time.Time `json:"billing_period_start,omitempty" gorm:"index"` // 计费周期开始
	BillingPeriodEnd   *time.Time `json:"billing_period_end,omitempty" gorm:"index"`   // 计费周期结束

	// Payment Information
	PaymentMethod  string     `json:"payment_method,omitempty" gorm:"size:50"`        // 支付方式
	PaymentGateway string     `json:"payment_gateway,omitempty" gorm:"size:50"`       // 支付网关
	TransactionID  string     `json:"transaction_id,omitempty" gorm:"size:100;index"` // 交易ID
	PaymentStatus  string     `json:"payment_status,omitempty" gorm:"size:20;index"`  // 支付状态
	PaidAt         *time.Time `json:"paid_at,omitempty" gorm:"index"`                 // 支付时间

	// Discount & Coupon
	CouponCode    string  `json:"coupon_code,omitempty" gorm:"size:50;index"`         // 优惠券代码
	DiscountType  string  `json:"discount_type,omitempty" gorm:"size:20"`             // percentage, fixed
	DiscountValue float64 `json:"discount_value,omitempty" gorm:"type:decimal(10,2)"` // 折扣值

	// Refund Information
	RefundAmount float64    `json:"refund_amount,omitempty" gorm:"type:decimal(10,2);default:0"` // 退款金额
	RefundedAt   *time.Time `json:"refunded_at,omitempty" gorm:"index"`                          // 退款时间
	RefundReason string     `json:"refund_reason,omitempty" gorm:"size:255"`                     // 退款原因

	// Invoice Information
	InvoiceNumber string     `json:"invoice_number,omitempty" gorm:"size:50;index"` // 发票号
	InvoiceStatus string     `json:"invoice_status,omitempty" gorm:"size:20;index"` // 发票状态
	InvoicedAt    *time.Time `json:"invoiced_at,omitempty" gorm:"index"`            // 开票时间

	// Metadata
	Metadata string `json:"metadata,omitempty" gorm:"type:text"` // 额外元数据(JSON)
	Notes    string `json:"notes,omitempty" gorm:"type:text"`    // 备注

	// Note: Relationships removed to avoid cross-domain dependencies
	// Related data should be fetched and assembled at the application layer

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for SubscriptionOrder model
func (SubscriptionOrder) TableName() string {
	return "subscription_orders"
}

// Order type constants
const (
	OrderTypeNew       = "new"
	OrderTypeRenewal   = "renewal"
	OrderTypeUpgrade   = "upgrade"
	OrderTypeDowngrade = "downgrade"
)

// Order status constants
const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusFailed    = "failed"
	OrderStatusCancelled = "cancelled"
	OrderStatusRefunded  = "refunded"
)

// Payment status constants
const (
	PaymentStatusPending    = "pending"
	PaymentStatusProcessing = "processing"
	PaymentStatusCompleted  = "completed"
	PaymentStatusFailed     = "failed"
	PaymentStatusCancelled  = "cancelled"
	PaymentStatusRefunded   = "refunded"
)

// Invoice status constants
const (
	InvoiceStatusPending = "pending"
	InvoiceStatusSent    = "sent"
	InvoiceStatusPaid    = "paid"
	InvoiceStatusOverdue = "overdue"
	InvoiceStatusVoided  = "voided"
)

// Discount type constants
const (
	DiscountTypePercentage = "percentage"
	DiscountTypeFixed      = "fixed"
)

// IsPaid checks if the order is paid
func (so *SubscriptionOrder) IsPaid() bool {
	return so.Status == OrderStatusPaid
}

// IsPending checks if the order is pending
func (so *SubscriptionOrder) IsPending() bool {
	return so.Status == OrderStatusPending
}

// IsFailed checks if the order has failed
func (so *SubscriptionOrder) IsFailed() bool {
	return so.Status == OrderStatusFailed
}

// IsCancelled checks if the order is cancelled
func (so *SubscriptionOrder) IsCancelled() bool {
	return so.Status == OrderStatusCancelled
}

// IsRefunded checks if the order is refunded
func (so *SubscriptionOrder) IsRefunded() bool {
	return so.Status == OrderStatusRefunded
}

// IsDeleted checks if the order is soft deleted
func (so *SubscriptionOrder) IsDeleted() bool {
	return so.DeletedAt.Valid
}

// CanBeRefunded checks if the order can be refunded
func (so *SubscriptionOrder) CanBeRefunded() bool {
	return so.IsPaid() && so.RefundedAt == nil
}

// GetRefundableAmount returns the amount that can be refunded
func (so *SubscriptionOrder) GetRefundableAmount() float64 {
	if !so.CanBeRefunded() {
		return 0
	}
	return so.TotalAmount - so.RefundAmount
}

// SubscriptionOrderResponse represents the subscription order data structure for API responses
type SubscriptionOrderResponse struct {
	ID                 uint       `json:"id" example:"1"`                                                // Order ID
	UserID             uint       `json:"user_id" example:"1"`                                           // User ID
	SubscriptionPlanID uint       `json:"subscription_plan_id" example:"1"`                              // Plan ID
	UserSubscriptionID *uint      `json:"user_subscription_id,omitempty" example:"1"`                    // Subscription ID
	OrderNumber        string     `json:"order_number" example:"ORD-2024-001"`                           // Order number
	OrderType          string     `json:"order_type" example:"new"`                                      // Order type
	Status             string     `json:"status" example:"paid"`                                         // Status
	Amount             float64    `json:"amount" example:"29.99"`                                        // Amount
	Currency           string     `json:"currency" example:"USD"`                                        // Currency
	SetupFee           float64    `json:"setup_fee" example:"0"`                                         // Setup fee
	DiscountAmount     float64    `json:"discount_amount" example:"0"`                                   // Discount amount
	TotalAmount        float64    `json:"total_amount" example:"29.99"`                                  // Total amount
	BillingPeriodStart *time.Time `json:"billing_period_start,omitempty" example:"2024-01-01T00:00:00Z"` // Billing start
	BillingPeriodEnd   *time.Time `json:"billing_period_end,omitempty" example:"2024-02-01T00:00:00Z"`   // Billing end
	PaymentMethod      string     `json:"payment_method,omitempty" example:"credit_card"`                // Payment method
	PaymentGateway     string     `json:"payment_gateway,omitempty" example:"stripe"`                    // Payment gateway
	TransactionID      string     `json:"transaction_id,omitempty" example:"txn_123456"`                 // Transaction ID
	PaymentStatus      string     `json:"payment_status,omitempty" example:"completed"`                  // Payment status
	PaidAt             *time.Time `json:"paid_at,omitempty" example:"2024-01-01T10:30:00Z"`              // Paid time
	CouponCode         string     `json:"coupon_code,omitempty" example:"SAVE20"`                        // Coupon code
	DiscountType       string     `json:"discount_type,omitempty" example:"percentage"`                  // Discount type
	DiscountValue      float64    `json:"discount_value,omitempty" example:"20"`                         // Discount value
	RefundAmount       float64    `json:"refund_amount,omitempty" example:"0"`                           // Refund amount
	RefundedAt         *time.Time `json:"refunded_at,omitempty" example:"2024-01-15T10:30:00Z"`          // Refunded time
	RefundReason       string     `json:"refund_reason,omitempty" example:"User request"`                // Refund reason
	InvoiceNumber      string     `json:"invoice_number,omitempty" example:"INV-2024-001"`               // Invoice number
	InvoiceStatus      string     `json:"invoice_status,omitempty" example:"sent"`                       // Invoice status
	InvoicedAt         *time.Time `json:"invoiced_at,omitempty" example:"2024-01-01T10:30:00Z"`          // Invoiced time
	CreatedAt          time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`                     // Creation time
	UpdatedAt          time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`                     // Update time

	// Related data (to be populated at application layer)
	User             *dto.UserBasicDTO         `json:"user,omitempty"`              // User info
	SubscriptionPlan *SubscriptionPlanResponse `json:"subscription_plan,omitempty"` // Plan info
	UserSubscription *UserSubscriptionResponse `json:"user_subscription,omitempty"` // Subscription info
}

// ToResponse converts SubscriptionOrder to SubscriptionOrderResponse
func (so *SubscriptionOrder) ToResponse() *SubscriptionOrderResponse {
	resp := &SubscriptionOrderResponse{
		ID:                 so.ID,
		UserID:             so.UserID,
		SubscriptionPlanID: so.SubscriptionPlanID,
		UserSubscriptionID: so.UserSubscriptionID,
		OrderNumber:        so.OrderNumber,
		OrderType:          so.OrderType,
		Status:             so.Status,
		Amount:             so.Amount,
		Currency:           so.Currency,
		SetupFee:           so.SetupFee,
		DiscountAmount:     so.DiscountAmount,
		TotalAmount:        so.TotalAmount,
		BillingPeriodStart: so.BillingPeriodStart,
		BillingPeriodEnd:   so.BillingPeriodEnd,
		PaymentMethod:      so.PaymentMethod,
		PaymentGateway:     so.PaymentGateway,
		TransactionID:      so.TransactionID,
		PaymentStatus:      so.PaymentStatus,
		PaidAt:             so.PaidAt,
		CouponCode:         so.CouponCode,
		DiscountType:       so.DiscountType,
		DiscountValue:      so.DiscountValue,
		RefundAmount:       so.RefundAmount,
		RefundedAt:         so.RefundedAt,
		RefundReason:       so.RefundReason,
		InvoiceNumber:      so.InvoiceNumber,
		InvoiceStatus:      so.InvoiceStatus,
		InvoicedAt:         so.InvoicedAt,
		CreatedAt:          so.CreatedAt,
		UpdatedAt:          so.UpdatedAt,
	}

	// Note: Related data should be populated at the application layer
	// to avoid cross-domain dependencies

	return resp
}
