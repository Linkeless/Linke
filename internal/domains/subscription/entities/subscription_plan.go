package entities

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/subscription/constants"
)

// SubscriptionPlan represents a subscription plan/package
type SubscriptionPlan struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Core Fields
	Name        string  `json:"name" gorm:"size:100;not null;index"`                 // 套餐名称
	Code        string  `json:"code" gorm:"uniqueIndex;size:50;not null"`            // 套餐代码 (唯一标识)
	Description string  `json:"description" gorm:"type:text"`                        // 套餐描述
	Price       float64 `json:"price" gorm:"type:decimal(10,2);not null"`            // 价格
	Currency    string  `json:"currency" gorm:"size:3;not null;default:'CNY';index"` // 货币代码

	// Duration Fields
	BillingCycle    string `json:"billing_cycle" gorm:"size:20;not null;index"` // monthly, yearly, lifetime
	BillingInterval int    `json:"billing_interval" gorm:"not null;default:1"`  // 计费间隔 (1=每月, 3=季度, 12=年)
	TrialPeriodDays int    `json:"trial_period_days" gorm:"default:0"`          // 试用期天数

	// Features & Limits (JSON格式存储，便于灵活扩展)
	Features string `json:"features" gorm:"type:text"` // JSON格式的功能特性
	Limits   string `json:"limits" gorm:"type:text"`   // JSON格式的限制条件

	// Status & Visibility
	Status        string `json:"status" gorm:"size:20;not null;default:'active';index"` // active, inactive, archived
	IsVisible     bool   `json:"is_visible" gorm:"not null;default:true;index"`         // 是否在前端显示
	SortOrder     int    `json:"sort_order" gorm:"default:0;index"`                     // 排序权重
	IsPopular     bool   `json:"is_popular" gorm:"not null;default:false"`              // 是否为热门套餐
	IsRecommended bool   `json:"is_recommended" gorm:"not null;default:false"`          // 是否推荐

	// Setup & Cancellation
	SetupFee        float64 `json:"setup_fee" gorm:"type:decimal(10,2);default:0"`        // 初装费
	CancellationFee float64 `json:"cancellation_fee" gorm:"type:decimal(10,2);default:0"` // 取消费用

	// Traffic Configuration (Required for all plans)
	TrafficLimit      int64  `json:"traffic_limit" gorm:"not null;default:10737418240;comment:Traffic limit in bytes (default: 10GB)"` // 流量限制（字节）
	TrafficResetCycle string `json:"traffic_reset_cycle" gorm:"size:20;not null;default:'monthly';comment:Traffic reset cycle"`        // 流量重置周期

	// Server Group Access Configuration
	DefaultServerGroupIDs string `json:"default_server_group_ids" gorm:"column:default_server_groups;type:text;comment:Default server groups for subscriptions (JSON)"` // 默认服务器组（JSON格式）

	// Metadata
	Metadata string `json:"metadata,omitempty" gorm:"type:text"` // 额外元数据(JSON)

	// Runtime fields (not stored in database)
	DefaultServerGroupName string `json:"-" gorm:"-"` // 服务器组名称（运行时填充）

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for SubscriptionPlan model
func (SubscriptionPlan) TableName() string {
	return "subscription_plans"
}

// Currency constants moved to payment/constants/payment.go
// Traffic reset cycle constants are defined in user_subscription.go

// IsActive checks if the subscription plan is active
func (sp *SubscriptionPlan) IsActive() bool {
	return sp.Status == constants.SubscriptionPlanStatusActive && !sp.IsDeleted()
}

// IsDeleted checks if the subscription plan is soft deleted
func (sp *SubscriptionPlan) IsDeleted() bool {
	return sp.DeletedAt.Valid
}

// IsAvailableForPurchase checks if the plan is available for purchase
func (sp *SubscriptionPlan) IsAvailableForPurchase() bool {
	return sp.IsActive() && sp.IsVisible
}

// GetMonthlyPrice calculates the monthly equivalent price
func (sp *SubscriptionPlan) GetMonthlyPrice() float64 {
	if sp.BillingCycle == constants.BillingCycleLifetime {
		return 0 // Lifetime plans don't have monthly equivalent
	}
	return sp.Price / float64(sp.BillingInterval)
}

// HasTrafficLimit checks if the plan has a traffic limit configured
func (sp *SubscriptionPlan) HasTrafficLimit() bool {
	return sp.TrafficLimit > 0
}

// GetTrafficLimitGB returns the traffic limit in GB for display purposes
func (sp *SubscriptionPlan) GetTrafficLimitGB() float64 {
	if sp.TrafficLimit == 0 {
		return 0 // Unlimited
	}
	return float64(sp.TrafficLimit) / (1024 * 1024 * 1024) // Convert bytes to GB
}

