package event

import (
	"time"
	
	"linke/internal/coupon/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// CouponLifecycleEvent represents events in coupon's lifecycle
type CouponLifecycleEvent struct {
	BaseEvent
	CouponID     valueobject.CouponID     `json:"coupon_id"`
	CouponCode   valueobject.CouponCode   `json:"coupon_code"`
	TriggeredBy  sharedvo.UserID          `json:"triggered_by"`
	PreviousState *valueobject.CouponStatus `json:"previous_state,omitempty"`
	NewState     valueobject.CouponStatus  `json:"new_state"`
	Reason       string                   `json:"reason,omitempty"`
}

// BaseEvent provides common event functionality
type BaseEvent struct {
	EventID      string    `json:"event_id"`
	EventType    string    `json:"event_type"`
	AggregateID  string    `json:"aggregate_id"`
	OccurredAt   time.Time `json:"occurred_at"`
	EventVersion int       `json:"event_version"`
}

// CouponCreated event
type CouponCreated struct {
	CouponLifecycleEvent
	DiscountRule      valueobject.DiscountRule     `json:"discount_rule"`
	ValidityPeriod   valueobject.ValidityPeriod   `json:"validity_period"`
	UsagePolicy      valueobject.UsageLimits      `json:"usage_policy"`
	IsPublic         bool                         `json:"is_public"`
	ApplicablePlans  []uint64                     `json:"applicable_plans"`
}

// NewCouponCreated creates a new coupon created event
func NewCouponCreated(
	eventID string,
	couponID valueobject.CouponID,
	couponCode valueobject.CouponCode,
	discountRule valueobject.DiscountRule,
	validityPeriod valueobject.ValidityPeriod,
	usagePolicy valueobject.UsageLimits,
	createdBy sharedvo.UserID,
	isPublic bool,
	applicablePlans []uint64,
) *CouponCreated {
	return &CouponCreated{
		CouponLifecycleEvent: CouponLifecycleEvent{
			BaseEvent: BaseEvent{
				EventID:      eventID,
				EventType:    "CouponCreated",
				AggregateID:  couponID.String(),
				OccurredAt:   time.Now(),
				EventVersion: 1,
			},
			CouponID:     couponID,
			CouponCode:   couponCode,
			TriggeredBy:  createdBy,
			NewState:     valueobject.CouponStatusActive,
		},
		DiscountRule:     discountRule,
		ValidityPeriod:   validityPeriod,
		UsagePolicy:      usagePolicy,
		IsPublic:         isPublic,
		ApplicablePlans:  applicablePlans,
	}
}

// CouponStatusChanged event
type CouponStatusChanged struct {
	CouponLifecycleEvent
}

// NewCouponStatusChanged creates a new coupon status changed event
func NewCouponStatusChanged(
	eventID string,
	couponID valueobject.CouponID,
	couponCode valueobject.CouponCode,
	previousState valueobject.CouponStatus,
	newState valueobject.CouponStatus,
	changedBy sharedvo.UserID,
	reason string,
) *CouponStatusChanged {
	return &CouponStatusChanged{
		CouponLifecycleEvent: CouponLifecycleEvent{
			BaseEvent: BaseEvent{
				EventID:      eventID,
				EventType:    "CouponStatusChanged",
				AggregateID:  couponID.String(),
				OccurredAt:   time.Now(),
				EventVersion: 1,
			},
			CouponID:      couponID,
			CouponCode:    couponCode,
			TriggeredBy:   changedBy,
			PreviousState: &previousState,
			NewState:      newState,
			Reason:        reason,
		},
	}
}

// CouponExpired event
type CouponExpired struct {
	CouponLifecycleEvent
	ExpiredAt time.Time `json:"expired_at"`
}

// NewCouponExpired creates a new coupon expired event
func NewCouponExpired(
	eventID string,
	couponID valueobject.CouponID,
	couponCode valueobject.CouponCode,
	expiredAt time.Time,
) *CouponExpired {
	return &CouponExpired{
		CouponLifecycleEvent: CouponLifecycleEvent{
			BaseEvent: BaseEvent{
				EventID:      eventID,
				EventType:    "CouponExpired",
				AggregateID:  couponID.String(),
				OccurredAt:   time.Now(),
				EventVersion: 1,
			},
			CouponID:      couponID,
			CouponCode:    couponCode,
			PreviousState: &[]valueobject.CouponStatus{valueobject.CouponStatusActive}[0],
			NewState:      valueobject.CouponStatusExpired,
			Reason:        "Validity period ended",
		},
		ExpiredAt: expiredAt,
	}
}

// CouponUsageLimitReached event
type CouponUsageLimitReached struct {
	CouponLifecycleEvent
	TotalUsageCount int `json:"total_usage_count"`
	MaxUsageLimit   int `json:"max_usage_limit"`
}

// NewCouponUsageLimitReached creates a new usage limit reached event
func NewCouponUsageLimitReached(
	eventID string,
	couponID valueobject.CouponID,
	couponCode valueobject.CouponCode,
	totalUsageCount int,
	maxUsageLimit int,
) *CouponUsageLimitReached {
	return &CouponUsageLimitReached{
		CouponLifecycleEvent: CouponLifecycleEvent{
			BaseEvent: BaseEvent{
				EventID:      eventID,
				EventType:    "CouponUsageLimitReached",
				AggregateID:  couponID.String(),
				OccurredAt:   time.Now(),
				EventVersion: 1,
			},
			CouponID:      couponID,
			CouponCode:    couponCode,
			PreviousState: &[]valueobject.CouponStatus{valueobject.CouponStatusActive}[0],
			NewState:      valueobject.CouponStatusInactive,
			Reason:        "Usage limit reached",
		},
		TotalUsageCount: totalUsageCount,
		MaxUsageLimit:   maxUsageLimit,
	}
}

// Event interface implementations
func (e CouponCreated) GetEventID() string      { return e.EventID }
func (e CouponCreated) GetEventType() string    { return e.EventType }
func (e CouponCreated) GetAggregateID() string  { return e.AggregateID }
func (e CouponCreated) GetOccurredAt() time.Time { return e.OccurredAt }
func (e CouponCreated) GetEventVersion() int    { return e.EventVersion }

func (e CouponStatusChanged) GetEventID() string      { return e.EventID }
func (e CouponStatusChanged) GetEventType() string    { return e.EventType }
func (e CouponStatusChanged) GetAggregateID() string  { return e.AggregateID }
func (e CouponStatusChanged) GetOccurredAt() time.Time { return e.OccurredAt }
func (e CouponStatusChanged) GetEventVersion() int    { return e.EventVersion }

func (e CouponExpired) GetEventID() string      { return e.EventID }
func (e CouponExpired) GetEventType() string    { return e.EventType }
func (e CouponExpired) GetAggregateID() string  { return e.AggregateID }
func (e CouponExpired) GetOccurredAt() time.Time { return e.OccurredAt }
func (e CouponExpired) GetEventVersion() int    { return e.EventVersion }

func (e CouponUsageLimitReached) GetEventID() string      { return e.EventID }
func (e CouponUsageLimitReached) GetEventType() string    { return e.EventType }
func (e CouponUsageLimitReached) GetAggregateID() string  { return e.AggregateID }
func (e CouponUsageLimitReached) GetOccurredAt() time.Time { return e.OccurredAt }
func (e CouponUsageLimitReached) GetEventVersion() int    { return e.EventVersion }