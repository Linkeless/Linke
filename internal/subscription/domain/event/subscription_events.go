package event

import (
	"time"

	"linke/internal/subscription/domain/valueobject"
)

type DomainEvent interface {
	OccurredOn() time.Time
	EventType() string
}

type SubscriptionCreated struct {
	subscriptionID valueobject.SubscriptionID
	userID         valueobject.UserID
	planID         valueobject.PlanID
	occurredOn     time.Time
}

func NewSubscriptionCreated(
	subscriptionID valueobject.SubscriptionID,
	userID valueobject.UserID,
	planID valueobject.PlanID,
) *SubscriptionCreated {
	return &SubscriptionCreated{
		subscriptionID: subscriptionID,
		userID:         userID,
		planID:         planID,
		occurredOn:     time.Now(),
	}
}

func (e SubscriptionCreated) OccurredOn() time.Time {
	return e.occurredOn
}

func (e SubscriptionCreated) EventType() string {
	return "subscription.created"
}

func (e SubscriptionCreated) SubscriptionID() valueobject.SubscriptionID {
	return e.subscriptionID
}

func (e SubscriptionCreated) UserID() valueobject.UserID {
	return e.userID
}

func (e SubscriptionCreated) PlanID() valueobject.PlanID {
	return e.planID
}

type SubscriptionRenewed struct {
	subscriptionID valueobject.SubscriptionID
	userID         valueobject.UserID
	newPeriodEnd   time.Time
	occurredOn     time.Time
}

func NewSubscriptionRenewed(
	subscriptionID valueobject.SubscriptionID,
	userID valueobject.UserID,
	newPeriodEnd time.Time,
) *SubscriptionRenewed {
	return &SubscriptionRenewed{
		subscriptionID: subscriptionID,
		userID:         userID,
		newPeriodEnd:   newPeriodEnd,
		occurredOn:     time.Now(),
	}
}

func (e SubscriptionRenewed) OccurredOn() time.Time {
	return e.occurredOn
}

func (e SubscriptionRenewed) EventType() string {
	return "subscription.renewed"
}

func (e SubscriptionRenewed) SubscriptionID() valueobject.SubscriptionID {
	return e.subscriptionID
}

func (e SubscriptionRenewed) UserID() valueobject.UserID {
	return e.userID
}

func (e SubscriptionRenewed) NewPeriodEnd() time.Time {
	return e.newPeriodEnd
}

type SubscriptionCancelled struct {
	subscriptionID valueobject.SubscriptionID
	userID         valueobject.UserID
	reason         string
	immediately    bool
	occurredOn     time.Time
}

func NewSubscriptionCancelled(
	subscriptionID valueobject.SubscriptionID,
	userID valueobject.UserID,
	reason string,
	immediately bool,
) *SubscriptionCancelled {
	return &SubscriptionCancelled{
		subscriptionID: subscriptionID,
		userID:         userID,
		reason:         reason,
		immediately:    immediately,
		occurredOn:     time.Now(),
	}
}

func (e SubscriptionCancelled) OccurredOn() time.Time {
	return e.occurredOn
}

func (e SubscriptionCancelled) EventType() string {
	return "subscription.cancelled"
}

func (e SubscriptionCancelled) SubscriptionID() valueobject.SubscriptionID {
	return e.subscriptionID
}

func (e SubscriptionCancelled) UserID() valueobject.UserID {
	return e.userID
}

func (e SubscriptionCancelled) Reason() string {
	return e.reason
}

func (e SubscriptionCancelled) Immediately() bool {
	return e.immediately
}

type SubscriptionExpired struct {
	subscriptionID valueobject.SubscriptionID
	userID         valueobject.UserID
	expiredAt      time.Time
	occurredOn     time.Time
}

func NewSubscriptionExpired(
	subscriptionID valueobject.SubscriptionID,
	userID valueobject.UserID,
	expiredAt time.Time,
) *SubscriptionExpired {
	return &SubscriptionExpired{
		subscriptionID: subscriptionID,
		userID:         userID,
		expiredAt:      expiredAt,
		occurredOn:     time.Now(),
	}
}

func (e SubscriptionExpired) OccurredOn() time.Time {
	return e.occurredOn
}

func (e SubscriptionExpired) EventType() string {
	return "subscription.expired"
}

func (e SubscriptionExpired) SubscriptionID() valueobject.SubscriptionID {
	return e.subscriptionID
}

func (e SubscriptionExpired) UserID() valueobject.UserID {
	return e.userID
}

func (e SubscriptionExpired) ExpiredAt() time.Time {
	return e.expiredAt
}