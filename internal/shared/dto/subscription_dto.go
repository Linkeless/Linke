package dto

import "time"

// SubscriptionPlanBasicDTO represents basic subscription plan information
type SubscriptionPlanBasicDTO struct {
	ID           uint    `json:"id" example:"1"`
	Name         string  `json:"name" example:"Premium Plan"`
	Description  string  `json:"description" example:"Premium subscription with advanced features"`
	Price        float64 `json:"price" example:"29.99"`
	Currency     string  `json:"currency" example:"CNY"`
	Duration     int     `json:"duration" example:"30"`
	DurationType string  `json:"duration_type" example:"days"`
	Status       string  `json:"status" example:"active"`
}

// SubscriptionOrderBasicDTO represents basic subscription order information for cross-domain references
type SubscriptionOrderBasicDTO struct {
	ID                 uint       `json:"id" example:"1"`
	UserID             uint       `json:"user_id" example:"1"`
	SubscriptionPlanID uint       `json:"subscription_plan_id" example:"1"`
	OrderNumber        string     `json:"order_number" example:"ORD-2024-001"`
	OrderType          string     `json:"order_type" example:"new"`
	Status             string     `json:"status" example:"paid"`
	Amount             float64    `json:"amount" example:"29.99"`
	Currency           string     `json:"currency" example:"CNY"`
	TotalAmount        float64    `json:"total_amount" example:"29.99"`
	PaymentMethod      string     `json:"payment_method,omitempty" example:"credit_card"`
	PaymentGateway     string     `json:"payment_gateway,omitempty" example:"stripe"`
	PaidAt             *time.Time `json:"paid_at,omitempty" example:"2024-01-01T10:30:00Z"`
	CreatedAt          time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt          time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// SubscriptionOrderSummaryDTO represents minimal subscription order information for listings
type SubscriptionOrderSummaryDTO struct {
	ID          uint      `json:"id" example:"1"`
	OrderNumber string    `json:"order_number" example:"ORD-2024-001"`
	OrderType   string    `json:"order_type" example:"new"`
	Status      string    `json:"status" example:"paid"`
	TotalAmount float64   `json:"total_amount" example:"29.99"`
	Currency    string    `json:"currency" example:"CNY"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// UserSubscriptionBasicDTO represents basic user subscription information
type UserSubscriptionBasicDTO struct {
	ID                 uint      `json:"id" example:"1"`
	UserID             uint      `json:"user_id" example:"1"`
	SubscriptionPlanID uint      `json:"subscription_plan_id" example:"1"`
	Status             string    `json:"status" example:"active"`
	StartedAt          time.Time `json:"started_at" example:"2024-01-01T00:00:00Z"`
	ExpiresAt          time.Time `json:"expires_at" example:"2024-02-01T00:00:00Z"`
	AutoRenew          bool      `json:"auto_renew" example:"true"`
	CreatedAt          time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt          time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}
