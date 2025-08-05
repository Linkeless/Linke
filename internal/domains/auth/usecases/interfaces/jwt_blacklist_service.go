package interfaces

import (
	"context"
	"time"
)

// JWTBlacklistService defines the interface for JWT token blacklisting operations
type JWTBlacklistService interface {
	// Token blacklisting
	BlacklistToken(ctx context.Context, token string, userID *uint, reason string, expiresAt time.Time) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)

	// User-wide blacklisting
	BlacklistAllUserTokens(ctx context.Context, userID uint, reason string, tokenExpiresAt time.Time) error
	IsUserTokensBlacklisted(ctx context.Context, userID uint, tokenIssuedAt time.Time) (bool, error)

	// Maintenance and statistics
	CleanupExpiredEntries(ctx context.Context) error
	GetBlacklistStats(ctx context.Context) (map[string]any, error)
}
