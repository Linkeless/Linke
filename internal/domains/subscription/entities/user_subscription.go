package entities

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"linke/internal/shared/dto"
)

// UserSubscription represents a user's subscription to a plan
type UserSubscription struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	UserID             uint `json:"user_id" gorm:"not null;index"`
	SubscriptionPlanID uint `json:"subscription_plan_id" gorm:"not null;index"`

	// Unique Identifier for Server Access
	UUID string `json:"uuid" gorm:"size:36;unique;index;comment:Unique identifier for server access"`

	// Subscription Details
	Status       string     `json:"status" gorm:"size:20;not null;default:'active';index"` // active, paused, cancelled, expired
	StartDate    time.Time  `json:"start_date" gorm:"not null;index"`                      // 订阅开始时间
	EndDate      *time.Time `json:"end_date,omitempty" gorm:"index"`                       // 订阅结束时间 (lifetime为null)
	TrialEndDate *time.Time `json:"trial_end_date,omitempty" gorm:"index"`                 // 试用期结束时间

	// Billing Information
	CurrentPeriodStart *time.Time `json:"current_period_start,omitempty" gorm:"index"` // 当前计费周期开始
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty" gorm:"index"`   // 当前计费周期结束
	NextBillingDate    *time.Time `json:"next_billing_date,omitempty" gorm:"index"`    // 下次扣费时间

	// Pricing (订阅时的价格，可能与当前套餐价格不同)
	Price           float64 `json:"price" gorm:"type:decimal(10,2);not null"`      // 订阅价格
	Currency        string  `json:"currency" gorm:"size:3;not null;default:'USD'"` // 货币
	BillingCycle    string  `json:"billing_cycle" gorm:"size:20;not null"`         // 计费周期
	BillingInterval int     `json:"billing_interval" gorm:"not null;default:1"`    // 计费间隔

	// Cancellation Information
	CancelledAt        *time.Time `json:"cancelled_at,omitempty" gorm:"index"`                // 取消时间
	CancellationReason string     `json:"cancellation_reason,omitempty" gorm:"size:255"`      // 取消原因
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end" gorm:"not null;default:false"` // 是否在期末取消

	// Auto-renewal Information
	AutoRenew         bool       `json:"auto_renew" gorm:"not null;default:false"`      // 是否自动续费
	RenewalAttempts   int        `json:"renewal_attempts" gorm:"not null;default:0"`    // 续费尝试次数
	LastRenewalFailed *time.Time `json:"last_renewal_failed,omitempty" gorm:"index"`    // 最后一次续费失败时间
	RenewalFailReason string     `json:"renewal_fail_reason,omitempty" gorm:"size:255"` // 续费失败原因

	// Usage Tracking
	LastUsedAt *time.Time `json:"last_used_at,omitempty" gorm:"index"`   // 最后使用时间
	UsageData  string     `json:"usage_data,omitempty" gorm:"type:text"` // 使用数据(JSON)

	// Traffic Configuration and Usage
	TrafficLimit      int64      `json:"traffic_limit" gorm:"not null;default:0;comment:Total traffic limit in bytes (0 = unlimited)"`                    // 流量限制（字节）
	TrafficUsed       int64      `json:"traffic_used" gorm:"not null;default:0;comment:Total traffic used in bytes"`                                      // 已使用流量（字节）
	TrafficResetDate  *time.Time `json:"traffic_reset_date,omitempty" gorm:"index;comment:Next traffic reset date"`                                       // 流量重置日期
	TrafficResetCycle string     `json:"traffic_reset_cycle" gorm:"size:20;not null;default:'monthly';comment:Traffic reset cycle"`                       // 流量重置周期 (monthly, never)
	TrafficSuspended  bool       `json:"traffic_suspended" gorm:"not null;default:false;index;comment:Whether account is suspended due to traffic limit"` // 是否因流量超限暂停

	// Server Group Access
	ServerGroupIDs string `json:"server_group_ids,omitempty" gorm:"type:text"` // 可访问的服务器组ID列表(JSON)

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

// TableName returns the table name for UserSubscription model
func (UserSubscription) TableName() string {
	return "user_subscriptions"
}

