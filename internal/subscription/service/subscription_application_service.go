package service

import (
	"context"
	"fmt"
	"time"

	"linke/internal/subscription/domain/repository"
	"linke/internal/subscription/domain/service"
	"linke/internal/subscription/domain/valueobject"
	"linke/internal/subscription/handler/dto"
	"linke/internal/subscription/service/command"
	"linke/internal/subscription/service/query"
)

type SubscriptionApplicationService struct {
	commandHandler *command.SubscriptionCommandHandler
	queryHandler   *query.SubscriptionQueryHandler
	domainService  *service.SubscriptionDomainService
	dtoMapper      *dto.SubscriptionDTOMapper
}

func NewSubscriptionApplicationService(
	subscriptionRepo repository.SubscriptionRepository,
	eventPublisher command.EventPublisher,
	domainService *service.SubscriptionDomainService,
) *SubscriptionApplicationService {
	commandHandler := command.NewSubscriptionCommandHandler(subscriptionRepo, domainService, eventPublisher)
	queryHandler := query.NewSubscriptionQueryHandler(subscriptionRepo)
	dtoMapper := dto.NewSubscriptionDTOMapper(domainService)

	return &SubscriptionApplicationService{
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
		domainService:  domainService,
		dtoMapper:      dtoMapper,
	}
}

func (s *SubscriptionApplicationService) CreateSubscription(
	ctx context.Context,
	req *dto.CreateSubscriptionRequest,
) (*dto.SubscriptionResponse, error) {
	userID, err := valueobject.NewUserID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	planID, err := valueobject.NewPlanID(req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("invalid plan ID: %w", err)
	}

	billingCycle, err := valueobject.NewBillingCycle(req.BillingCycle)
	if err != nil {
		return nil, fmt.Errorf("invalid billing cycle: %w", err)
	}

	currency, err := valueobject.NewCurrency(req.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	price, err := valueobject.NewPrice(req.Price, *currency)
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}

	var startDate, endDate time.Time
	if req.StartDate != "" {
		if startDate, err = time.Parse(time.RFC3339, req.StartDate); err != nil {
			return nil, fmt.Errorf("invalid start date format: %w", err)
		}
	} else {
		startDate = time.Now()
	}

	if req.EndDate != "" {
		if endDate, err = time.Parse(time.RFC3339, req.EndDate); err != nil {
			return nil, fmt.Errorf("invalid end date format: %w", err)
		}
	} else {
		endDate = s.domainService.CalculatePeriodEnd(startDate, *billingCycle, req.BillingInterval)
	}

	cmd := &command.CreateSubscriptionCommand{
		UserID:          *userID,
		PlanID:          *planID,
		OrderID:         req.OrderID,
		StartDate:       startDate,
		EndDate:         endDate,
		BillingCycle:    *billingCycle,
		BillingInterval: req.BillingInterval,
		Price:           *price,
		AutoRenew:       req.AutoRenew != nil && *req.AutoRenew,
		ServerGroupIDs:  req.ServerGroupIDs,
		Notes:           req.Notes,
	}

	if req.TrialDays != nil && *req.TrialDays > 0 {
		trialEndDate := startDate.AddDate(0, 0, *req.TrialDays)
		cmd.TrialEndDate = &trialEndDate
	}

	subscription, err := s.commandHandler.HandleCreateSubscription(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return s.dtoMapper.DomainToResponse(subscription), nil
}

func (s *SubscriptionApplicationService) GetSubscriptionByID(
	ctx context.Context,
	subscriptionID uint,
) (*dto.SubscriptionResponse, error) {
	id, err := valueobject.NewSubscriptionID(subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription ID: %w", err)
	}

	query := &query.GetSubscriptionByIDQuery{SubscriptionID: *id}
	subscription, err := s.queryHandler.HandleGetSubscriptionByID(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return s.dtoMapper.DomainToResponse(subscription), nil
}

func (s *SubscriptionApplicationService) GetSubscriptionByUUID(
	ctx context.Context,
	uuid string,
) (*dto.SubscriptionResponse, error) {
	subscriptionUUID, err := valueobject.NewSubscriptionUUIDFromString(uuid)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription UUID: %w", err)
	}

	query := &query.GetSubscriptionByUUIDQuery{UUID: *subscriptionUUID}
	subscription, err := s.queryHandler.HandleGetSubscriptionByUUID(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return s.dtoMapper.DomainToResponse(subscription), nil
}

func (s *SubscriptionApplicationService) GetSubscriptionsByUserID(
	ctx context.Context,
	userID uint,
	activeOnly bool,
) ([]*dto.SubscriptionResponse, error) {
	id, err := valueobject.NewUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	query := &query.GetSubscriptionsByUserIDQuery{
		UserID:     *id,
		ActiveOnly: activeOnly,
	}

	subscriptions, err := s.queryHandler.HandleGetSubscriptionsByUserID(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	return s.dtoMapper.DomainListToResponse(subscriptions), nil
}

func (s *SubscriptionApplicationService) ListSubscriptions(
	ctx context.Context,
	req *dto.SubscriptionListRequest,
) (*dto.SubscriptionListResponse, error) {
	listQuery := &query.ListSubscriptionsQuery{
		Search:    req.Search,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Limit:     req.Limit,
		Offset:    req.Offset,
		AutoRenew: req.AutoRenew,
		InTrial:   req.InTrial,
		Expired:   req.Expired,
		Overdue:   req.Overdue,
	}

	if req.UserID != nil {
		userID, err := valueobject.NewUserID(*req.UserID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID: %w", err)
		}
		listQuery.UserID = userID
	}

	if req.PlanID != nil {
		planID, err := valueobject.NewPlanID(*req.PlanID)
		if err != nil {
			return nil, fmt.Errorf("invalid plan ID: %w", err)
		}
		listQuery.PlanID = planID
	}

	if req.Status != "" {
		status, err := valueobject.NewSubscriptionStatus(req.Status)
		if err != nil {
			return nil, fmt.Errorf("invalid status: %w", err)
		}
		listQuery.Status = status
	}

	if req.Currency != "" {
		currency, err := valueobject.NewCurrency(req.Currency)
		if err != nil {
			return nil, fmt.Errorf("invalid currency: %w", err)
		}
		listQuery.Currency = currency
	}

	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start date format: %w", err)
		}
		listQuery.StartDate = &startDate
	}

	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format: %w", err)
		}
		listQuery.EndDate = &endDate
	}

	subscriptions, total, err := s.queryHandler.HandleListSubscriptions(ctx, listQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	return &dto.SubscriptionListResponse{
		Subscriptions: s.dtoMapper.DomainListToResponse(subscriptions),
		Total:         total,
		Limit:         req.Limit,
		Offset:        req.Offset,
	}, nil
}

