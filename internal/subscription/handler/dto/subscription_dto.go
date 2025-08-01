package dto

import (
	"time"
)

type SubscriptionResponse struct {
	ID                 uint       `json:"id" example:"1"`
	UUID               string     `json:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID             uint       `json:"user_id" example:"1"`
	PlanID             uint       `json:"plan_id" example:"1"`
	OrderID            uint       `json:"order_id" example:"1"`
	Status             string     `json:"status" example:"active"`
	StartDate          time.Time  `json:"start_date" example:"2024-01-01T00:00:00Z"`
	EndDate            time.Time  `json:"end_date" example:"2024-12-31T23:59:59Z"`
	CurrentPeriodStart time.Time  `json:"current_period_start" example:"2024-01-01T00:00:00Z"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end" example:"2024-02-01T00:00:00Z"`
	BillingCycle       string     `json:"billing_cycle" example:"monthly"`
	BillingInterval    int        `json:"billing_interval" example:"1"`
	Price              float64    `json:"price" example:"29.99"`
	Currency           string     `json:"currency" example:"USD"`
	AutoRenew          bool       `json:"auto_renew" example:"true"`
	NextBillingDate    *time.Time `json:"next_billing_date,omitempty" example:"2024-02-01T00:00:00Z"`
	TrialEndDate       *time.Time `json:"trial_end_date,omitempty" example:"2024-01-08T00:00:00Z"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end" example:"false"`
	CancellationReason string     `json:"cancellation_reason,omitempty" example:"User request"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty" example:"2024-06-01T00:00:00Z"`
	RenewalAttempts    int        `json:"renewal_attempts" example:"0"`
	LastRenewalFailed  *time.Time `json:"last_renewal_failed,omitempty" example:"2024-01-10T10:30:00Z"`
	RenewalFailReason  string     `json:"renewal_fail_reason,omitempty" example:"Payment failed"`
	LastUsedAt         *time.Time `json:"last_used_at,omitempty" example:"2024-01-15T10:30:00Z"`
	ServerGroupIDs     []uint     `json:"server_group_ids,omitempty" example:"[1,2,3]"`
	Notes              string     `json:"notes,omitempty" example:"Premium customer"`
	CreatedAt          time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt          time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`

	IsInTrial    bool `json:"is_in_trial" example:"false"`
	IsExpired    bool `json:"is_expired" example:"false"`
	DaysLeft     int  `json:"days_left" example:"365"`
	ShouldRenew  bool `json:"should_renew" example:"false"`
	CanRenew     bool `json:"can_renew" example:"true"`
	IsOverdue    bool `json:"is_overdue" example:"false"`
}

type CreateSubscriptionRequest struct {
	UserID             uint     `json:"user_id" binding:"required" example:"1"`
	PlanID             uint     `json:"plan_id" binding:"required" example:"1"`
	OrderID            uint     `json:"order_id" binding:"required" example:"1"`
	StartDate          string   `json:"start_date,omitempty" example:"2024-01-01T00:00:00Z"`
	EndDate            string   `json:"end_date,omitempty" example:"2024-12-31T23:59:59Z"`
	BillingCycle       string   `json:"billing_cycle" binding:"required" example:"monthly"`
	BillingInterval    int      `json:"billing_interval" binding:"required,min=1" example:"1"`
	Price              float64  `json:"price" binding:"required,min=0" example:"29.99"`
	Currency           string   `json:"currency" binding:"required" example:"USD"`
	AutoRenew          *bool    `json:"auto_renew,omitempty" example:"true"`
	TrialDays          *int     `json:"trial_days,omitempty" example:"7"`
	ServerGroupIDs     []uint   `json:"server_group_ids,omitempty"`
	Notes              string   `json:"notes,omitempty"`
}

type UpdateSubscriptionRequest struct {
	Status             *string    `json:"status,omitempty" example:"active"`
	AutoRenew          *bool      `json:"auto_renew,omitempty" example:"true"`
	CancelAtPeriodEnd  *bool      `json:"cancel_at_period_end,omitempty" example:"false"`
	ServerGroupIDs     []uint     `json:"server_group_ids,omitempty"`
	Notes              *string    `json:"notes,omitempty"`
	EndDate            *time.Time `json:"end_date,omitempty"`
}

type CancelSubscriptionRequest struct {
	Reason      string `json:"reason" binding:"required" example:"User requested cancellation"`
	Immediately bool   `json:"immediately,omitempty" example:"false"`
}

type RenewSubscriptionRequest struct {
	NewPeriodStart  time.Time  `json:"new_period_start" binding:"required" example:"2024-02-01T00:00:00Z"`
	NewPeriodEnd    time.Time  `json:"new_period_end" binding:"required" example:"2024-03-01T00:00:00Z"`
	NextBillingDate *time.Time `json:"next_billing_date,omitempty" example:"2024-03-01T00:00:00Z"`
}

type PauseSubscriptionRequest struct {
	Reason string `json:"reason" binding:"required" example:"User requested pause"`
}

type SubscriptionListRequest struct {
	UserID      *uint  `form:"user_id"`
	PlanID      *uint  `form:"plan_id"`
	Status      string `form:"status"`
	Currency    string `form:"currency"`
	AutoRenew   *bool  `form:"auto_renew"`
	InTrial     *bool  `form:"in_trial"`
	Expired     *bool  `form:"expired"`
	Overdue     *bool  `form:"overdue"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
	Search      string `form:"search"`
	SortBy      string `form:"sort_by"`
	SortOrder   string `form:"sort_order"`
	Limit       int    `form:"limit"`
	Offset      int    `form:"offset"`
}

type SubscriptionListResponse struct {
	Subscriptions []*SubscriptionResponse `json:"subscriptions"`
	Total         int64                   `json:"total"`
	Limit         int                     `json:"limit"`
	Offset        int                     `json:"offset"`
}