// BeforeCreate hook ensures UUID is generated before creating the record
func (us *UserSubscription) BeforeCreate(tx *gorm.DB) error {
	if us.UUID == "" {
		us.UUID = uuid.New().String()
	}
	return nil
}

// Status constants
const (
	UserSubscriptionStatusActive    = "active"
	UserSubscriptionStatusPaused    = "paused"
	UserSubscriptionStatusCancelled = "cancelled"
	UserSubscriptionStatusExpired   = "expired"
	UserSubscriptionStatusTrial     = "trial"
)

// Traffic reset cycle constants
const (
	TrafficResetCycleMonthly = "monthly"
	TrafficResetCycleNever   = "never"
)

// Traffic limit constants
const (
	TrafficUnlimited = int64(0) // 0 means unlimited traffic
)

// IsActive checks if the subscription is currently active
func (us *UserSubscription) IsActive() bool {
	if us.Status != UserSubscriptionStatusActive {
		return false
	}

	// Check if subscription has expired
	if us.EndDate != nil && us.EndDate.Before(time.Now()) {
		return false
	}

	// Check if account is suspended due to traffic limit
	if us.TrafficSuspended {
		return false
	}

	return !us.IsDeleted()
}

// IsInTrial checks if the subscription is in trial period
func (us *UserSubscription) IsInTrial() bool {
	if us.TrialEndDate == nil {
		return false
	}
	return us.TrialEndDate.After(time.Now())
}

// IsExpired checks if the subscription has expired
func (us *UserSubscription) IsExpired() bool {
	if us.EndDate == nil {
		return false // Lifetime subscription
	}
	return us.EndDate.Before(time.Now())
}

// IsCancelled checks if the subscription is cancelled
func (us *UserSubscription) IsCancelled() bool {
	return us.Status == UserSubscriptionStatusCancelled
}

// IsDeleted checks if the subscription is soft deleted
func (us *UserSubscription) IsDeleted() bool {
	return us.DeletedAt.Valid
}

// DaysUntilExpiry returns the number of days until the subscription expires
func (us *UserSubscription) DaysUntilExpiry() int {
	if us.EndDate == nil {
		return -1 // Lifetime subscription
	}

	duration := us.EndDate.Sub(time.Now())
	return int(duration.Hours() / 24)
}

// ShouldRenew checks if the subscription should be renewed
func (us *UserSubscription) ShouldRenew() bool {
	return us.IsActive() && !us.CancelAtPeriodEnd && us.NextBillingDate != nil && us.NextBillingDate.Before(time.Now().Add(24*time.Hour))
}

// ShouldAutoRenew checks if the subscription should be auto-renewed
func (us *UserSubscription) ShouldAutoRenew() bool {
	return us.AutoRenew && us.ShouldRenew() && us.RenewalAttempts < 3 // Max 3 attempts
}

// CanAttemptRenewal checks if renewal can be attempted (not too soon after last failure)
func (us *UserSubscription) CanAttemptRenewal() bool {
	if us.LastRenewalFailed == nil {
		return true
	}
	// Wait at least 1 hour after last failed attempt
	return time.Since(*us.LastRenewalFailed) > time.Hour
}

