package interfaces

import (
	"context"

	"linke/internal/domains/user/entities"
)

// UserAccountBindingRepository defines the interface for user account binding data operations
type UserAccountBindingRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, binding *entities.UserAccountBinding) error
	GetByID(ctx context.Context, id uint) (*entities.UserAccountBinding, error)
	Update(ctx context.Context, binding *entities.UserAccountBinding) error
	Delete(ctx context.Context, id uint) error

	// Query operations
	GetByUserID(ctx context.Context, userID uint) ([]*entities.UserAccountBinding, error)
	GetByUserIDAndProvider(ctx context.Context, userID uint, provider string) (*entities.UserAccountBinding, error)
	GetByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (*entities.UserAccountBinding, error)
	GetByProviderAndEmail(ctx context.Context, provider, email string) (*entities.UserAccountBinding, error)

	// Advanced queries
	GetPrimaryBindingByUserID(ctx context.Context, userID uint) (*entities.UserAccountBinding, error)
	ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]*entities.UserAccountBinding, int64, error)
	ExistsByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (bool, error)
	ExistsByUserIDAndProvider(ctx context.Context, userID uint, provider string) (bool, error)

	// Batch operations
	DeleteByUserID(ctx context.Context, userID uint) error
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	CountByProvider(ctx context.Context, provider string) (int64, error)

	// Primary binding management
	SetPrimaryBinding(ctx context.Context, userID uint, bindingID uint) error
	UnsetAllPrimaryBindings(ctx context.Context, userID uint) error

	// Maintenance operations
	UpdateLastUsed(ctx context.Context, id uint) error
	CleanupOldBindings(ctx context.Context, days int) error
}