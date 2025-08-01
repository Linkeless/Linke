package model

import (
	"encoding/json"
	"errors"
	"time"

	sharedvo "linke/internal/shared/valueobject"
	"linke/internal/subscription/domain/event"
	"linke/internal/subscription/domain/valueobject"
)

type Subscription struct {
	id                 valueobject.SubscriptionID
	uuid               valueobject.SubscriptionUUID
	userID             *sharedvo.UserID
	planID             valueobject.PlanID
	orderID            uint
	status             valueobject.SubscriptionStatus
	startDate          time.Time
	endDate            time.Time
	currentPeriodStart time.Time
	currentPeriodEnd   time.Time
	billingCycle       valueobject.BillingCycle
	billingInterval    int
	price              sharedvo.Money
	autoRenew          bool
	nextBillingDate    *time.Time
	trialEndDate       *time.Time
	cancelAtPeriodEnd  bool
	cancellationReason string
	cancelledAt        *time.Time
	renewalAttempts    int
	lastRenewalFailed  *time.Time
	renewalFailReason  string
	lastUsedAt         *time.Time
	serverGroupIDs     []uint
	notes              string
	metadata           map[string]interface{}
	createdAt          time.Time
	updatedAt          time.Time
	deletedAt          *time.Time
	domainEvents       []event.DomainEvent
}

func NewSubscription(
	userID *sharedvo.UserID,
	planID valueobject.PlanID,
	orderID uint,
	startDate, endDate time.Time,
	billingCycle valueobject.BillingCycle,
	billingInterval int,
	price sharedvo.Money,
) (*Subscription, error) {
	if startDate.After(endDate) {
		return nil, errors.New("start date cannot be after end date")
	}

	if billingInterval <= 0 {
		return nil, errors.New("billing interval must be positive")
	}

	status, _ := valueobject.NewSubscriptionStatus("active")
	now := time.Now()

	subscription := &Subscription{
		uuid:               *valueobject.NewSubscriptionUUID(),
		userID:             userID,
		planID:             planID,
		orderID:            orderID,
		status:             *status,
		startDate:          startDate,
		endDate:            endDate,
		currentPeriodStart: startDate,
		currentPeriodEnd:   startDate.AddDate(0, billingInterval, 0),
		billingCycle:       billingCycle,
		billingInterval:    billingInterval,
		price:              price,
		autoRenew:          true,
		cancelAtPeriodEnd:  false,
		renewalAttempts:    0,
		serverGroupIDs:     []uint{},
		metadata:           make(map[string]interface{}),
		createdAt:          now,
		updatedAt:          now,
		domainEvents:       []event.DomainEvent{},
	}

	if !billingCycle.IsLifetime() {
		nextBilling := subscription.currentPeriodEnd
		subscription.nextBillingDate = &nextBilling
	}

	// Convert shared UserID to domain UserID for domain event
	domainUserID, err := valueobject.ConvertFromSharedUserID(*subscription.userID)
	if err != nil {
		return nil, err
	}
	
	subscription.AddDomainEvent(event.NewSubscriptionCreated(
		subscription.id,
		*domainUserID,
		subscription.planID,
	))

	return subscription, nil
}

func (s *Subscription) ID() valueobject.SubscriptionID {
	return s.id
}

func (s *Subscription) SetID(id valueobject.SubscriptionID) {
	s.id = id
}

func (s *Subscription) UUID() valueobject.SubscriptionUUID {
	return s.uuid
}

func (s *Subscription) UserID() valueobject.UserID {
	// Convert shared UserID back to domain UserID for external usage
	domainUserID, _ := valueobject.ConvertFromSharedUserID(*s.userID)
	return *domainUserID
}

func (s *Subscription) PlanID() valueobject.PlanID {
	return s.planID
}

func (s *Subscription) OrderID() uint {
	return s.orderID
}

func (s *Subscription) Status() valueobject.SubscriptionStatus {
	return s.status
}

func (s *Subscription) StartDate() time.Time {
	return s.startDate
}

func (s *Subscription) EndDate() time.Time {
	return s.endDate
}

func (s *Subscription) CurrentPeriodStart() time.Time {
	return s.currentPeriodStart
}

func (s *Subscription) CurrentPeriodEnd() time.Time {
	return s.currentPeriodEnd
}

