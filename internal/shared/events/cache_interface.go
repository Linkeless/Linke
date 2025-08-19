package events

import (
	"context"
	"time"
)

// EventCacheStore defines the cache operations needed by event handlers
// This interface breaks the circular dependency between events and cache packages
type EventCacheStore interface {
	// Basic cache operations for event handlers
	Set(ctx context.Context, key, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// JSON operations for structured data
	SetJSON(ctx context.Context, key string, value any, expiration time.Duration) error
	GetJSON(ctx context.Context, key string, dest any) error

	// Pattern-based operations
	DeletePattern(ctx context.Context, pattern string) error
}
