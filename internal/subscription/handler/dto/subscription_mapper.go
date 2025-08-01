package dto

import (
	"linke/internal/subscription/domain/model"
	"linke/internal/subscription/domain/service"
)

type SubscriptionDTOMapper struct {
	domainService *service.SubscriptionDomainService
}

func NewSubscriptionDTOMapper(domainService *service.SubscriptionDomainService) *SubscriptionDTOMapper {
	return &SubscriptionDTOMapper{
		domainService: domainService,
	}
}

func (m *SubscriptionDTOMapper) DomainToResponse(subscription *model.Subscription) *SubscriptionResponse {
	return &SubscriptionResponse{
		ID:                 subscription.ID().Value(),
		UUID:               subscription.UUID().Value(),
		UserID:             subscription.UserID().Value(),
		PlanID:             subscription.PlanID().Value(),
		OrderID:            subscription.OrderID(),
		Status:             subscription.Status().String(),
		StartDate:          subscription.StartDate(),
		EndDate:            subscription.EndDate(),
		CurrentPeriodStart: subscription.CurrentPeriodStart(),
		CurrentPeriodEnd:   subscription.CurrentPeriodEnd(),
		BillingCycle:       subscription.BillingCycle().String(),
		BillingInterval:    subscription.BillingInterval(),
		Price:              subscription.Price().Amount(),
		Currency:           subscription.Price().Currency().String(),
		AutoRenew:          subscription.AutoRenew(),
		NextBillingDate:    subscription.NextBillingDate(),
		TrialEndDate:       subscription.TrialEndDate(),
		CancelAtPeriodEnd:  subscription.CancelAtPeriodEnd(),
		CancellationReason: subscription.CancellationReason(),
		CancelledAt:        subscription.CancelledAt(),
		RenewalAttempts:    subscription.RenewalAttempts(),
		LastRenewalFailed:  subscription.LastRenewalFailed(),
		RenewalFailReason:  subscription.RenewalFailReason(),
		LastUsedAt:         subscription.LastUsedAt(),
		ServerGroupIDs:     subscription.ServerGroupIDs(),
		Notes:              subscription.Notes(),
		CreatedAt:          subscription.CreatedAt(),
		UpdatedAt:          subscription.UpdatedAt(),

		IsInTrial:   subscription.IsInTrial(),
		IsExpired:   subscription.IsExpired(),
		DaysLeft:    subscription.DaysUntilExpiry(),
		ShouldRenew: m.domainService.IsRenewalRequired(subscription),
		CanRenew:    m.domainService.CanAttemptRenewal(subscription),
		IsOverdue:   m.domainService.IsRenewalOverdue(subscription),
	}
}

func (m *SubscriptionDTOMapper) DomainListToResponse(subscriptions []*model.Subscription) []*SubscriptionResponse {
	responses := make([]*SubscriptionResponse, len(subscriptions))
	for i, subscription := range subscriptions {
		responses[i] = m.DomainToResponse(subscription)
	}
	return responses
}

func (m *SubscriptionDTOMapper) UserResponseToResponse(subscription *model.Subscription) *SubscriptionResponse {
	response := m.DomainToResponse(subscription)
	return response
}