func (s *SubscriptionApplicationService) RenewSubscription(
	ctx context.Context,
	subscriptionID uint,
	req *dto.RenewSubscriptionRequest,
) error {
	id, err := valueobject.NewSubscriptionID(subscriptionID)
	if err != nil {
		return fmt.Errorf("invalid subscription ID: %w", err)
	}

	cmd := &command.RenewSubscriptionCommand{
		SubscriptionID:  *id,
		NewPeriodStart:  req.NewPeriodStart,
		NewPeriodEnd:    req.NewPeriodEnd,
		NextBillingDate: req.NextBillingDate,
	}

	return s.commandHandler.HandleRenewSubscription(ctx, cmd)
}

func (s *SubscriptionApplicationService) CancelSubscription(
	ctx context.Context,
	subscriptionID uint,
	req *dto.CancelSubscriptionRequest,
) error {
	id, err := valueobject.NewSubscriptionID(subscriptionID)
	if err != nil {
		return fmt.Errorf("invalid subscription ID: %w", err)
	}

	cmd := &command.CancelSubscriptionCommand{
		SubscriptionID: *id,
		Reason:         req.Reason,
		Immediately:    req.Immediately,
	}

	return s.commandHandler.HandleCancelSubscription(ctx, cmd)
}

func (s *SubscriptionApplicationService) PauseSubscription(
	ctx context.Context,
	subscriptionID uint,
	req *dto.PauseSubscriptionRequest,
) error {
	id, err := valueobject.NewSubscriptionID(subscriptionID)
	if err != nil {
		return fmt.Errorf("invalid subscription ID: %w", err)
	}

	cmd := &command.PauseSubscriptionCommand{
		SubscriptionID: *id,
		Reason:         req.Reason,
	}

	return s.commandHandler.HandlePauseSubscription(ctx, cmd)
}

func (s *SubscriptionApplicationService) ResumeSubscription(
	ctx context.Context,
	subscriptionID uint,
) error {
	id, err := valueobject.NewSubscriptionID(subscriptionID)
	if err != nil {
		return fmt.Errorf("invalid subscription ID: %w", err)
	}

	cmd := &command.ResumeSubscriptionCommand{
		SubscriptionID: *id,
	}

	return s.commandHandler.HandleResumeSubscription(ctx, cmd)
}

func (s *SubscriptionApplicationService) UpdateSubscriptionUsage(
	ctx context.Context,
	subscriptionID uint,
) error {
	id, err := valueobject.NewSubscriptionID(subscriptionID)
	if err != nil {
		return fmt.Errorf("invalid subscription ID: %w", err)
	}

	cmd := &command.UpdateSubscriptionUsageCommand{
		SubscriptionID: *id,
	}

	return s.commandHandler.HandleUpdateSubscriptionUsage(ctx, cmd)
}

func (s *SubscriptionApplicationService) UpdateSubscriptionServerGroups(
	ctx context.Context,
	subscriptionID uint,
	serverGroupIDs []uint,
) error {
	id, err := valueobject.NewSubscriptionID(subscriptionID)
	if err != nil {
		return fmt.Errorf("invalid subscription ID: %w", err)
	}

	cmd := &command.UpdateSubscriptionServerGroupsCommand{
		SubscriptionID: *id,
		ServerGroupIDs: serverGroupIDs,
	}

	return s.commandHandler.HandleUpdateSubscriptionServerGroups(ctx, cmd)
}

func (s *SubscriptionApplicationService) DeleteSubscription(
	ctx context.Context,
	subscriptionID uint,
) error {
	id, err := valueobject.NewSubscriptionID(subscriptionID)
	if err != nil {
		return fmt.Errorf("invalid subscription ID: %w", err)
	}

	cmd := &command.DeleteSubscriptionCommand{
		SubscriptionID: *id,
	}

	return s.commandHandler.HandleDeleteSubscription(ctx, cmd)
}