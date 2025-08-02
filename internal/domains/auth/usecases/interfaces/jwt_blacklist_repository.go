package interfaces

import (
	"context"
	"time"

	"linke/internal/domains/auth/entities"
)

// JWTBlacklistRepository defines the interface for JWT blacklist data access operations
type JWTBlacklistRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, blacklist *entities.JWTBlacklist) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entities.JWTBlacklist, error)
	Delete(ctx context.Context, tokenHash string) error

	// Token blacklisting operations
	IsTokenBlacklisted(ctx context.Context, tokenHash string) (bool, error)
	BlacklistToken(ctx context.Context, tokenHash string, userID *uint, reason string, expiresAt time.Time) error

	// User-wide operations
	BlacklistAllUserTokens(ctx context.Context, userID uint, reason string, beforeTime time.Time, expiresAt time.Time) error
	IsUserTokensBlacklisted(ctx context.Context, userID uint, tokenIssuedAt time.Time) (bool, error)
	GetUserBlacklistedTokens(ctx context.Context, userID uint) ([]*entities.JWTBlacklist, error)

	// List operations with pagination
	List(ctx context.Context, limit, offset int) ([]*entities.JWTBlacklist, int64, error)
	ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.JWTBlacklist, int64, error)
	ListByReason(ctx context.Context, reason string, limit, offset int) ([]*entities.JWTBlacklist, int64, error)

	// Maintenance operations
	DeleteExpired(ctx context.Context) (int64, error)
	DeleteByUser(ctx context.Context, userID uint) (int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)

	// Statistics
	CountTotal(ctx context.Context) (int64, error)
	CountByReason(ctx context.Context, reason string) (int64, error)
	CountByUser(ctx context.Context, userID uint) (int64, error)
	CountExpired(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)

	// Time-based queries
	GetExpiredBefore(ctx context.Context, before time.Time, limit int) ([]*entities.JWTBlacklist, error)
	GetCreatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*entities.JWTBlacklist, int64, error)
}