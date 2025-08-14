package interfaces

import (
	"context"

	"linke/internal/domains/user/entities"
	"linke/internal/shared/framework"
)

// UserRepository defines the interface for user data access operations
// It extends GenericRepository with streamlined user-specific methods
type UserRepository interface {
	framework.GenericRepository[entities.User, uint]

	// Generic field-based queries (replaces GetByEmail, GetByGoogleID, GetByGitHubID, GetByTelegramID)
	GetByField(ctx context.Context, field string, value interface{}) (*entities.User, error)
	GetActiveByField(ctx context.Context, field string, value interface{}) (*entities.User, error)
	ExistsByField(ctx context.Context, field string, value interface{}) (bool, error)

	// Generic listing with filters
	ListByField(ctx context.Context, field string, value interface{}, limit, offset int) ([]*entities.User, int64, error)

	// Statistics with generic counting
	CountByField(ctx context.Context, field string, value interface{}) (int64, error)
	CountRecentSignups(ctx context.Context, days int) (int64, error)

	// Batch operations by field
	GetMultipleByField(ctx context.Context, field string, values []interface{}) ([]*entities.User, error)

	// Additional methods from implementation (now part of interface)
	ListByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error)
	ListByRole(ctx context.Context, role string, limit, offset int) ([]*entities.User, int64, error)
	UpdateRole(ctx context.Context, id uint, role string) error
	GetByInviteCodeUsed(ctx context.Context, inviteCode string) ([]*entities.User, error)
}
