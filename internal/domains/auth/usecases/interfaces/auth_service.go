package interfaces

import (
	"context"
	"linke/internal/domains/auth/entities"
	referralEntities "linke/internal/domains/referral/entities"
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

// Service dependencies (interfaces that AuthService depends on)
type UserService interface {
	CreateUser(ctx context.Context, user *userEntities.User) error
	GetUserByEmail(ctx context.Context, email string) (*userEntities.User, error)
	GetUserByID(ctx context.Context, id uint) (*userEntities.User, error)
	UpdateUser(ctx context.Context, user *userEntities.User) error
}

type JWTService interface {
	GenerateToken(user *userEntities.User) (*TokenResponse, error)
	ValidateToken(tokenString string) (*Claims, error)
	RefreshToken(tokenString string) (*TokenResponse, error)
	RevokeToken(tokenString string, userID *uint, reason string) error
	RevokeAllUserTokens(userID uint, reason string) error
}

type InviteCodeService interface {
	ValidateInviteCode(ctx context.Context, code string) (*referralEntities.InviteCode, error)
	UseInviteCode(ctx context.Context, code string, userID uint, ipAddress, userAgent string) (*referralEntities.InviteCodeUsage, error)
}

type LoginSecurityService interface {
	RecordLoginAttempt(ctx context.Context, email, ip, userAgent, reason string, success bool, userID *uint) error
	IsAccountLocked(ctx context.Context, email string) (bool, *entities.AccountLockout, error)
	GetFailureCount(ctx context.Context, email string) (int, error)
	UnlockAccount(ctx context.Context, email string, reason string) error
	GetLoginAttemptStats(ctx context.Context, since time.Time) (map[string]interface{}, error)
	CleanupOldAttempts(ctx context.Context, olderThan time.Duration) error
}

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