// GetRenewalDelayDuration returns the delay before next renewal attempt based on attempts count
func (us *UserSubscription) GetRenewalDelayDuration() time.Duration {
	switch us.RenewalAttempts {
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
func (us *UserSubscription) IsRenewalOverdue() bool {
	if us.NextBillingDate == nil {
		return false
	}
	// Consider overdue after 7 days past billing date
	return time.Since(*us.NextBillingDate) > 7*24*time.Hour
}

// HasTrafficLimit checks if the subscription has a traffic limit
func (us *UserSubscription) HasTrafficLimit() bool {
	return us.TrafficLimit > TrafficUnlimited
}

// IsTrafficExceeded checks if the subscription has exceeded its traffic limit
func (us *UserSubscription) IsTrafficExceeded() bool {
	if !us.HasTrafficLimit() {
		return false
	}
	return us.TrafficUsed >= us.TrafficLimit
}

// GetRemainingTraffic returns the remaining traffic in bytes
func (us *UserSubscription) GetRemainingTraffic() int64 {
	if !us.HasTrafficLimit() {
		return -1 // Unlimited
	}
	remaining := us.TrafficLimit - us.TrafficUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetTrafficUsagePercentage returns the traffic usage percentage (0-100)
func (us *UserSubscription) GetTrafficUsagePercentage() float64 {
	if !us.HasTrafficLimit() {
		return 0.0
	}
	if us.TrafficLimit == 0 {
		return 0.0
	}
	percentage := float64(us.TrafficUsed) / float64(us.TrafficLimit) * 100
	if percentage > 100 {
		return 100.0
	}
	return percentage
}

// ShouldResetTraffic checks if the traffic should be reset based on reset cycle
func (us *UserSubscription) ShouldResetTraffic() bool {
	if us.TrafficResetCycle == TrafficResetCycleNever {
		return false
	}
	if us.TrafficResetDate == nil {
		return false
	}
	return us.TrafficResetDate.Before(time.Now())
}

// AddTrafficUsage adds traffic usage to the subscription and checks limits
func (us *UserSubscription) AddTrafficUsage(bytes int64) bool {
	us.TrafficUsed += bytes

	// Check if traffic limit is exceeded
	if us.HasTrafficLimit() && us.IsTrafficExceeded() {
		us.TrafficSuspended = true
		return true // Indicates that limit was exceeded
	}

	return false
}

// UserSubscriptionResponse represents the user subscription data structure for API responses
type UserSubscriptionResponse struct {
	ID                 uint       `json:"id" example:"1"`                                                // Subscription ID
	UserID             uint       `json:"user_id" example:"1"`                                           // User ID
	SubscriptionPlanID uint       `json:"subscription_plan_id" example:"1"`                              // Plan ID
	UUID               string     `json:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`           // Unique identifier
	Status             string     `json:"status" example:"active"`                                       // Status
	StartDate          time.Time  `json:"start_date" example:"2024-01-01T00:00:00Z"`                     // Start date
	EndDate            *time.Time `json:"end_date,omitempty" example:"2024-12-31T23:59:59Z"`             // End date
	TrialEndDate       *time.Time `json:"trial_end_date,omitempty" example:"2024-01-08T00:00:00Z"`       // Trial end
	CurrentPeriodStart *time.Time `json:"current_period_start,omitempty" example:"2024-01-01T00:00:00Z"` // Current period start
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty" example:"2024-02-01T00:00:00Z"`   // Current period end
	NextBillingDate    *time.Time `json:"next_billing_date,omitempty" example:"2024-02-01T00:00:00Z"`    // Next billing
	Price              float64    `json:"price" example:"29.99"`                                         // Price
	Currency           string     `json:"currency" example:"USD"`                                        // Currency
	BillingCycle       string     `json:"billing_cycle" example:"monthly"`                               // Billing cycle
	BillingInterval    int        `json:"billing_interval" example:"1"`                                  // Billing interval
	CancelledAt        *time.Time `json:"cancelled_at,omitempty" example:"2024-06-01T00:00:00Z"`         // Cancelled date
	CancellationReason string     `json:"cancellation_reason,omitempty" example:"User request"`          // Cancellation reason
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end" example:"false"`                          // Cancel at period end
	AutoRenew          bool       `json:"auto_renew" example:"true"`                                     // Auto renewal enabled
	RenewalAttempts    int        `json:"renewal_attempts" example:"0"`                                  // Renewal attempts count
	LastRenewalFailed  *time.Time `json:"last_renewal_failed,omitempty" example:"2024-01-10T10:30:00Z"`  // Last renewal failure
	RenewalFailReason  string     `json:"renewal_fail_reason,omitempty" example:"Payment failed"`        // Renewal failure reason
	LastUsedAt         *time.Time `json:"last_used_at,omitempty" example:"2024-01-15T10:30:00Z"`         // Last used
	CreatedAt          time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`                     // Creation time
	UpdatedAt          time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`                     // Update time

	// Related data
	User             *dto.UserBasicDTO         `json:"user,omitempty"`              // User info
	SubscriptionPlan *SubscriptionPlanResponse `json:"subscription_plan,omitempty"` // Plan info

	// Computed fields
	IsInTrial bool `json:"is_in_trial"` // Trial status
	IsExpired bool `json:"is_expired"`  // Expiry status
	DaysLeft  int  `json:"days_left"`   // Days until expiry (-1 for lifetime)
}

// ToResponse converts UserSubscription to UserSubscriptionResponse
func (us *UserSubscription) ToResponse() *UserSubscriptionResponse {
	resp := &UserSubscriptionResponse{
		ID:                 us.ID,
		UserID:             us.UserID,
		SubscriptionPlanID: us.SubscriptionPlanID,
		UUID:               us.UUID,
		Status:             us.Status,
		StartDate:          us.StartDate,
		EndDate:            us.EndDate,
		TrialEndDate:       us.TrialEndDate,
		CurrentPeriodStart: us.CurrentPeriodStart,
		CurrentPeriodEnd:   us.CurrentPeriodEnd,
		NextBillingDate:    us.NextBillingDate,
		Price:              us.Price,
		Currency:           us.Currency,
		BillingCycle:       us.BillingCycle,
		BillingInterval:    us.BillingInterval,
		CancelledAt:        us.CancelledAt,
		CancellationReason: us.CancellationReason,
		CancelAtPeriodEnd:  us.CancelAtPeriodEnd,
		AutoRenew:          us.AutoRenew,
		RenewalAttempts:    us.RenewalAttempts,
		LastRenewalFailed:  us.LastRenewalFailed,
		RenewalFailReason:  us.RenewalFailReason,
		LastUsedAt:         us.LastUsedAt,
		CreatedAt:          us.CreatedAt,
		UpdatedAt:          us.UpdatedAt,

		// Computed fields
		IsInTrial: us.IsInTrial(),
		IsExpired: us.IsExpired(),
		DaysLeft:  us.DaysUntilExpiry(),
	}

	// Note: Related data should be populated at the application layer
	// to avoid cross-domain dependencies

	return resp
}

// ToUserResponse converts UserSubscription to a response suitable for the subscribed user
func (us *UserSubscription) ToUserResponse() *UserSubscriptionResponse {
	resp := us.ToResponse()

	// Include plan info but not user info (since user already knows their own info)
	resp.User = nil

	return resp
}

// Server Group Management Methods

// GetServerGroupIDs returns the list of server group IDs this subscription can access
func (us *UserSubscription) GetServerGroupIDs() []uint {
	if us.ServerGroupIDs == "" {
		return []uint{}
	}

	var groupIDs []uint
	if err := json.Unmarshal([]byte(us.ServerGroupIDs), &groupIDs); err != nil {
		// If parsing fails, return empty slice
		return []uint{}
	}

	return groupIDs
}

// SetServerGroupIDs sets the server group IDs that this subscription can access
func (us *UserSubscription) SetServerGroupIDs(groupIDs []uint) error {
	if len(groupIDs) == 0 {
		us.ServerGroupIDs = ""
		return nil
	}

	jsonBytes, err := json.Marshal(groupIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal server group IDs: %w", err)
	}

	us.ServerGroupIDs = string(jsonBytes)
	return nil
}

// HasAccessToServerGroup checks if this subscription has access to a specific server group
// SECURITY: Follows "deny by default" principle from CLAUDE.md:
// - Empty server_group_ids means no access (secure default)
// - [0] explicitly grants access to all server groups
// - Specific IDs grant access to those groups only
func (us *UserSubscription) HasAccessToServerGroup(groupID uint) bool {
	if us.ServerGroupIDs == "" {
		// SECURITY: Follow "deny by default" principle
		// Empty server_group_ids means no access to any servers
		return false
	}

	groupIDs := us.GetServerGroupIDs()
	for _, id := range groupIDs {
		if id == 0 {
			// SECURITY: Group ID 0 grants explicit access to all server groups
			// This follows the documented behavior in CLAUDE.md
			return true
		}
		if id == groupID {
			return true
		}
	}
	return false
}
