package query

import (
	"time"

	"linke/internal/subscription/domain/valueobject"
)

type GetSubscriptionByIDQuery struct {
	SubscriptionID valueobject.SubscriptionID
}

type GetSubscriptionByUUIDQuery struct {
	UUID valueobject.SubscriptionUUID
}

type GetSubscriptionsByUserIDQuery struct {
	UserID         valueobject.UserID
	ActiveOnly     bool
}

type ListSubscriptionsQuery struct {
	UserID      *valueobject.UserID
	PlanID      *valueobject.PlanID
	Status      *valueobject.SubscriptionStatus
	Currency    *valueobject.Currency
	AutoRenew   *bool
	InTrial     *bool
	Expired     *bool
	Overdue     *bool
	StartDate   *time.Time
	EndDate     *time.Time
	Search      string
	SortBy      string
	SortOrder   string
	Limit       int
	Offset      int
}

type GetExpiringSubscriptionsQuery struct {
	Before time.Time
}

type GetPendingRenewalSubscriptionsQuery struct {
}

type GetSubscriptionStatsQuery struct {
	UserID    *valueobject.UserID
	PlanID    *valueobject.PlanID
	StartDate *time.Time
	EndDate   *time.Time
}