package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Subscription represents an active service instance
type Subscription struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	OrderID uint `json:"order_id" gorm:"not null;index"`
	UserID  uint `json:"user_id" gorm:"not null;index"`
	PlanID  uint `json:"plan_id" gorm:"not null;index"`

	// Unique Identifier for Server Access
	UUID string `json:"uuid" gorm:"size:36;unique;index;comment:Unique identifier for server access"`

	// Subscription Status
	Status string `json:"status" gorm:"size:20;not null;default:'active';index"` // active, paused, cancelled, expired

	// Service Time
	StartDate          time.Time `json:"start_date" gorm:"not null;index"`
	EndDate            time.Time `json:"end_date" gorm:"not null;index"`
	CurrentPeriodStart time.Time `json:"current_period_start" gorm:"not null;index"`
	CurrentPeriodEnd   time.Time `json:"current_period_end" gorm:"not null;index"`

	// Billing Configuration
	BillingCycle    string  `json:"billing_cycle" gorm:"size:20;not null"`   // monthly, yearly, lifetime
	BillingInterval int     `json:"billing_interval" gorm:"not null;default:1"`
	Price           float64 `json:"price" gorm:"type:decimal(10,2);not null"`
	Currency        string  `json:"currency" gorm:"size:3;not null;default:'USD'"`

	// Renewal Configuration
	AutoRenew       bool       `json:"auto_renew" gorm:"not null;default:true"`
	NextBillingDate *time.Time `json:"next_billing_date,omitempty" gorm:"index"`

	// Trial Configuration
	TrialEndDate *time.Time `json:"trial_end_date,omitempty" gorm:"index"`

	// Cancellation Configuration
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end" gorm:"not null;default:false"`
	CancellationReason string     `json:"cancellation_reason,omitempty" gorm:"type:text"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty" gorm:"index"`

	// Renewal Failures
	RenewalAttempts   int        `json:"renewal_attempts" gorm:"not null;default:0"`
	LastRenewalFailed *time.Time `json:"last_renewal_failed,omitempty" gorm:"index"`
	RenewalFailReason string     `json:"renewal_fail_reason,omitempty" gorm:"type:text"`

	// Usage Records
	LastUsedAt *time.Time `json:"last_used_at,omitempty" gorm:"index"`

	// Server Group Permissions
	ServerGroupIDs string `json:"server_group_ids,omitempty" gorm:"type:json"` // JSON array

	// Business Fields
	Notes    string `json:"notes,omitempty" gorm:"type:text"`
	Metadata string `json:"metadata,omitempty" gorm:"type:json"`

	// Relationships (no foreign key constraints for performance)
	Order *Order            `json:"order,omitempty" gorm:"-"`
	User  *User             `json:"user,omitempty" gorm:"-"`
	Plan  *SubscriptionPlan `json:"plan,omitempty" gorm:"-"`

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for Subscription model
func (Subscription) TableName() string {
	return "subscriptions"
}

// BeforeCreate hook ensures UUID is generated before creating the record
func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.UUID == "" {
		s.UUID = uuid.New().String()
	}
	return nil
}

// Subscription status constants
const (
	SubscriptionStatusActive    = "active"
	SubscriptionStatusPaused    = "paused"
	SubscriptionStatusCancelled = "cancelled"
	SubscriptionStatusExpired   = "expired"
)

// Business logic methods

// IsActive checks if the subscription is currently active
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive && !s.IsExpired() && !s.IsDeleted()
}

// IsPaused checks if the subscription is paused
func (s *Subscription) IsPaused() bool {
	return s.Status == SubscriptionStatusPaused
}

// IsCancelled checks if the subscription is cancelled
func (s *Subscription) IsCancelled() bool {
	return s.Status == SubscriptionStatusCancelled
}

// IsExpired checks if the subscription has expired
func (s *Subscription) IsExpired() bool {
	return s.Status == SubscriptionStatusExpired || time.Now().After(s.EndDate)
}

// IsDeleted checks if the subscription is soft deleted
func (s *Subscription) IsDeleted() bool {
	return s.DeletedAt.Valid
}

// IsInTrial checks if the subscription is in trial period
func (s *Subscription) IsInTrial() bool {
	return s.TrialEndDate != nil && time.Now().Before(*s.TrialEndDate)
}

