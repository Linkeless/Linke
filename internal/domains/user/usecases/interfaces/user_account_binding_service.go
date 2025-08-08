package interfaces

import (
	"context"

	"linke/internal/domains/user/entities"
)

// UserAccountBindingService defines the interface for user account binding business operations
type UserAccountBindingService interface {
	// Basic binding operations
	CreateBinding(ctx context.Context, userID uint, req *entities.CreateBindingRequest) (*entities.UserAccountBinding, error)
	GetUserBindings(ctx context.Context, userID uint) ([]*entities.UserAccountBinding, error)
	UpdateBinding(ctx context.Context, userID uint, provider string, req *entities.UpdateBindingRequest) (*entities.UserAccountBinding, error)
	DeleteBinding(ctx context.Context, userID uint, provider string) error
	
	// OAuth integration
	FindUserByProviderAccount(ctx context.Context, provider, providerUserID string) (*entities.UserAccountBinding, error)
	SetPrimaryBinding(ctx context.Context, userID uint, provider string) error
}