package event

import (
	"time"

	"linke/internal/coupon/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// BaseDomainEvent provides common functionality for domain events
type BaseDomainEvent struct {
	eventID      string
	occurredAt   time.Time
	eventType    string
	aggregateID  string
	eventVersion int
}

// EventID returns the event ID
func (e BaseDomainEvent) EventID() string {
	return e.eventID
}

// OccurredAt returns when the event occurred
func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// EventType returns the event type
func (e BaseDomainEvent) EventType() string {
	return e.eventType
}

// AggregateID returns the aggregate ID
func (e BaseDomainEvent) AggregateID() string {
	return e.aggregateID
}

// EventVersion returns the event version
func (e BaseDomainEvent) EventVersion() int {
	return e.eventVersion
}

// EventData returns the event data
func (e BaseDomainEvent) EventData() interface{} {
	return e
}

// CouponCreatedEvent represents the event when a coupon is created
type CouponCreatedEvent struct {
	BaseDomainEvent
	CouponID        valueobject.CouponID
	CouponCode      valueobject.CouponCode
	CouponType      valueobject.CouponType
	DiscountValue   valueobject.DiscountValue
	MinOrderAmount  valueobject.Money
	ValidityPeriod  valueobject.ValidityPeriod
	UsageLimits     valueobject.UsageLimits
	CreatedBy       sharedvo.UserID
	Name            string
	Description     string
	ApplicablePlans []uint64
	IsPublic        bool
}

// NewCouponCreatedEvent creates a new coupon created event
func NewCouponCreatedEvent(
	eventID string,
	couponID valueobject.CouponID,
	couponCode valueobject.CouponCode,
	couponType valueobject.CouponType,
	discountValue valueobject.DiscountValue,
	minOrderAmount valueobject.Money,
	validityPeriod valueobject.ValidityPeriod,
	usageLimits valueobject.UsageLimits,
	createdBy sharedvo.UserID,
	name string,
	description string,
	applicablePlans []uint64,
	isPublic bool,
) *CouponCreatedEvent {
	return &CouponCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			eventID:      eventID,
			occurredAt:   time.Now(),
			eventType:    "CouponCreated",
			aggregateID:  couponID.String(),
			eventVersion: 1,
		},
		CouponID:        couponID,
		CouponCode:      couponCode,
		CouponType:      couponType,
		DiscountValue:   discountValue,
		MinOrderAmount:  minOrderAmount,
		ValidityPeriod:  validityPeriod,
		UsageLimits:     usageLimits,
		CreatedBy:       createdBy,
		Name:            name,
		Description:     description,
		ApplicablePlans: applicablePlans,
		IsPublic:        isPublic,
	}
}

// EventData returns the event data
func (e *CouponCreatedEvent) EventData() interface{} {
	return map[string]interface{}{
		"coupon_id":        e.CouponID.String(),
		"coupon_code":      e.CouponCode.String(),
		"coupon_type":      e.CouponType.String(),
		"discount_value":   e.DiscountValue.Value(),
		"min_order_amount": e.MinOrderAmount.Amount(),
		"currency":         e.MinOrderAmount.Currency().String(),
		"name":             e.Name,
		"description":      e.Description,
		"applicable_plans": e.ApplicablePlans,
		"is_public":        e.IsPublic,
		"created_by":       e.CreatedBy.String(),
	}
}

// CouponUsedEvent represents the event when a coupon is used
type CouponUsedEvent struct {
	BaseDomainEvent
	CouponID       valueobject.CouponID
	UserID         sharedvo.UserID
	OrderID        uint64
	DiscountAmount valueobject.Money
	OrderAmount    valueobject.Money
	RemainingUses  *int // nil if unlimited
}

// NewCouponUsedEvent creates a new coupon used event
func NewCouponUsedEvent(
	eventID string,
	couponID valueobject.CouponID,
	userID sharedvo.UserID,
	orderID uint64,
	discountAmount valueobject.Money,
	orderAmount valueobject.Money,
	remainingUses *int,
) *CouponUsedEvent {
	return &CouponUsedEvent{
		BaseDomainEvent: BaseDomainEvent{
			eventID:      eventID,
			occurredAt:   time.Now(),
			eventType:    "CouponUsed",
			aggregateID:  couponID.String(),
			eventVersion: 1,
		},
		CouponID:       couponID,
		UserID:         userID,
		OrderID:        orderID,
		DiscountAmount: discountAmount,
		OrderAmount:    orderAmount,
		RemainingUses:  remainingUses,
	}
}

