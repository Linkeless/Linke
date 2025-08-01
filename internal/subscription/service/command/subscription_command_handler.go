package command

import (
	"context"
	"fmt"

	"linke/internal/subscription/domain/model"
	"linke/internal/subscription/domain/repository"
	"linke/internal/subscription/domain/service"
	"linke/internal/subscription/domain/valueobject"
)

type SubscriptionCommandHandler struct {
	subscriptionRepo    repository.SubscriptionRepository
	domainService      *service.SubscriptionDomainService
	eventPublisher     EventPublisher
}

type EventPublisher interface {
	PublishDomainEvents(ctx context.Context, events []interface{}) error
}

func NewSubscriptionCommandHandler(
	subscriptionRepo repository.SubscriptionRepository,
	domainService *service.SubscriptionDomainService,
	eventPublisher EventPublisher,
) *SubscriptionCommandHandler {
	return &SubscriptionCommandHandler{
		subscriptionRepo: subscriptionRepo,
		domainService:    domainService,
		eventPublisher:   eventPublisher,
	}
}

func (h *SubscriptionCommandHandler) HandleCreateSubscription(
	ctx context.Context,
	cmd *CreateSubscriptionCommand,
) (*model.Subscription, error) {
	err := h.domainService.CanUserSubscribeToPlan(ctx, cmd.UserID, cmd.PlanID)
	if err != nil {
		return nil, fmt.Errorf("user cannot subscribe to plan: %w", err)
	}

	// Convert domain UserID to shared UserID
	sharedUserID, err := valueobject.ConvertToSharedUserID(&cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert user ID: %w", err)
	}
	
	// Convert domain Price to shared Money
	sharedMoney, err := valueobject.ConvertToSharedMoney(&cmd.Price)
	if err != nil {
		return nil, fmt.Errorf("failed to convert price: %w", err)
	}

	subscription, err := model.NewSubscription(
		&sharedUserID,
		cmd.PlanID,
		cmd.OrderID,
		cmd.StartDate,
		cmd.EndDate,
		cmd.BillingCycle,
		cmd.BillingInterval,
		sharedMoney,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	if cmd.TrialEndDate != nil {
		subscription.SetTrial(*cmd.TrialEndDate)
	}

	if len(cmd.ServerGroupIDs) > 0 {
		subscription.SetServerGroupIDs(cmd.ServerGroupIDs)
	}

	if cmd.Notes != "" {
		subscription.SetNotes(cmd.Notes)
	}

	if err := h.subscriptionRepo.Save(ctx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription: %w", err)
	}

	events := make([]interface{}, len(subscription.DomainEvents()))
	for i, event := range subscription.DomainEvents() {
		events[i] = event
	}

	if err := h.eventPublisher.PublishDomainEvents(ctx, events); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}

	subscription.ClearDomainEvents()

	return subscription, nil
}

