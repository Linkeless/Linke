package service

import (
	"context"
	"errors"
	"time"

	"linke/internal/subscription/domain/model"
	"linke/internal/subscription/domain/repository"
	"linke/internal/subscription/domain/valueobject"
)

type SubscriptionDomainService struct {
	subscriptionRepo repository.SubscriptionRepository
}

func NewSubscriptionDomainService(subscriptionRepo repository.SubscriptionRepository) *SubscriptionDomainService {
	return &SubscriptionDomainService{
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *SubscriptionDomainService) CanUserSubscribeToPlan(
	ctx context.Context,
	userID valueobject.UserID,
	planID valueobject.PlanID,
) error {
	activeSubscriptions, err := s.subscriptionRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, subscription := range activeSubscriptions {
		if subscription.PlanID().Equals(planID) {
			return errors.New("user already has an active subscription to this plan")
		}
	}

	return nil
}

func (s *SubscriptionDomainService) CalculateNextBillingDate(
	periodEnd time.Time,
	billingCycle valueobject.BillingCycle,
	billingInterval int,
) time.Time {
	switch billingCycle {
	case valueobject.Monthly:
		return periodEnd.AddDate(0, billingInterval, 0)
	case valueobject.Yearly:
		return periodEnd.AddDate(billingInterval, 0, 0)
	default:
		return periodEnd.AddDate(0, billingInterval, 0)
	}
}

func (s *SubscriptionDomainService) CalculatePeriodEnd(
	start time.Time,
	billingCycle valueobject.BillingCycle,
	billingInterval int,
) time.Time {
	switch billingCycle {
	case valueobject.Monthly:
		return start.AddDate(0, billingInterval, 0)
	case valueobject.Yearly:
		return start.AddDate(billingInterval, 0, 0)
	case valueobject.Lifetime:
		return start.AddDate(99, 0, 0)
	default:
		return start.AddDate(0, billingInterval, 0)
	}
}

func (s *SubscriptionDomainService) IsRenewalRequired(subscription *model.Subscription) bool {
	if !subscription.AutoRenew() || subscription.CancelAtPeriodEnd() {
		return false
	}

	if subscription.BillingCycle().IsLifetime() {
		return false
	}

	nextBillingDate := subscription.NextBillingDate()
	if nextBillingDate == nil {
		return false
	}

	return time.Now().After(nextBillingDate.Add(-24 * time.Hour))
}

func (s *SubscriptionDomainService) CanAttemptRenewal(subscription *model.Subscription) bool {
	lastRenewalFailed := subscription.LastRenewalFailed()
	if lastRenewalFailed == nil {
		return true
	}

	return time.Since(*lastRenewalFailed) > time.Hour
}

func (s *SubscriptionDomainService) GetRenewalDelayDuration(renewalAttempts int) time.Duration {
	switch renewalAttempts {
	case 0:
		return 0
	case 1:
		return time.Hour
	case 2:
		return 6 * time.Hour
	default:
		return 24 * time.Hour
	}
}

func (s *SubscriptionDomainService) IsRenewalOverdue(subscription *model.Subscription) bool {
	nextBillingDate := subscription.NextBillingDate()
	if nextBillingDate == nil {
		return false
	}

	return time.Since(*nextBillingDate) > 7*24*time.Hour
}