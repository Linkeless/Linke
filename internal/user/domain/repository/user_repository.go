package repository

import (
	"context"

	"linke/internal/user/domain/aggregate"
	"linke/internal/user/domain/valueobject"
)

// UserRepository defines the interface for user persistence operations
type UserRepository interface {
	// Core CRUD operations
	Save(ctx context.Context, user *aggregate.User) error
	FindByID(ctx context.Context, id valueobject.UserID) (*aggregate.User, error)
	FindByEmail(ctx context.Context, email valueobject.Email) (*aggregate.User, error)
	Update(ctx context.Context, user *aggregate.User) error
	Delete(ctx context.Context, id valueobject.UserID) error

	// Query operations
	ExistsByEmail(ctx context.Context, email valueobject.Email) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	FindByUsername(ctx context.Context, username string) (*aggregate.User, error)

	// OAuth-specific operations
	FindByProviderID(ctx context.Context, provider string, providerID valueobject.ProviderID) (*aggregate.User, error)
	ExistsByProviderID(ctx context.Context, provider string, providerID valueobject.ProviderID) (bool, error)

	// List operations with pagination
	FindAll(ctx context.Context, offset, limit int) ([]*aggregate.User, error)
	FindByStatus(ctx context.Context, status valueobject.UserStatus, offset, limit int) ([]*aggregate.User, error)
	FindByRole(ctx context.Context, role valueobject.UserRole, offset, limit int) ([]*aggregate.User, error)
	FindByProvider(ctx context.Context, provider valueobject.Provider, offset, limit int) ([]*aggregate.User, error)

	// Count operations
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status valueobject.UserStatus) (int64, error)
	CountByRole(ctx context.Context, role valueobject.UserRole) (int64, error)
	CountByProvider(ctx context.Context, provider valueobject.Provider) (int64, error)

	// Search operations
	SearchByEmailOrUsername(ctx context.Context, query string, offset, limit int) ([]*aggregate.User, error)

	// Soft delete operations
	SoftDelete(ctx context.Context, id valueobject.UserID) error
	Restore(ctx context.Context, id valueobject.UserID) error
	FindDeleted(ctx context.Context, offset, limit int) ([]*aggregate.User, error)

	// Batch operations
	SaveBatch(ctx context.Context, users []*aggregate.User) error
	FindByIDs(ctx context.Context, ids []valueobject.UserID) ([]*aggregate.User, error)
}

// UserReadRepository defines read-only operations for queries
type UserReadRepository interface {
	FindByID(ctx context.Context, id valueobject.UserID) (*aggregate.User, error)
	FindByEmail(ctx context.Context, email valueobject.Email) (*aggregate.User, error)
	FindByUsername(ctx context.Context, username string) (*aggregate.User, error)
	FindByProviderID(ctx context.Context, provider string, providerID valueobject.ProviderID) (*aggregate.User, error)

	ExistsByEmail(ctx context.Context, email valueobject.Email) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByProviderID(ctx context.Context, provider string, providerID valueobject.ProviderID) (bool, error)

	FindAll(ctx context.Context, offset, limit int) ([]*aggregate.User, error)
	FindByStatus(ctx context.Context, status valueobject.UserStatus, offset, limit int) ([]*aggregate.User, error)
	FindByRole(ctx context.Context, role valueobject.UserRole, offset, limit int) ([]*aggregate.User, error)
	FindByProvider(ctx context.Context, provider valueobject.Provider, offset, limit int) ([]*aggregate.User, error)

	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status valueobject.UserStatus) (int64, error)
	CountByRole(ctx context.Context, role valueobject.UserRole) (int64, error)
	CountByProvider(ctx context.Context, provider valueobject.Provider) (int64, error)

	SearchByEmailOrUsername(ctx context.Context, query string, offset, limit int) ([]*aggregate.User, error)
	FindByIDs(ctx context.Context, ids []valueobject.UserID) ([]*aggregate.User, error)
}