func (h *SubscriptionCommandHandler) HandleRenewSubscription(
	ctx context.Context,
	cmd *RenewSubscriptionCommand,
) error {
	subscription, err := h.subscriptionRepo.FindByID(ctx, cmd.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	if !h.domainService.CanAttemptRenewal(subscription) {
		return fmt.Errorf("cannot attempt renewal at this time")
	}

	if err := subscription.Renew(cmd.NewPeriodStart, cmd.NewPeriodEnd, cmd.NextBillingDate); err != nil {
		return fmt.Errorf("failed to renew subscription: %w", err)
	}

	if err := h.subscriptionRepo.Save(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save renewed subscription: %w", err)
	}

	events := make([]interface{}, len(subscription.DomainEvents()))
	for i, event := range subscription.DomainEvents() {
		events[i] = event
	}

	if err := h.eventPublisher.PublishDomainEvents(ctx, events); err != nil {
		return fmt.Errorf("failed to publish domain events: %w", err)
	}

	subscription.ClearDomainEvents()

	return nil
}

func (h *SubscriptionCommandHandler) HandleCancelSubscription(
	ctx context.Context,
	cmd *CancelSubscriptionCommand,
) error {
	subscription, err := h.subscriptionRepo.FindByID(ctx, cmd.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	if err := subscription.Cancel(cmd.Reason, cmd.Immediately); err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	if err := h.subscriptionRepo.Save(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save cancelled subscription: %w", err)
	}

	events := make([]interface{}, len(subscription.DomainEvents()))
	for i, event := range subscription.DomainEvents() {
		events[i] = event
	}

	if err := h.eventPublisher.PublishDomainEvents(ctx, events); err != nil {
		return fmt.Errorf("failed to publish domain events: %w", err)
	}

	subscription.ClearDomainEvents()

	return nil
}

func (h *SubscriptionCommandHandler) HandlePauseSubscription(
	ctx context.Context,
	cmd *PauseSubscriptionCommand,
) error {
	subscription, err := h.subscriptionRepo.FindByID(ctx, cmd.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	if err := subscription.Pause(cmd.Reason); err != nil {
		return fmt.Errorf("failed to pause subscription: %w", err)
	}

	if err := h.subscriptionRepo.Save(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save paused subscription: %w", err)
	}

	return nil
}

func (h *SubscriptionCommandHandler) HandleResumeSubscription(
	ctx context.Context,
	cmd *ResumeSubscriptionCommand,
) error {
	subscription, err := h.subscriptionRepo.FindByID(ctx, cmd.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	if err := subscription.Resume(); err != nil {
		return fmt.Errorf("failed to resume subscription: %w", err)
	}

	if err := h.subscriptionRepo.Save(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save resumed subscription: %w", err)
	}

	return nil
}

func (h *SubscriptionCommandHandler) HandleUpdateSubscriptionUsage(
	ctx context.Context,
	cmd *UpdateSubscriptionUsageCommand,
) error {
	subscription, err := h.subscriptionRepo.FindByID(ctx, cmd.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	subscription.UpdateUsage()

	if err := h.subscriptionRepo.Save(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save subscription usage: %w", err)
	}

	return nil
}

func (h *SubscriptionCommandHandler) HandleUpdateSubscriptionServerGroups(
	ctx context.Context,
	cmd *UpdateSubscriptionServerGroupsCommand,
) error {
	subscription, err := h.subscriptionRepo.FindByID(ctx, cmd.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	subscription.SetServerGroupIDs(cmd.ServerGroupIDs)

	if err := h.subscriptionRepo.Save(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save subscription server groups: %w", err)
	}

	return nil
}

func (h *SubscriptionCommandHandler) HandleUpdateSubscriptionNotes(
	ctx context.Context,
	cmd *UpdateSubscriptionNotesCommand,
) error {
	subscription, err := h.subscriptionRepo.FindByID(ctx, cmd.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	subscription.SetNotes(cmd.Notes)

	if err := h.subscriptionRepo.Save(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save subscription notes: %w", err)
	}

	return nil
}

func (h *SubscriptionCommandHandler) HandleFailSubscriptionRenewal(
	ctx context.Context,
	cmd *FailSubscriptionRenewalCommand,
) error {
	subscription, err := h.subscriptionRepo.FindByID(ctx, cmd.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	subscription.FailRenewal(cmd.Reason)

	if err := h.subscriptionRepo.Save(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save subscription renewal failure: %w", err)
	}

	events := make([]interface{}, len(subscription.DomainEvents()))
	for i, event := range subscription.DomainEvents() {
		events[i] = event
	}

	if err := h.eventPublisher.PublishDomainEvents(ctx, events); err != nil {
		return fmt.Errorf("failed to publish domain events: %w", err)
	}

	subscription.ClearDomainEvents()

	return nil
}

func (h *SubscriptionCommandHandler) HandleExpireSubscription(
	ctx context.Context,
	cmd *ExpireSubscriptionCommand,
) error {
	subscription, err := h.subscriptionRepo.FindByID(ctx, cmd.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	subscription.Expire()

	if err := h.subscriptionRepo.Save(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save expired subscription: %w", err)
	}

	events := make([]interface{}, len(subscription.DomainEvents()))
	for i, event := range subscription.DomainEvents() {
		events[i] = event
	}

	if err := h.eventPublisher.PublishDomainEvents(ctx, events); err != nil {
		return fmt.Errorf("failed to publish domain events: %w", err)
	}

	subscription.ClearDomainEvents()

	return nil
}

func (h *SubscriptionCommandHandler) HandleDeleteSubscription(
	ctx context.Context,
	cmd *DeleteSubscriptionCommand,
) error {
	subscription, err := h.subscriptionRepo.FindByID(ctx, cmd.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	subscription.Delete()

	if err := h.subscriptionRepo.Save(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save deleted subscription: %w", err)
	}

	return nil
}