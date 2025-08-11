package interfaces

import (
	"context"
	"linke/internal/domains/auth/entities"
	userEntities "linke/internal/domains/user/entities"
	"time"
)

// AuthService defines the interface for authentication service operations
type AuthService interface {
	// Registration and login
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	Logout(ctx context.Context, tokenString string, userID uint) error

	// Password management
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
	AdminResetPassword(ctx context.Context, adminUserID, targetUserID uint, newPassword string) error

	// Token validation
	ValidateToken(tokenString string) (*userEntities.User, error)

	// OAuth authentication (使用 oauth_service.UserInfo 类型)
	CreateOrUpdateOAuthUser(ctx context.Context, userInfo *UserInfo) (*userEntities.User, error)
}

// RegisterRequest represents registration request data
type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	InviteCode string `json:"invite_code"` // Optional invite code
}

// LoginRequest represents login request data
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User  *userEntities.UserResponse `json:"user"`
	Token *TokenResponse             `json:"token"`
}

// JWTService defines JWT token management operations (auth domain owned)
type JWTService interface {
	GenerateToken(user *userEntities.User) (*TokenResponse, error)
	ValidateToken(tokenString string) (*Claims, error)
	RefreshToken(tokenString string) (*TokenResponse, error)
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

// Claims represents JWT token claims
type Claims struct {
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Provider  string    `json:"provider"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// TokenResponse represents token information
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

// UserInfo represents OAuth user information from providers
type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Provider string `json:"provider"`
}