// FormatTrafficLimit returns a human-readable traffic limit string
func (sp *SubscriptionPlan) FormatTrafficLimit() string {
	if sp.TrafficLimit == 0 {
		return "Unlimited"
	}

	gb := sp.GetTrafficLimitGB()
	if gb >= 1024 {
		return fmt.Sprintf("%.1f TB", gb/1024)
	}
	return fmt.Sprintf("%.1f GB", gb)
}

// IsTrafficResetEnabled checks if traffic reset is enabled for this plan
func (sp *SubscriptionPlan) IsTrafficResetEnabled() bool {
	return sp.TrafficResetCycle != constants.TrafficResetCycleNever
}

// GetDefaultServerGroupID returns the default server group ID for this plan
func (sp *SubscriptionPlan) GetDefaultServerGroupID() (uint, error) {
	if sp.DefaultServerGroupIDs == "" {
		// If no server group is configured, return 0 (no specific group)
		return 0, nil
	}

	// First try to parse as JSON array and get the first element
	var groupIDs []uint
	if err := json.Unmarshal([]byte(sp.DefaultServerGroupIDs), &groupIDs); err == nil {
		if len(groupIDs) > 0 {
			return groupIDs[0], nil // Return first group ID
		}
		return 0, nil
	}

	// If JSON array parsing fails, try to parse as a single number (legacy format)
	var singleID uint
	if err := json.Unmarshal([]byte(sp.DefaultServerGroupIDs), &singleID); err == nil {
		return singleID, nil
	}

	// If both fail, try to parse as plain integer string
	if id, parseErr := parseUint(sp.DefaultServerGroupIDs); parseErr == nil {
		return id, nil
	}

	return 0, fmt.Errorf("failed to parse default server group ID: %s", sp.DefaultServerGroupIDs)
}

// GetDefaultServerGroupIDs returns the default server group IDs for this plan (legacy method)
// Deprecated: Use GetDefaultServerGroupID() instead
func (sp *SubscriptionPlan) GetDefaultServerGroupIDs() ([]uint, error) {
	id, err := sp.GetDefaultServerGroupID()
	if err != nil {
		return nil, err
	}
	if id == 0 {
		return []uint{}, nil
	}
	return []uint{id}, nil
}

// parseUint is a helper function to parse string to uint
func parseUint(s string) (uint, error) {
	var result uint
	for _, char := range s {
		if char < '0' || char > '9' {
			return 0, fmt.Errorf("invalid character: %c", char)
		}
		result = result*10 + uint(char-'0')
	}
	return result, nil
}

// SetDefaultServerGroupIDs sets the default server group IDs for this plan
func (sp *SubscriptionPlan) SetDefaultServerGroupIDs(groupIDs []uint) error {
	data, err := json.Marshal(groupIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal server group IDs: %w", err)
	}

	sp.DefaultServerGroupIDs = string(data)
	return nil
}

// PlanLimits represents the limits configuration for a subscription plan
type PlanLimits struct {
	MaxActiveSubscriptions *int `json:"max_active_subscriptions,omitempty"` // Maximum active subscriptions per user (nil = unlimited)
	APICallsPerMonth       *int `json:"api_calls_per_month,omitempty"`      // API calls limit per month
}

// GetLimits parses and returns the limits configuration for this plan
func (sp *SubscriptionPlan) GetLimits() (*PlanLimits, error) {
	if sp.Limits == "" {
		return &PlanLimits{}, nil
	}

	var limits PlanLimits
	if err := json.Unmarshal([]byte(sp.Limits), &limits); err != nil {
		return nil, fmt.Errorf("failed to parse plan limits: %w", err)
	}

	return &limits, nil
}

// AllowsMultipleActiveSubscriptions checks if this plan allows multiple active subscriptions per user
func (sp *SubscriptionPlan) AllowsMultipleActiveSubscriptions() bool {
	limits, err := sp.GetLimits()
	if err != nil {
		// If we can't parse limits, default to allowing multiple subscriptions for backward compatibility
		return true
	}

	// If max_active_subscriptions is not set or is greater than 1, allow multiple
	return limits.MaxActiveSubscriptions == nil || *limits.MaxActiveSubscriptions != 1
}

// GetMaxActiveSubscriptions returns the maximum number of active subscriptions allowed per user
func (sp *SubscriptionPlan) GetMaxActiveSubscriptions() int {
	limits, err := sp.GetLimits()
	if err != nil {
		// Default to unlimited if we can't parse limits
		return 0 // 0 means unlimited
	}

	if limits.MaxActiveSubscriptions == nil {
		return 0 // unlimited
	}

	return *limits.MaxActiveSubscriptions
}

