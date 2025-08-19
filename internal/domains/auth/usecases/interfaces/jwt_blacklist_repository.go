package interfaces

import (
	"context"
	"time"

	"linke/internal/domains/auth/entities"
	"linke/internal/shared/framework"
)

// JWTBlacklistRepository defines the interface for JWT blacklist data access operations
type JWTBlacklistRepository interface {
	framework.TimeBasedRepository[entities.JWTBlacklist, string]

	// Token blacklisting operations
	GetByTokenHash(ctx context.Context, tokenHash string) (*entities.JWTBlacklist, error)
	IsTokenBlacklisted(ctx context.Context, tokenHash string) (bool, error)
	BlacklistToken(ctx context.Context, tokenHash string, userID *uint, reason string, expiresAt time.Time) error
	Delete(ctx context.Context, tokenHash string) error

	// User-wide operations
	BlacklistAllUserTokens(ctx context.Context, userID uint, reason string, beforeTime, expiresAt time.Time) error
	IsUserTokensBlacklisted(ctx context.Context, userID uint, tokenIssuedAt time.Time) (bool, error)
	GetUserBlacklistedTokens(ctx context.Context, userID uint) ([]*entities.JWTBlacklist, error)

	// List operations with pagination
	ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.JWTBlacklist, int64, error)
	ListByReason(ctx context.Context, reason string, limit, offset int) ([]*entities.JWTBlacklist, int64, error)

	// Maintenance operations
	DeleteExpired(ctx context.Context) (int64, error)
	DeleteByUser(ctx context.Context, userID uint) (int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)

	// Statistics
	CountByReason(ctx context.Context, reason string) (int64, error)
	CountByUser(ctx context.Context, userID uint) (int64, error)
	CountExpired(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)

	// Time-based queries
	GetExpiredBefore(ctx context.Context, before time.Time, limit int) ([]*entities.JWTBlacklist, error)
}
