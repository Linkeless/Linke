package repository

import (
	"context"
	"time"

	"linke/internal/subscription/domain/model"
	"linke/internal/subscription/domain/valueobject"
)

type SubscriptionRepository interface {
	Save(ctx context.Context, subscription *model.Subscription) error
	FindByID(ctx context.Context, id valueobject.SubscriptionID) (*model.Subscription, error)
	FindByUUID(ctx context.Context, uuid valueobject.SubscriptionUUID) (*model.Subscription, error)
	FindByUserID(ctx context.Context, userID valueobject.UserID) ([]*model.Subscription, error)
	FindActiveByUserID(ctx context.Context, userID valueobject.UserID) ([]*model.Subscription, error)
	FindExpiringBefore(ctx context.Context, before time.Time) ([]*model.Subscription, error)
	FindPendingRenewal(ctx context.Context) ([]*model.Subscription, error)
	Delete(ctx context.Context, id valueobject.SubscriptionID) error
	Count(ctx context.Context, filters *SubscriptionFilters) (int64, error)
	FindWithFilters(ctx context.Context, filters *SubscriptionFilters) ([]*model.Subscription, error)
}

type SubscriptionFilters struct {
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