func (s *Subscription) BillingCycle() valueobject.BillingCycle {
	return s.billingCycle
}

func (s *Subscription) BillingInterval() int {
	return s.billingInterval
}

func (s *Subscription) Price() valueobject.Price {
	// Convert shared Money back to domain Price for external usage
	domainPrice, _ := valueobject.ConvertFromSharedMoney(s.price)
	return *domainPrice
}

func (s *Subscription) AutoRenew() bool {
	return s.autoRenew
}

func (s *Subscription) NextBillingDate() *time.Time {
	return s.nextBillingDate
}

func (s *Subscription) TrialEndDate() *time.Time {
	return s.trialEndDate
}

func (s *Subscription) CancelAtPeriodEnd() bool {
	return s.cancelAtPeriodEnd
}

func (s *Subscription) CancellationReason() string {
	return s.cancellationReason
}

func (s *Subscription) CancelledAt() *time.Time {
	return s.cancelledAt
}

func (s *Subscription) RenewalAttempts() int {
	return s.renewalAttempts
}

func (s *Subscription) LastRenewalFailed() *time.Time {
	return s.lastRenewalFailed
}

func (s *Subscription) RenewalFailReason() string {
	return s.renewalFailReason
}

func (s *Subscription) LastUsedAt() *time.Time {
	return s.lastUsedAt
}

func (s *Subscription) ServerGroupIDs() []uint {
	return s.serverGroupIDs
}

func (s *Subscription) Notes() string {
	return s.notes
}

func (s *Subscription) Metadata() map[string]interface{} {
	return s.metadata
}

func (s *Subscription) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Subscription) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s *Subscription) DeletedAt() *time.Time {
	return s.deletedAt
}

func (s *Subscription) IsActive() bool {
	return s.status.IsActive() && !s.IsExpired() && !s.IsDeleted()
}

func (s *Subscription) IsPaused() bool {
	return s.status.IsPaused()
}

func (s *Subscription) IsCancelled() bool {
	return s.status.IsCancelled()
}

func (s *Subscription) IsExpired() bool {
	return s.status.IsExpired() || time.Now().After(s.endDate)
}

func (s *Subscription) IsDeleted() bool {
	return s.deletedAt != nil
}

func (s *Subscription) IsInTrial() bool {
	return s.trialEndDate != nil && time.Now().Before(*s.trialEndDate)
}

