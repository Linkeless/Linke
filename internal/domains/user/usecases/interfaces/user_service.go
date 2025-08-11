package interfaces

import (
	"context"
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
	BatchDeleteUsers(ctx context.Context, ids []uint) (*BatchOperationResult, error)
	BatchRestoreUsers(ctx context.Context, ids []uint) (*BatchOperationResult, error)

	// User management
	SoftDeleteUser(ctx context.Context, id uint) error
	RestoreUser(ctx context.Context, id uint) error
	HardDeleteUser(ctx context.Context, id uint) error
	UpdateUserStatus(ctx context.Context, id uint, status string) (*entities.User, error)
	UpdateUserRole(ctx context.Context, id uint, role string) (*entities.User, error)

	// Queries and statistics
	ListUsersFiltered(ctx context.Context, req *AdvancedUserSearchRequest) ([]*entities.User, int64, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error)
	ListDeletedUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error)
	ListUsersByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, int64, error)
	GetUserStats(ctx context.Context) (*UserStats, error)
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

// AdvancedUserSearchRequest represents advanced search parameters for users
type AdvancedUserSearchRequest struct {
	Query         string `form:"query" json:"query"`
	Status        string `form:"status" json:"status"`
	Provider      string `form:"provider" json:"provider"`
	Role          string `form:"role" json:"role"`
	EmailVerified *bool  `form:"email_verified" json:"email_verified"`
	Limit         int    `form:"limit" json:"limit" binding:"omitempty,min=1,max=100"`
	Offset        int    `form:"offset" json:"offset" binding:"omitempty,min=0"`
}
