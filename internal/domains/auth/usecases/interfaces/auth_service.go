package interfaces

import (
	"context"
	"linke/internal/domains/auth/dto"
	"linke/internal/domains/auth/entities"
	userEntities "linke/internal/domains/user/entities"
	"time"
)

// AuthService defines the interface for authentication service operations
type AuthService interface {
	// Registration and login
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	Logout(ctx context.Context, tokenString string, userID uint) error

	// Password management
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
	AdminResetPassword(ctx context.Context, adminUserID, targetUserID uint, newPassword string) error

	// Token validation
	ValidateToken(tokenString string) (*userEntities.User, error)

	// OAuth authentication (使用 dto.UserInfo 类型)
	CreateOrUpdateOAuthUser(ctx context.Context, userInfo *dto.UserInfo) (*userEntities.User, error)
}


// JWTService defines JWT token management operations (auth domain owned)
type JWTService interface {
	GenerateToken(user *userEntities.User) (*dto.TokenResponse, error)
	ValidateToken(tokenString string) (*dto.Claims, error)
	RefreshToken(tokenString string) (*dto.TokenResponse, error)
	RevokeToken(tokenString string, userID *uint, reason string) error
	RevokeAllUserTokens(userID uint, reason string) error
}

// LoginSecurityService defines login security operations (auth domain owned)
type LoginSecurityService interface {
	RecordLoginAttempt(ctx context.Context, email, ip, userAgent, reason string, success bool, userID *uint) error
	IsAccountLocked(ctx context.Context, email string) (bool, *entities.AccountLockout, error)
	GetFailureCount(ctx context.Context, email string) (int, error)
	UnlockAccount(ctx context.Context, email string, reason string) error
	GetLoginAttemptStats(ctx context.Context, since time.Time) (map[string]any, error)
	CleanupOldAttempts(ctx context.Context, olderThan time.Duration) error
}

// Note: UserService and InviteCodeService are imported from their respective domains
// - UserService from domains/user/usecases/interfaces
// - InviteCodeService from domains/referral/usecases/interfaces

