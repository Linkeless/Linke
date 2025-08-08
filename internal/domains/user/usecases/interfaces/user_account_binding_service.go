package interfaces

import (
	"context"

	"linke/internal/domains/user/entities"
)

// UserAccountBindingService defines the interface for user account binding business operations
type UserAccountBindingService interface {
	// Core binding operations
	CreateBinding(ctx context.Context, userID uint, req *entities.CreateBindingRequest) (*entities.UserAccountBinding, error)
	GetUserBindings(ctx context.Context, userID uint) ([]*entities.UserAccountBinding, error)
	GetUserBinding(ctx context.Context, userID uint, provider string) (*entities.UserAccountBinding, error)
	UpdateBinding(ctx context.Context, userID uint, provider string, req *entities.UpdateBindingRequest) (*entities.UserAccountBinding, error)
	DeleteBinding(ctx context.Context, userID uint, provider string) error

	// OAuth integration - find user by provider account
	FindUserByProviderAccount(ctx context.Context, provider, providerUserID string) (*entities.UserAccountBinding, error)
	FindUserByProviderEmail(ctx context.Context, provider, email string) (*entities.UserAccountBinding, error)

	// Primary binding management
	SetPrimaryBinding(ctx context.Context, userID uint, provider string) error
	GetPrimaryBinding(ctx context.Context, userID uint) (*entities.UserAccountBinding, error)

	// Validation and checks
	ValidateBindingRequest(req *entities.CreateBindingRequest) error
	CanBindProvider(ctx context.Context, userID uint, provider string) error
	IsProviderAccountBound(ctx context.Context, provider, providerUserID string) (bool, error)

	// OAuth user creation/update
	CreateOrUpdateFromOAuth(ctx context.Context, provider string, userInfo *OAuthUserInfo) (*entities.UserAccountBinding, error)
	UpdateLastUsed(ctx context.Context, bindingID uint) error

	// Administrative operations
	ListBindings(ctx context.Context, userID uint, offset, limit int) ([]*entities.UserAccountBinding, int64, error)
	GetBindingStats(ctx context.Context) (*BindingStats, error)
	CleanupInactiveBindings(ctx context.Context, days int) error
}

// OAuthUserInfo represents OAuth user information for binding
type OAuthUserInfo struct {
	UserID           uint    `json:"user_id"`
	Provider         string  `json:"provider"`
	ProviderUserID   string  `json:"provider_user_id"`
	ProviderEmail    *string `json:"provider_email,omitempty"`
	ProviderUsername *string `json:"provider_username,omitempty"`
	ProviderName     *string `json:"provider_name,omitempty"`
	ProviderAvatar   *string `json:"provider_avatar,omitempty"`
	ProviderData     *string `json:"provider_data,omitempty"`
}

// BindingStats represents binding statistics
type BindingStats struct {
	TotalBindings    int64            `json:"total_bindings"`
	BindingsByProvider map[string]int64 `json:"bindings_by_provider"`
	UsersWithBindings int64            `json:"users_with_bindings"`
	ActiveBindings   int64            `json:"active_bindings"`
}