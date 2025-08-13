package constants

// User Subscription Status Constants
const (
	UserSubscriptionStatusActive    = "active"
	UserSubscriptionStatusPaused    = "paused"
	UserSubscriptionStatusCancelled = "cancelled"
	UserSubscriptionStatusExpired   = "expired"
	UserSubscriptionStatusTrial     = "trial"
)

// Traffic Reset Cycle Constants
const (
	TrafficResetCycleMonthly = "monthly"
	TrafficResetCycleNever   = "never"
)

// Traffic Limit Constants
const (
	TrafficUnlimited = int64(0) // 0 means unlimited traffic
)

// Billing Cycle Constants
const (
	BillingCycleMonthly  = "monthly"
	BillingCycleYearly   = "yearly"
	BillingCycleLifetime = "lifetime"
)

// Subscription Plan Status Constants
const (
	SubscriptionPlanStatusActive   = "active"
	SubscriptionPlanStatusInactive = "inactive"
	SubscriptionPlanStatusArchived = "archived"
)
