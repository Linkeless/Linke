package interfaces

import (
	"context"
	"linke/internal/domains/user/entities"
	"linke/internal/shared/framework"
)

// UserRepository defines the interface for user-specific data access operations
// It extends GenericRepository with User-specific methods
type UserRepository interface {
	framework.GenericRepository[entities.User, uint]
	
	// User-specific query methods
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	
	// OAuth provider lookups
	GetByGoogleID(ctx context.Context, googleID string) (*entities.User, error)
	GetByGitHubID(ctx context.Context, githubID string) (*entities.User, error)
	GetByTelegramID(ctx context.Context, telegramID string) (*entities.User, error)

	// Active user operations (excludes soft deleted and inactive users)
	GetActiveByID(ctx context.Context, id uint) (*entities.User, error)
	GetActiveByEmail(ctx context.Context, email string) (*entities.User, error)

	// User-specific list operations
	ListByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error)
	ListByRole(ctx context.Context, role string, limit, offset int) ([]*entities.User, int64, error)

	// Role management (in addition to status management from base)
	UpdateRole(ctx context.Context, id uint, role string) error

	// User-specific statistics
	CountByProvider(ctx context.Context, provider string) (int64, error)
	CountRecentSignups(ctx context.Context, days int) (int64, error)

	// User-specific existence checks
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// Invite code related
	GetByInviteCodeUsed(ctx context.Context, inviteCode string) ([]*entities.User, error)
	CountByInviteCodeUsed(ctx context.Context, inviteCode string) (int64, error)
}
