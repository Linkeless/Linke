package persistence

import (
	"time"

	"gorm.io/gorm"
)

type SubscriptionPO struct {
	ID uint `json:"id" gorm:"primaryKey"`

	OrderID uint `json:"order_id" gorm:"not null;index"`
	UserID  uint `json:"user_id" gorm:"not null;index"`
	PlanID  uint `json:"plan_id" gorm:"not null;index"`

	UUID string `json:"uuid" gorm:"size:36;unique;index;comment:Unique identifier for server access"`

	Status string `json:"status" gorm:"size:20;not null;default:'active';index"`

	StartDate          time.Time `json:"start_date" gorm:"not null;index"`
	EndDate            time.Time `json:"end_date" gorm:"not null;index"`
	CurrentPeriodStart time.Time `json:"current_period_start" gorm:"not null;index"`
	CurrentPeriodEnd   time.Time `json:"current_period_end" gorm:"not null;index"`

	BillingCycle    string  `json:"billing_cycle" gorm:"size:20;not null"`
	BillingInterval int     `json:"billing_interval" gorm:"not null;default:1"`
	Price           float64 `json:"price" gorm:"type:decimal(10,2);not null"`
	Currency        string  `json:"currency" gorm:"size:3;not null;default:'USD'"`

	AutoRenew       bool       `json:"auto_renew" gorm:"not null;default:true"`
	NextBillingDate *time.Time `json:"next_billing_date,omitempty" gorm:"index"`

	TrialEndDate *time.Time `json:"trial_end_date,omitempty" gorm:"index"`

	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end" gorm:"not null;default:false"`
	CancellationReason string     `json:"cancellation_reason,omitempty" gorm:"type:text"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty" gorm:"index"`

	RenewalAttempts   int        `json:"renewal_attempts" gorm:"not null;default:0"`
	LastRenewalFailed *time.Time `json:"last_renewal_failed,omitempty" gorm:"index"`
	RenewalFailReason string     `json:"renewal_fail_reason,omitempty" gorm:"type:text"`

	LastUsedAt *time.Time `json:"last_used_at,omitempty" gorm:"index"`

	ServerGroupIDs string `json:"server_group_ids,omitempty" gorm:"type:json"`

	Notes    string `json:"notes,omitempty" gorm:"type:text"`
	Metadata string `json:"metadata,omitempty" gorm:"type:json"`

	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (SubscriptionPO) TableName() string {
	return "subscriptions"
}