// DaysUntilExpiry returns the number of days until the subscription expires
func (s *Subscription) DaysUntilExpiry() int {
	duration := time.Until(s.EndDate)
	days := int(duration.Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// ShouldRenew checks if the subscription should be renewed
func (s *Subscription) ShouldRenew() bool {
	if !s.AutoRenew || s.CancelAtPeriodEnd {
		return false
	}
	return s.NextBillingDate != nil && time.Now().After(s.NextBillingDate.Add(-24*time.Hour))
}

// CanAttemptRenewal checks if renewal can be attempted (not too soon after last failure)
func (s *Subscription) CanAttemptRenewal() bool {
	if s.LastRenewalFailed == nil {
		return true
	}
	// Wait at least 1 hour after last failed attempt
	return time.Since(*s.LastRenewalFailed) > time.Hour
}

// GetRenewalDelayDuration returns the delay before next renewal attempt based on attempts count
func (s *Subscription) GetRenewalDelayDuration() time.Duration {
	switch s.RenewalAttempts {
	case 0:
		return 0 // Immediate
	case 1:
		return time.Hour // 1 hour
	case 2:
		return 6 * time.Hour // 6 hours
	default:
		return 24 * time.Hour // 24 hours (final attempt)
	}
}

// IsRenewalOverdue checks if the subscription renewal is overdue
func (s *Subscription) IsRenewalOverdue() bool {
	if s.NextBillingDate == nil {
		return false
	}
	// Consider overdue after 7 days past billing date
	return time.Since(*s.NextBillingDate) > 7*24*time.Hour
}

// Server Group Access Methods

// GetServerGroupIDs returns the list of server group IDs this subscription can access
func (s *Subscription) GetServerGroupIDs() []uint {
	if s.ServerGroupIDs == "" {
		return []uint{}
	}
	
	var groupIDs []uint
	if err := json.Unmarshal([]byte(s.ServerGroupIDs), &groupIDs); err != nil {
		// If parsing fails, return empty slice
		return []uint{}
	}
	
	return groupIDs
}

// SetServerGroupIDs sets the server group IDs that this subscription can access
func (s *Subscription) SetServerGroupIDs(groupIDs []uint) error {
	if len(groupIDs) == 0 {
		s.ServerGroupIDs = ""
		return nil
	}
	
	jsonBytes, err := json.Marshal(groupIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal server group IDs: %w", err)
	}
	
	s.ServerGroupIDs = string(jsonBytes)
	return nil
}

// HasAccessToServerGroup checks if this subscription has access to a specific server group
func (s *Subscription) HasAccessToServerGroup(groupID uint) bool {
	if s.ServerGroupIDs == "" {
		// SECURITY: Follow "deny by default" principle
		return false
	}
	
	groupIDs := s.GetServerGroupIDs()
	for _, id := range groupIDs {
		if id == 0 {
			// Group ID 0 grants access to all server groups
			return true
		}
		if id == groupID {
			return true
		}
	}
	return false
}

// SubscriptionResponse represents the subscription data structure for API responses
type SubscriptionResponse struct {
	ID      uint   `json:"id" example:"1"`
	OrderID uint   `json:"order_id" example:"1"`
	UserID  uint   `json:"user_id" example:"1"`
	PlanID  uint   `json:"plan_id" example:"1"`
	UUID    string `json:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status  string `json:"status" example:"active"`
	
	// Service Time
	StartDate          time.Time `json:"start_date" example:"2024-01-01T00:00:00Z"`
	EndDate            time.Time `json:"end_date" example:"2024-12-31T23:59:59Z"`
	CurrentPeriodStart time.Time `json:"current_period_start" example:"2024-01-01T00:00:00Z"`
	CurrentPeriodEnd   time.Time `json:"current_period_end" example:"2024-02-01T00:00:00Z"`
	
	// Billing Configuration
	BillingCycle    string  `json:"billing_cycle" example:"monthly"`
	BillingInterval int     `json:"billing_interval" example:"1"`
	Price           float64 `json:"price" example:"29.99"`
	Currency        string  `json:"currency" example:"USD"`
	
	// Renewal Configuration
	AutoRenew       bool       `json:"auto_renew" example:"true"`
	NextBillingDate *time.Time `json:"next_billing_date,omitempty" example:"2024-02-01T00:00:00Z"`
	
	// Trial Configuration
	TrialEndDate *time.Time `json:"trial_end_date,omitempty" example:"2024-01-08T00:00:00Z"`
	
	// Cancellation Configuration
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end" example:"false"`
	CancellationReason string     `json:"cancellation_reason,omitempty" example:"User request"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty" example:"2024-06-01T00:00:00Z"`
	
	// Renewal Failures
	RenewalAttempts   int        `json:"renewal_attempts" example:"0"`
	LastRenewalFailed *time.Time `json:"last_renewal_failed,omitempty" example:"2024-01-10T10:30:00Z"`
	RenewalFailReason string     `json:"renewal_fail_reason,omitempty" example:"Payment failed"`
	
	// Usage Records
	LastUsedAt *time.Time `json:"last_used_at,omitempty" example:"2024-01-15T10:30:00Z"`
	
	// Server Group Permissions
	ServerGroupIDs []uint `json:"server_group_ids,omitempty" example:"[1,2,3]"`
	
	// Business Fields
	Notes string `json:"notes,omitempty" example:"Premium customer"`
	
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	
	// Related data
	Order *OrderResponse            `json:"order,omitempty"`
	User  *UserResponse             `json:"user,omitempty"`
	Plan  *SubscriptionPlanResponse `json:"plan,omitempty"`
	
	// Computed fields
	IsInTrial    bool `json:"is_in_trial" example:"false"`
	IsExpired    bool `json:"is_expired" example:"false"`
	DaysLeft     int  `json:"days_left" example:"365"`
	ShouldRenew  bool `json:"should_renew" example:"false"`
	CanRenew     bool `json:"can_renew" example:"true"`
	IsOverdue    bool `json:"is_overdue" example:"false"`
}

// ToResponse converts NewSubscription to SubscriptionResponse
func (s *Subscription) ToResponse() *SubscriptionResponse {
	resp := &SubscriptionResponse{
		ID:                 s.ID,
		OrderID:            s.OrderID,
		UserID:             s.UserID,
		PlanID:             s.PlanID,
		UUID:               s.UUID,
		Status:             s.Status,
		StartDate:          s.StartDate,
		EndDate:            s.EndDate,
		CurrentPeriodStart: s.CurrentPeriodStart,
		CurrentPeriodEnd:   s.CurrentPeriodEnd,
		BillingCycle:       s.BillingCycle,
		BillingInterval:    s.BillingInterval,
		Price:              s.Price,
		Currency:           s.Currency,
		AutoRenew:          s.AutoRenew,
		NextBillingDate:    s.NextBillingDate,
		TrialEndDate:       s.TrialEndDate,
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		CancellationReason: s.CancellationReason,
		CancelledAt:        s.CancelledAt,
		RenewalAttempts:    s.RenewalAttempts,
		LastRenewalFailed:  s.LastRenewalFailed,
		RenewalFailReason:  s.RenewalFailReason,
		LastUsedAt:         s.LastUsedAt,
		ServerGroupIDs:     s.GetServerGroupIDs(),
		Notes:              s.Notes,
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
		
		// Computed fields
		IsInTrial:   s.IsInTrial(),
		IsExpired:   s.IsExpired(),
		DaysLeft:    s.DaysUntilExpiry(),
		ShouldRenew: s.ShouldRenew(),
		CanRenew:    s.CanAttemptRenewal(),
		IsOverdue:   s.IsRenewalOverdue(),
	}
	
	// Include related data if loaded
	if s.Order != nil {
		resp.Order = s.Order.ToResponse()
	}
	if s.User != nil {
		resp.User = s.User.ToResponse()
	}
	if s.Plan != nil {
		resp.Plan = s.Plan.ToResponse()
	}
	
	return resp
}

// ToUserResponse converts NewSubscription to a response suitable for the subscribed user
func (s *Subscription) ToUserResponse() *SubscriptionResponse {
	resp := s.ToResponse()
	
	// Remove sensitive admin information for user responses
	resp.User = nil
	
	return resp
}