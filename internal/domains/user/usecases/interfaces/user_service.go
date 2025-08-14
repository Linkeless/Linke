package interfaces

import (
	"context"

	"linke/internal/domains/user/dto"
	"linke/internal/domains/user/entities"
)

// UserService provides core user management operations
type UserService interface {
	// Core CRUD operations
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByID(ctx context.Context, id uint) (*entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error

	// Domain-specific lookups
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID string) (*entities.User, error)
	GetActiveUserByID(ctx context.Context, id uint) (*entities.User, error)
	GetActiveUserByEmail(ctx context.Context, email string) (*entities.User, error)

	// Batch operations
	GetUsersByIDs(ctx context.Context, ids []uint) ([]*entities.User, error)
	BatchDeleteUsers(ctx context.Context, ids []uint) (*dto.BatchOperationResult, error)
	BatchRestoreUsers(ctx context.Context, ids []uint) (*dto.BatchOperationResult, error)

	// User management
	SoftDeleteUser(ctx context.Context, id uint) error
	RestoreUser(ctx context.Context, id uint) error
	HardDeleteUser(ctx context.Context, id uint) error
	UpdateUserStatus(ctx context.Context, id uint, status string) (*entities.User, error)
	UpdateUserRole(ctx context.Context, id uint, role string) (*entities.User, error)

	// Queries and statistics
	ListUsersFiltered(ctx context.Context, req *dto.AdvancedUserSearchRequest) ([]*entities.User, int64, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error)
	ListDeletedUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error)
	ListUsersByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, int64, error)
	GetUserStats(ctx context.Context) (*dto.UserStats, error)
}
