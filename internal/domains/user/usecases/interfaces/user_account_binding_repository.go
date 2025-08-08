package interfaces

import (
	"context"

	"linke/internal/domains/user/entities"
)

// UserAccountBindingRepository defines the interface for user account binding data operations
type UserAccountBindingRepository interface {
	// Basic CRUD
	Create(ctx context.Context, binding *entities.UserAccountBinding) error
	GetByID(ctx context.Context, id uint) (*entities.UserAccountBinding, error)
	Update(ctx context.Context, binding *entities.UserAccountBinding) error
	Delete(ctx context.Context, id uint) error
	
	// Core queries
	GetByUserID(ctx context.Context, userID uint) ([]*entities.UserAccountBinding, error)
	GetByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (*entities.UserAccountBinding, error)
	
	// Primary binding management
	SetPrimaryBinding(ctx context.Context, userID uint, bindingID uint) error
	UpdateLastUsed(ctx context.Context, id uint) error
}