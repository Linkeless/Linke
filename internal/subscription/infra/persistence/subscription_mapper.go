package persistence

import (
	"encoding/json"
	"fmt"

	"linke/internal/subscription/domain/model"
	"linke/internal/subscription/domain/valueobject"
)

type SubscriptionMapper struct{}

func NewSubscriptionMapper() *SubscriptionMapper {
	return &SubscriptionMapper{}
}

func (m *SubscriptionMapper) DomainToPO(subscription *model.Subscription) (*SubscriptionPO, error) {
	serverGroupIDsJSON, err := subscription.GetServerGroupIDsJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal server group IDs: %w", err)
	}

	metadataJSON, err := json.Marshal(subscription.Metadata())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	po := &SubscriptionPO{
		ID:                 subscription.ID().Value(),
		OrderID:            subscription.OrderID(),
		UserID:             subscription.UserID().Value(),
		PlanID:             subscription.PlanID().Value(),
		UUID:               subscription.UUID().Value(),
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
		ServerGroupIDs:     serverGroupIDsJSON,
		Notes:              subscription.Notes(),
		Metadata:           string(metadataJSON),
		CreatedAt:          subscription.CreatedAt(),
		UpdatedAt:          subscription.UpdatedAt(),
	}

	if subscription.DeletedAt() != nil {
		po.DeletedAt.Time = *subscription.DeletedAt()
		po.DeletedAt.Valid = true
	}

	return po, nil
}

func (m *SubscriptionMapper) POToDomain(po *SubscriptionPO) (*model.Subscription, error) {
	subscriptionID, err := valueobject.NewSubscriptionID(po.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription ID: %w", err)
	}

	_, err = valueobject.NewSubscriptionUUIDFromString(po.UUID)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription UUID: %w", err)
	}

	userID, err := valueobject.NewUserID(po.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	planID, err := valueobject.NewPlanID(po.PlanID)
	if err != nil {
		return nil, fmt.Errorf("invalid plan ID: %w", err)
	}

	_, err = valueobject.NewSubscriptionStatus(po.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription status: %w", err)
	}

	billingCycle, err := valueobject.NewBillingCycle(po.BillingCycle)
	if err != nil {
		return nil, fmt.Errorf("invalid billing cycle: %w", err)
	}

	currency, err := valueobject.NewCurrency(po.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	price, err := valueobject.NewPrice(po.Price, *currency)
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}

	var serverGroupIDs []uint
	if po.ServerGroupIDs != "" {
		if err := json.Unmarshal([]byte(po.ServerGroupIDs), &serverGroupIDs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal server group IDs: %w", err)
		}
	}

	var metadata map[string]interface{}
	if po.Metadata != "" {
		if err := json.Unmarshal([]byte(po.Metadata), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	} else {
		metadata = make(map[string]interface{})
	}

	// Convert domain UserID to shared UserID
	sharedUserID, err := valueobject.ConvertToSharedUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert user ID: %w", err)
	}
	
	// Convert domain Price to shared Money
	sharedMoney, err := valueobject.ConvertToSharedMoney(price)
	if err != nil {
		return nil, fmt.Errorf("failed to convert price: %w", err)
	}

	subscription, err := model.NewSubscription(
		&sharedUserID,
		*planID,
		po.OrderID,
		po.StartDate,
		po.EndDate,
		*billingCycle,
		po.BillingInterval,
		sharedMoney,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription domain object: %w", err)
	}

	subscription.SetID(*subscriptionID)

	if len(serverGroupIDs) > 0 {
		subscription.SetServerGroupIDs(serverGroupIDs)
	}

	if po.Notes != "" {
		subscription.SetNotes(po.Notes)
	}

	for key, value := range metadata {
		subscription.SetMetadata(key, value)
	}

	if po.TrialEndDate != nil {
		subscription.SetTrial(*po.TrialEndDate)
	}

	subscription.ClearDomainEvents()

	return subscription, nil
}