// EventData returns the event data
func (e *CouponUsedEvent) EventData() interface{} {
	return map[string]interface{}{
		"coupon_id":       e.CouponID.String(),
		"user_id":         e.UserID.String(),
		"order_id":        e.OrderID,
		"discount_amount": e.DiscountAmount.Amount(),
		"order_amount":    e.OrderAmount.Amount(),
		"currency":        e.DiscountAmount.Currency().String(),
		"remaining_uses":  e.RemainingUses,
	}
}

// CouponExpiredEvent represents the event when a coupon expires
type CouponExpiredEvent struct {
	BaseDomainEvent
	CouponID   valueobject.CouponID
	CouponCode valueobject.CouponCode
	ExpiredAt  time.Time
}

// NewCouponExpiredEvent creates a new coupon expired event
func NewCouponExpiredEvent(
	eventID string,
	couponID valueobject.CouponID,
	couponCode valueobject.CouponCode,
	expiredAt time.Time,
) *CouponExpiredEvent {
	return &CouponExpiredEvent{
		BaseDomainEvent: BaseDomainEvent{
			eventID:      eventID,
			occurredAt:   time.Now(),
			eventType:    "CouponExpired",
			aggregateID:  couponID.String(),
			eventVersion: 1,
		},
		CouponID:   couponID,
		CouponCode: couponCode,
		ExpiredAt:  expiredAt,
	}
}

// EventData returns the event data
func (e *CouponExpiredEvent) EventData() interface{} {
	return map[string]interface{}{
		"coupon_id":   e.CouponID.String(),
		"coupon_code": e.CouponCode.String(),
		"expired_at":  e.ExpiredAt.Format(time.RFC3339),
	}
}

// CouponStatusChangedEvent represents the event when a coupon status changes
type CouponStatusChangedEvent struct {
	BaseDomainEvent
	CouponID  valueobject.CouponID
	OldStatus valueobject.CouponStatus
	NewStatus valueobject.CouponStatus
	ChangedBy sharedvo.UserID
	Reason    string
}

// NewCouponStatusChangedEvent creates a new coupon status changed event
func NewCouponStatusChangedEvent(
	eventID string,
	couponID valueobject.CouponID,
	oldStatus valueobject.CouponStatus,
	newStatus valueobject.CouponStatus,
	changedBy sharedvo.UserID,
	reason string,
) *CouponStatusChangedEvent {
	return &CouponStatusChangedEvent{
		BaseDomainEvent: BaseDomainEvent{
			eventID:      eventID,
			occurredAt:   time.Now(),
			eventType:    "CouponStatusChanged",
			aggregateID:  couponID.String(),
			eventVersion: 1,
		},
		CouponID:  couponID,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		ChangedBy: changedBy,
		Reason:    reason,
	}
}

// EventData returns the event data
func (e *CouponStatusChangedEvent) EventData() interface{} {
	return map[string]interface{}{
		"coupon_id":  e.CouponID.String(),
		"old_status": e.OldStatus.String(),
		"new_status": e.NewStatus.String(),
		"changed_by": e.ChangedBy.String(),
		"reason":     e.Reason,
	}
}

// CouponUsageLimitReachedEvent represents the event when a coupon reaches its usage limit
type CouponUsageLimitReachedEvent struct {
	BaseDomainEvent
	CouponID    valueobject.CouponID
	CouponCode  valueobject.CouponCode
	UsageLimits valueobject.UsageLimits
}

// NewCouponUsageLimitReachedEvent creates a new coupon usage limit reached event
func NewCouponUsageLimitReachedEvent(
	eventID string,
	couponID valueobject.CouponID,
	couponCode valueobject.CouponCode,
	usageLimits valueobject.UsageLimits,
) *CouponUsageLimitReachedEvent {
	return &CouponUsageLimitReachedEvent{
		BaseDomainEvent: BaseDomainEvent{
			eventID:      eventID,
			occurredAt:   time.Now(),
			eventType:    "CouponUsageLimitReached",
			aggregateID:  couponID.String(),
			eventVersion: 1,
		},
		CouponID:    couponID,
		CouponCode:  couponCode,
		UsageLimits: usageLimits,
	}
}

// EventData returns the event data
func (e *CouponUsageLimitReachedEvent) EventData() interface{} {
	return map[string]interface{}{
		"coupon_id":    e.CouponID.String(),
		"coupon_code":  e.CouponCode.String(),
		"max_uses":     e.UsageLimits.MaxUses(),
		"used_count":   e.UsageLimits.UsedCount(),
		"max_per_user": e.UsageLimits.MaxUsesPerUser(),
	}
}
