package interfaces

import (
	"context"
	"time"

	"linke/internal/domains/auth/entities"
)

// LoginAttemptRepository defines the interface for login attempt data access operations
type LoginAttemptRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, attempt *entities.LoginAttempt) error
	GetByID(ctx context.Context, id uint) (*entities.LoginAttempt, error)
	Delete(ctx context.Context, id uint) error

	// Query operations
	GetByEmail(ctx context.Context, email string, limit, offset int) ([]*entities.LoginAttempt, int64, error)
	GetByUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.LoginAttempt, int64, error)
	GetByIP(ctx context.Context, ip string, limit, offset int) ([]*entities.LoginAttempt, int64, error)
	GetSuccessfulAttempts(ctx context.Context, email string, limit, offset int) ([]*entities.LoginAttempt, int64, error)
	GetFailedAttempts(ctx context.Context, email string, limit, offset int) ([]*entities.LoginAttempt, int64, error)

	// Time-based queries
	GetRecentAttempts(ctx context.Context, email string, since time.Time) ([]*entities.LoginAttempt, error)
	GetRecentFailedAttempts(ctx context.Context, email string, since time.Time) ([]*entities.LoginAttempt, error)
	GetAttemptsInTimeRange(ctx context.Context, email string, start, end time.Time) ([]*entities.LoginAttempt, error)

	// Statistics
	CountTotalAttempts(ctx context.Context) (int64, error)
	CountSuccessfulAttempts(ctx context.Context) (int64, error)
	CountFailedAttempts(ctx context.Context) (int64, error)
	CountAttemptsByEmail(ctx context.Context, email string) (int64, error)
	CountFailedAttemptsByEmail(ctx context.Context, email string, since time.Time) (int64, error)
	CountAttemptsByIP(ctx context.Context, ip string, since time.Time) (int64, error)

	// Maintenance
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
	DeleteByEmail(ctx context.Context, email string) (int64, error)

	// List operations with pagination
	List(ctx context.Context, limit, offset int) ([]*entities.LoginAttempt, int64, error)
	ListBySuccess(ctx context.Context, success bool, limit, offset int) ([]*entities.LoginAttempt, int64, error)
	ListRecent(ctx context.Context, since time.Time, limit, offset int) ([]*entities.LoginAttempt, int64, error)
}

// AccountLockoutRepository defines the interface for account lockout data access operations
type AccountLockoutRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, lockout *entities.AccountLockout) error
	GetByID(ctx context.Context, id uint) (*entities.AccountLockout, error)
	GetByEmail(ctx context.Context, email string) (*entities.AccountLockout, error)
	Update(ctx context.Context, lockout *entities.AccountLockout) error
	Delete(ctx context.Context, id uint) error

	// Lockout status operations
	IsAccountLocked(ctx context.Context, email string) (bool, *entities.AccountLockout, error)
	IncrementFailureCount(ctx context.Context, email string, userID *uint) (*entities.AccountLockout, error)
	LockAccount(ctx context.Context, email string, userID *uint, duration time.Duration, reason string) (*entities.AccountLockout, error)
	UnlockAccount(ctx context.Context, email string, reason string) error
	ResetFailureCount(ctx context.Context, email string) error

	// Query operations
	GetLockedAccounts(ctx context.Context, limit, offset int) ([]*entities.AccountLockout, int64, error)
	GetAccountsWithFailures(ctx context.Context, minFailures int, limit, offset int) ([]*entities.AccountLockout, int64, error)
	GetByUser(ctx context.Context, userID uint) (*entities.AccountLockout, error)

	// Statistics
	CountTotalLockouts(ctx context.Context) (int64, error)
	CountActiveLockouts(ctx context.Context) (int64, error)
	CountLockoutsByReason(ctx context.Context, reason string) (int64, error)
	GetLockoutStats(ctx context.Context, since time.Time) (map[string]int64, error)

	// Maintenance
	CleanupExpiredLockouts(ctx context.Context) (int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)

	// List operations with pagination
	List(ctx context.Context, limit, offset int) ([]*entities.AccountLockout, int64, error)
	ListByReason(ctx context.Context, reason string, limit, offset int) ([]*entities.AccountLockout, int64, error)
	ListRecent(ctx context.Context, since time.Time, limit, offset int) ([]*entities.AccountLockout, int64, error)
}