// SubscriptionPlanResponse represents the subscription plan data structure for API responses
type SubscriptionPlanResponse struct {
	ID              uint    `json:"id" example:"1"`                         // Plan ID
	Name            string  `json:"name" example:"Premium Plan"`            // Plan name
	Code            string  `json:"code" example:"premium-monthly"`         // Plan code
	Description     string  `json:"description" example:"Premium features"` // Plan description
	Price           float64 `json:"price" example:"29.99"`                  // Price
	Currency        string  `json:"currency" example:"CNY"`                 // Currency
	BillingCycle    string  `json:"billing_cycle" example:"monthly"`        // Billing cycle
	BillingInterval int     `json:"billing_interval" example:"1"`           // Billing interval
	TrialPeriodDays int     `json:"trial_period_days" example:"7"`          // Trial period
	Features        string  `json:"features,omitempty"`                     // Features JSON
	Limits          string  `json:"limits,omitempty"`                       // Limits JSON
	Status          string  `json:"status" example:"active"`                // Status
	IsVisible       bool    `json:"is_visible" example:"true"`              // Visibility
	SortOrder       int     `json:"sort_order" example:"1"`                 // Sort order
	IsPopular       bool    `json:"is_popular" example:"false"`             // Popular flag
	IsRecommended   bool    `json:"is_recommended" example:"true"`          // Recommended flag
	SetupFee        float64 `json:"setup_fee" example:"0"`                  // Setup fee
	CancellationFee float64 `json:"cancellation_fee" example:"0"`           // Cancellation fee

	// Traffic Configuration (Always enabled)
	TrafficLimit      int64   `json:"traffic_limit" example:"107374182400"`  // Traffic limit in bytes
	TrafficLimitGB    float64 `json:"traffic_limit_gb" example:"100"`        // Traffic limit in GB (calculated)
	TrafficLimitText  string  `json:"traffic_limit_text" example:"100.0 GB"` // Human-readable traffic limit
	TrafficResetCycle string  `json:"traffic_reset_cycle" example:"monthly"` // Traffic reset cycle

	// Server Group Configuration (Single group per plan)
	DefaultServerGroupID   uint   `json:"default_server_group_id,omitempty"`   // Default server group ID
	DefaultServerGroupName string `json:"default_server_group_name,omitempty"` // Default server group name

	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"` // Creation time
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"` // Update time
}

// ToResponse converts SubscriptionPlan to SubscriptionPlanResponse
func (sp *SubscriptionPlan) ToResponse() *SubscriptionPlanResponse {
	return &SubscriptionPlanResponse{
		ID:              sp.ID,
		Name:            sp.Name,
		Code:            sp.Code,
		Description:     sp.Description,
		Price:           sp.Price,
		Currency:        sp.Currency,
		BillingCycle:    sp.BillingCycle,
		BillingInterval: sp.BillingInterval,
		TrialPeriodDays: sp.TrialPeriodDays,
		Features:        sp.Features,
		Limits:          sp.Limits,
		Status:          sp.Status,
		IsVisible:       sp.IsVisible,
		SortOrder:       sp.SortOrder,
		IsPopular:       sp.IsPopular,
		IsRecommended:   sp.IsRecommended,
		SetupFee:        sp.SetupFee,
		CancellationFee: sp.CancellationFee,

		// Traffic Configuration
		TrafficLimit:      sp.TrafficLimit,
		TrafficLimitGB:    sp.GetTrafficLimitGB(),
		TrafficLimitText:  sp.FormatTrafficLimit(),
		TrafficResetCycle: sp.TrafficResetCycle,

		// Server Group Configuration
		DefaultServerGroupID: func() uint {
			if id, err := sp.GetDefaultServerGroupID(); err == nil {
				return id
			}
			return 0
		}(),
		DefaultServerGroupName: sp.DefaultServerGroupName, // Populated by service layer

		CreatedAt: sp.CreatedAt,
		UpdatedAt: sp.UpdatedAt,
	}
}

// ToPublicResponse converts SubscriptionPlan to a public response (for non-authenticated users)
func (sp *SubscriptionPlan) ToPublicResponse() *SubscriptionPlanResponse {
	return &SubscriptionPlanResponse{
		ID:              sp.ID,
		Name:            sp.Name,
		Code:            sp.Code,
		Description:     sp.Description,
		Price:           sp.Price,
		Currency:        sp.Currency,
		BillingCycle:    sp.BillingCycle,
		BillingInterval: sp.BillingInterval,
		TrialPeriodDays: sp.TrialPeriodDays,
		Features:        sp.Features,
		IsVisible:       sp.IsVisible,
		SortOrder:       sp.SortOrder,
		IsPopular:       sp.IsPopular,
		IsRecommended:   sp.IsRecommended,
		SetupFee:        sp.SetupFee,

		// Traffic Configuration (for public viewing)
		TrafficLimit:      sp.TrafficLimit,
		TrafficLimitGB:    sp.GetTrafficLimitGB(),
		TrafficLimitText:  sp.FormatTrafficLimit(),
		TrafficResetCycle: sp.TrafficResetCycle,

		CreatedAt: sp.CreatedAt,
		UpdatedAt: sp.UpdatedAt,
	}
}