func (s *Subscription) DaysUntilExpiry() int {
	duration := time.Until(s.endDate)
	days := int(duration.Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

func (s *Subscription) HasAccessToServerGroup(groupID uint) bool {
	if len(s.serverGroupIDs) == 0 {
		return false
	}

	for _, id := range s.serverGroupIDs {
		if id == 0 {
			return true
		}
		if id == groupID {
			return true
		}
	}
	return false
}

func (s *Subscription) SetServerGroupIDs(groupIDs []uint) {
	s.serverGroupIDs = groupIDs
	s.updatedAt = time.Now()
}

func (s *Subscription) UpdateUsage() {
	now := time.Now()
	s.lastUsedAt = &now
	s.updatedAt = now
}

func (s *Subscription) Renew(newPeriodStart, newPeriodEnd time.Time, nextBillingDate *time.Time) error {
	if !s.IsActive() {
		return errors.New("cannot renew inactive subscription")
	}

	s.currentPeriodStart = newPeriodStart
	s.currentPeriodEnd = newPeriodEnd
	s.nextBillingDate = nextBillingDate
	s.renewalAttempts = 0
	s.lastRenewalFailed = nil
	s.renewalFailReason = ""
	s.updatedAt = time.Now()

	// Convert shared UserID to domain UserID for domain event
	domainUserID, _ := valueobject.ConvertFromSharedUserID(*s.userID)
	s.AddDomainEvent(event.NewSubscriptionRenewed(
		s.id,
		*domainUserID,
		newPeriodEnd,
	))

	return nil
}

func (s *Subscription) FailRenewal(reason string) {
	s.renewalAttempts++
	now := time.Now()
	s.lastRenewalFailed = &now
	s.renewalFailReason = reason
	s.updatedAt = now

	const maxAttempts = 3
	if s.renewalAttempts >= maxAttempts {
		cancelledStatus, _ := valueobject.NewSubscriptionStatus("cancelled")
		s.status = *cancelledStatus
		s.cancelledAt = &now
		s.cancellationReason = "Renewal failed after maximum attempts"

		// Convert shared UserID to domain UserID for domain event
		domainUserID, _ := valueobject.ConvertFromSharedUserID(*s.userID)
		s.AddDomainEvent(event.NewSubscriptionCancelled(
			s.id,
			*domainUserID,
			s.cancellationReason,
			true,
		))
	}
}

func (s *Subscription) Cancel(reason string, immediately bool) error {
	if s.IsCancelled() {
		return errors.New("subscription is already cancelled")
	}

	s.autoRenew = false
	s.cancellationReason = reason
	s.updatedAt = time.Now()

	if immediately {
		cancelledStatus, _ := valueobject.NewSubscriptionStatus("cancelled")
		s.status = *cancelledStatus
		now := time.Now()
		s.cancelledAt = &now
	} else {
		s.cancelAtPeriodEnd = true
	}

	// Convert shared UserID to domain UserID for domain event
	domainUserID, _ := valueobject.ConvertFromSharedUserID(*s.userID)
	s.AddDomainEvent(event.NewSubscriptionCancelled(
		s.id,
		*domainUserID,
		reason,
		immediately,
	))

	return nil
}

func (s *Subscription) Pause(reason string) error {
	if !s.IsActive() {
		return errors.New("can only pause active subscriptions")
	}

	pausedStatus, _ := valueobject.NewSubscriptionStatus("paused")
	s.status = *pausedStatus
	s.notes = reason
	s.updatedAt = time.Now()

	return nil
}

func (s *Subscription) Resume() error {
	if !s.IsPaused() {
		return errors.New("can only resume paused subscriptions")
	}

	activeStatus, _ := valueobject.NewSubscriptionStatus("active")
	s.status = *activeStatus
	s.updatedAt = time.Now()

	return nil
}

func (s *Subscription) Expire() {
	expiredStatus, _ := valueobject.NewSubscriptionStatus("expired")
	s.status = *expiredStatus
	s.updatedAt = time.Now()

	// Convert shared UserID to domain UserID for domain event
	domainUserID, _ := valueobject.ConvertFromSharedUserID(*s.userID)
	s.AddDomainEvent(event.NewSubscriptionExpired(
		s.id,
		*domainUserID,
		s.endDate,
	))
}

func (s *Subscription) SetTrial(trialEndDate time.Time) {
	s.trialEndDate = &trialEndDate
	s.currentPeriodEnd = trialEndDate
	s.updatedAt = time.Now()
}

func (s *Subscription) SetMetadata(key string, value interface{}) {
	if s.metadata == nil {
		s.metadata = make(map[string]interface{})
	}
	s.metadata[key] = value
	s.updatedAt = time.Now()
}

func (s *Subscription) GetMetadata(key string) (interface{}, bool) {
	value, exists := s.metadata[key]
	return value, exists
}

func (s *Subscription) SetNotes(notes string) {
	s.notes = notes
	s.updatedAt = time.Now()
}

func (s *Subscription) Delete() {
	now := time.Now()
	s.deletedAt = &now
	s.updatedAt = now
}

func (s *Subscription) GetServerGroupIDsJSON() (string, error) {
	if len(s.serverGroupIDs) == 0 {
		return "", nil
	}

	jsonBytes, err := json.Marshal(s.serverGroupIDs)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (s *Subscription) SetServerGroupIDsFromJSON(jsonStr string) error {
	if jsonStr == "" {
		s.serverGroupIDs = []uint{}
		return nil
	}

	var groupIDs []uint
	if err := json.Unmarshal([]byte(jsonStr), &groupIDs); err != nil {
		return err
	}

	s.serverGroupIDs = groupIDs
	s.updatedAt = time.Now()
	return nil
}

func (s *Subscription) DomainEvents() []event.DomainEvent {
	return s.domainEvents
}

func (s *Subscription) AddDomainEvent(domainEvent event.DomainEvent) {
	s.domainEvents = append(s.domainEvents, domainEvent)
}

func (s *Subscription) ClearDomainEvents() {
	s.domainEvents = []event.DomainEvent{}
}