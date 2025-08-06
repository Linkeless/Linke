package cache

import (
	"context"
	"time"
)

// CacheEvent defines the minimal interface for events that can trigger cache invalidation
// This interface breaks the circular dependency between cache and events packages
type CacheEvent interface {
	EventType() string
	EventID() string  
	EventTime() time.Time
	EventData() any
}

// CacheEventHandler defines the minimal interface for handling cache invalidation events
type CacheEventHandler interface {
	Handle(ctx context.Context, event CacheEvent) error
	EventTypes() []string
}

// UserEvent represents user-related events for cache invalidation
type UserCacheEvent struct {
	ID     string
	Type   string
	UserID uint
	Time   time.Time
	Data   any
}

func (e *UserCacheEvent) EventType() string    { return e.Type }
func (e *UserCacheEvent) EventID() string     { return e.ID }
func (e *UserCacheEvent) EventTime() time.Time { return e.Time }
func (e *UserCacheEvent) EventData() any      { return e.Data }

// SubscriptionEvent represents subscription-related events for cache invalidation
type SubscriptionCacheEvent struct {
	ID             string
	Type           string
	SubscriptionID uint
	UserID         uint
	Time           time.Time
	Data           any
}

func (e *SubscriptionCacheEvent) EventType() string    { return e.Type }
func (e *SubscriptionCacheEvent) EventID() string     { return e.ID }
func (e *SubscriptionCacheEvent) EventTime() time.Time { return e.Time }
func (e *SubscriptionCacheEvent) EventData() any      { return e.Data }

// PaymentEvent represents payment-related events for cache invalidation  
type PaymentCacheEvent struct {
	ID        string
	Type      string
	PaymentID string
	UserID    uint
	Time      time.Time
	Data      any
}

func (e *PaymentCacheEvent) EventType() string    { return e.Type }
func (e *PaymentCacheEvent) EventID() string     { return e.ID }
func (e *PaymentCacheEvent) EventTime() time.Time { return e.Time }
func (e *PaymentCacheEvent) EventData() any      { return e.Data }

// OrderEvent represents order-related events for cache invalidation
type OrderCacheEvent struct {
	ID      string
	Type    string
	OrderID uint
	UserID  uint
	Time    time.Time
	Data    any
}

func (e *OrderCacheEvent) EventType() string    { return e.Type }
func (e *OrderCacheEvent) EventID() string     { return e.ID }
func (e *OrderCacheEvent) EventTime() time.Time { return e.Time }
func (e *OrderCacheEvent) EventData() any      { return e.Data }

// InvoiceEvent represents invoice-related events for cache invalidation
type InvoiceCacheEvent struct {
	ID        string
	Type      string
	InvoiceID uint
	OrderID   uint
	UserID    uint
	Time      time.Time
	Data      any
}

func (e *InvoiceCacheEvent) EventType() string    { return e.Type }
func (e *InvoiceCacheEvent) EventID() string     { return e.ID }
func (e *InvoiceCacheEvent) EventTime() time.Time { return e.Time }
func (e *InvoiceCacheEvent) EventData() any      { return e.Data }