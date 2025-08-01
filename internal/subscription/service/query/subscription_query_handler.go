package query

import (
	"context"
	"fmt"

	"linke/internal/subscription/domain/model"
	"linke/internal/subscription/domain/repository"
)

type SubscriptionQueryHandler struct {
	subscriptionRepo repository.SubscriptionRepository
}

func NewSubscriptionQueryHandler(subscriptionRepo repository.SubscriptionRepository) *SubscriptionQueryHandler {
	return &SubscriptionQueryHandler{
		subscriptionRepo: subscriptionRepo,
	}
}

func (h *SubscriptionQueryHandler) HandleGetSubscriptionByID(
	ctx context.Context,
	query *GetSubscriptionByIDQuery,
) (*model.Subscription, error) {
	subscription, err := h.subscriptionRepo.FindByID(ctx, query.SubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscription by ID: %w", err)
	}
	return subscription, nil
}

func (h *SubscriptionQueryHandler) HandleGetSubscriptionByUUID(
	ctx context.Context,
	query *GetSubscriptionByUUIDQuery,
) (*model.Subscription, error) {
	subscription, err := h.subscriptionRepo.FindByUUID(ctx, query.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscription by UUID: %w", err)
	}
	return subscription, nil
}

func (h *SubscriptionQueryHandler) HandleGetSubscriptionsByUserID(
	ctx context.Context,
	query *GetSubscriptionsByUserIDQuery,
) ([]*model.Subscription, error) {
	var subscriptions []*model.Subscription
	var err error

	if query.ActiveOnly {
		subscriptions, err = h.subscriptionRepo.FindActiveByUserID(ctx, query.UserID)
	} else {
		subscriptions, err = h.subscriptionRepo.FindByUserID(ctx, query.UserID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find subscriptions by user ID: %w", err)
	}

	return subscriptions, nil
}

func (h *SubscriptionQueryHandler) HandleListSubscriptions(
	ctx context.Context,
	query *ListSubscriptionsQuery,
) ([]*model.Subscription, int64, error) {
	filters := &repository.SubscriptionFilters{
		UserID:      query.UserID,
		PlanID:      query.PlanID,
		Status:      query.Status,
		Currency:    query.Currency,
		AutoRenew:   query.AutoRenew,
		InTrial:     query.InTrial,
		Expired:     query.Expired,
		Overdue:     query.Overdue,
		StartDate:   query.StartDate,
		EndDate:     query.EndDate,
		Search:      query.Search,
		SortBy:      query.SortBy,
		SortOrder:   query.SortOrder,
		Limit:       query.Limit,
		Offset:      query.Offset,
	}

	subscriptions, err := h.subscriptionRepo.FindWithFilters(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find subscriptions with filters: %w", err)
	}

	count, err := h.subscriptionRepo.Count(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	return subscriptions, count, nil
}

func (h *SubscriptionQueryHandler) HandleGetExpiringSubscriptions(
	ctx context.Context,
	query *GetExpiringSubscriptionsQuery,
) ([]*model.Subscription, error) {
	subscriptions, err := h.subscriptionRepo.FindExpiringBefore(ctx, query.Before)
	if err != nil {
		return nil, fmt.Errorf("failed to find expiring subscriptions: %w", err)
	}
	return subscriptions, nil
}

func (h *SubscriptionQueryHandler) HandleGetPendingRenewalSubscriptions(
	ctx context.Context,
	query *GetPendingRenewalSubscriptionsQuery,
) ([]*model.Subscription, error) {
	subscriptions, err := h.subscriptionRepo.FindPendingRenewal(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending renewal subscriptions: %w", err)
	}
	return subscriptions, nil
}