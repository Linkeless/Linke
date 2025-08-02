package interfaces

import (
	"context"
	"linke/internal/domains/user/entities"
)

// UserService defines the interface for user service operations
type UserService interface {
	// User CRUD operations
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByID(ctx context.Context, id uint) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error

	// Active user operations
	GetActiveUserByID(ctx context.Context, id uint) (*entities.User, error)
	GetActiveUserByEmail(ctx context.Context, email string) (*entities.User, error)

	// Soft delete operations
	SoftDeleteUser(ctx context.Context, id uint) error
	RestoreUser(ctx context.Context, id uint) error
	HardDeleteUser(ctx context.Context, id uint) error

	// List operations
	ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error)
	ListDeletedUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error)
	ListUsersByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error)

	// Search operations
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, int64, error)

	// Status and role management
	UpdateUserStatus(ctx context.Context, id uint, status string) (*entities.User, error)
	UpdateUserRole(ctx context.Context, id uint, role string) (*entities.User, error)

	// Statistics
	GetUserStats(ctx context.Context) (*UserStats, error)

	// Batch operations
	BatchDeleteUsers(ctx context.Context, ids []uint) (*BatchOperationResult, error)
	BatchRestoreUsers(ctx context.Context, ids []uint) (*BatchOperationResult, error)
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers    int64            `json:"total_users"`
	ActiveUsers   int64            `json:"active_users"`
	InactiveUsers int64            `json:"inactive_users"`
	BannedUsers   int64            `json:"banned_users"`
	DeletedUsers  int64            `json:"deleted_users"`
	ByProvider    map[string]int64 `json:"by_provider"`
	RecentSignups int64            `json:"recent_signups"`
}

// BatchOperationResult represents the result of batch operations
type BatchOperationResult struct {
	DeletedCount  int    `json:"deleted_count,omitempty"`
	RestoredCount int    `json:"restored_count,omitempty"`
	FailedIDs     []uint `json:"failed_ids,omitempty"`
}
