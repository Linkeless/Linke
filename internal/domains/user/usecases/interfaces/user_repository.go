package interfaces

import (
	"context"
	"linke/internal/domains/user/entities"
)

// UserRepository defines the interface for user data access operations
type UserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id uint) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id uint) error

	// OAuth provider lookups
	GetByGoogleID(ctx context.Context, googleID string) (*entities.User, error)
	GetByGitHubID(ctx context.Context, githubID string) (*entities.User, error)
	GetByTelegramID(ctx context.Context, telegramID string) (*entities.User, error)

	// Active user operations (excludes soft deleted and inactive users)
	GetActiveByID(ctx context.Context, id uint) (*entities.User, error)
	GetActiveByEmail(ctx context.Context, email string) (*entities.User, error)

	// Soft delete operations
	SoftDelete(ctx context.Context, id uint) error
	Restore(ctx context.Context, id uint) error
	HardDelete(ctx context.Context, id uint) error

	// List operations with pagination
	List(ctx context.Context, limit, offset int) ([]*entities.User, int64, error)
	ListDeleted(ctx context.Context, limit, offset int) ([]*entities.User, int64, error)
	ListByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.User, int64, error)
	ListByRole(ctx context.Context, role string, limit, offset int) ([]*entities.User, int64, error)

	// Search operations
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.User, int64, error)

	// Status and role management
	UpdateStatus(ctx context.Context, id uint, status string) error
	UpdateRole(ctx context.Context, id uint, role string) error

	// Statistics
	CountTotal(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountByProvider(ctx context.Context, provider string) (int64, error)
	CountDeleted(ctx context.Context) (int64, error)
	CountRecentSignups(ctx context.Context, days int) (int64, error)

	// Batch operations
	BatchDelete(ctx context.Context, ids []uint) (int, []uint, error) // returns (deletedCount, failedIDs, error)
	BatchRestore(ctx context.Context, ids []uint) (int, []uint, error) // returns (restoredCount, failedIDs, error)

	// Existence checks
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)

	// Invite code related
	GetByInviteCodeUsed(ctx context.Context, inviteCode string) ([]*entities.User, error)
	CountByInviteCodeUsed(ctx context.Context, inviteCode string) (int64, error)
}