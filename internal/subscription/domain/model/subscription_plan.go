package model

import (
	"errors"
	"time"

	"linke/internal/subscription/domain/valueobject"
)

type SubscriptionPlan struct {
	id              valueobject.PlanID
	name            string
	description     string
	billingCycle    valueobject.BillingCycle
	billingInterval int
	price           valueobject.Price
	trialDays       int
	features        []string
	limitations     map[string]interface{}
	serverGroupIDs  []uint
	isActive        bool
	isVisible       bool
	sortOrder       int
	metadata        map[string]interface{}
	createdAt       time.Time
	updatedAt       time.Time
	deletedAt       *time.Time
}

func NewSubscriptionPlan(
	name, description string,
	billingCycle valueobject.BillingCycle,
	billingInterval int,
	price valueobject.Price,
) (*SubscriptionPlan, error) {
	if name == "" {
		return nil, errors.New("plan name cannot be empty")
	}

	if billingInterval <= 0 {
		return nil, errors.New("billing interval must be positive")
	}

	now := time.Now()

	return &SubscriptionPlan{
		name:            name,
		description:     description,
		billingCycle:    billingCycle,
		billingInterval: billingInterval,
		price:           price,
		trialDays:       0,
		features:        []string{},
		limitations:     make(map[string]interface{}),
		serverGroupIDs:  []uint{},
		isActive:        true,
		isVisible:       true,
		sortOrder:       0,
		metadata:        make(map[string]interface{}),
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

func (p *SubscriptionPlan) ID() valueobject.PlanID {
	return p.id
}

func (p *SubscriptionPlan) SetID(id valueobject.PlanID) {
	p.id = id
}

func (p *SubscriptionPlan) Name() string {
	return p.name
}

func (p *SubscriptionPlan) Description() string {
	return p.description
}

func (p *SubscriptionPlan) BillingCycle() valueobject.BillingCycle {
	return p.billingCycle
}

func (p *SubscriptionPlan) BillingInterval() int {
	return p.billingInterval
}

func (p *SubscriptionPlan) Price() valueobject.Price {
	return p.price
}

func (p *SubscriptionPlan) TrialDays() int {
	return p.trialDays
}

func (p *SubscriptionPlan) Features() []string {
	return p.features
}

func (p *SubscriptionPlan) Limitations() map[string]interface{} {
	return p.limitations
}

func (p *SubscriptionPlan) ServerGroupIDs() []uint {
	return p.serverGroupIDs
}

func (p *SubscriptionPlan) IsActive() bool {
	return p.isActive
}

func (p *SubscriptionPlan) IsVisible() bool {
	return p.isVisible
}

func (p *SubscriptionPlan) SortOrder() int {
	return p.sortOrder
}

func (p *SubscriptionPlan) Metadata() map[string]interface{} {
	return p.metadata
}

func (p *SubscriptionPlan) CreatedAt() time.Time {
	return p.createdAt
}

func (p *SubscriptionPlan) UpdatedAt() time.Time {
	return p.updatedAt
}

func (p *SubscriptionPlan) DeletedAt() *time.Time {
	return p.deletedAt
}

func (p *SubscriptionPlan) IsDeleted() bool {
	return p.deletedAt != nil
}

func (p *SubscriptionPlan) UpdateName(name string) error {
	if name == "" {
		return errors.New("plan name cannot be empty")
	}
	p.name = name
	p.updatedAt = time.Now()
	return nil
}

func (p *SubscriptionPlan) UpdateDescription(description string) {
	p.description = description
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) UpdatePrice(price valueobject.Price) {
	p.price = price
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) SetTrialDays(days int) error {
	if days < 0 {
		return errors.New("trial days cannot be negative")
	}
	p.trialDays = days
	p.updatedAt = time.Now()
	return nil
}

func (p *SubscriptionPlan) AddFeature(feature string) {
	if feature == "" {
		return
	}
	for _, f := range p.features {
		if f == feature {
			return
		}
	}
	p.features = append(p.features, feature)
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) RemoveFeature(feature string) {
	for i, f := range p.features {
		if f == feature {
			p.features = append(p.features[:i], p.features[i+1:]...)
			p.updatedAt = time.Now()
			break
		}
	}
}

func (p *SubscriptionPlan) SetFeatures(features []string) {
	p.features = features
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) HasFeature(feature string) bool {
	for _, f := range p.features {
		if f == feature {
			return true
		}
	}
	return false
}

func (p *SubscriptionPlan) SetLimitation(key string, value interface{}) {
	if p.limitations == nil {
		p.limitations = make(map[string]interface{})
	}
	p.limitations[key] = value
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) GetLimitation(key string) (interface{}, bool) {
	value, exists := p.limitations[key]
	return value, exists
}

func (p *SubscriptionPlan) RemoveLimitation(key string) {
	delete(p.limitations, key)
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) SetServerGroupIDs(groupIDs []uint) {
	p.serverGroupIDs = groupIDs
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) AddServerGroup(groupID uint) {
	for _, id := range p.serverGroupIDs {
		if id == groupID {
			return
		}
	}
	p.serverGroupIDs = append(p.serverGroupIDs, groupID)
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) RemoveServerGroup(groupID uint) {
	for i, id := range p.serverGroupIDs {
		if id == groupID {
			p.serverGroupIDs = append(p.serverGroupIDs[:i], p.serverGroupIDs[i+1:]...)
			p.updatedAt = time.Now()
			break
		}
	}
}

func (p *SubscriptionPlan) HasServerGroup(groupID uint) bool {
	for _, id := range p.serverGroupIDs {
		if id == groupID {
			return true
		}
	}
	return false
}

func (p *SubscriptionPlan) Activate() {
	p.isActive = true
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) Deactivate() {
	p.isActive = false
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) Show() {
	p.isVisible = true
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) Hide() {
	p.isVisible = false
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) SetSortOrder(order int) {
	p.sortOrder = order
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) SetMetadata(key string, value interface{}) {
	if p.metadata == nil {
		p.metadata = make(map[string]interface{})
	}
	p.metadata[key] = value
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) GetMetadata(key string) (interface{}, bool) {
	value, exists := p.metadata[key]
	return value, exists
}

func (p *SubscriptionPlan) RemoveMetadata(key string) {
	delete(p.metadata, key)
	p.updatedAt = time.Now()
}

func (p *SubscriptionPlan) Delete() {
	now := time.Now()
	p.deletedAt = &now
	p.updatedAt = now
}

func (p *SubscriptionPlan) CalculateServiceEndDate(startDate time.Time, servicePeriod int) time.Time {
	switch p.billingCycle {
	case valueobject.Monthly:
		return startDate.AddDate(0, servicePeriod, 0)
	case valueobject.Yearly:
		return startDate.AddDate(servicePeriod, 0, 0)
	case valueobject.Lifetime:
		return startDate.AddDate(99, 0, 0)
	default:
		return startDate.AddDate(0, servicePeriod, 0)
	}
}

func (p *SubscriptionPlan) IsAvailableForPurchase() bool {
	return p.isActive && p.isVisible && !p.IsDeleted()
}