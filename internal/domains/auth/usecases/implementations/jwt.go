package implementations

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/auth/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/config"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	cfg              *config.Config
	blacklistService *JWTBlacklistService
}

// Custom Claims struct for JWT parsing (includes jwt.RegisteredClaims)
type Claims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Provider string `json:"provider"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	jwt.RegisteredClaims
}

func NewJWTService(cfg *config.Config, blacklistService *JWTBlacklistService) *JWTService {
	return &JWTService{
		cfg:              cfg,
		blacklistService: blacklistService,
	}
}

// GenerateToken generates a JWT token for the given user
func (j *JWTService) GenerateToken(user *userEntities.User) (*interfaces.TokenResponse, error) {
	expirationTime := time.Now().Add(time.Duration(j.cfg.JWT.ExpireHours) * time.Hour)

	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Provider: user.Provider,
		Role:     user.Role,
		Status:   user.Status,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "linke-api",
			Subject:   fmt.Sprintf("user:%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &interfaces.TokenResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   j.cfg.JWT.ExpireHours * 3600, // Convert hours to seconds
		ExpiresAt:   expirationTime,
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTService) ValidateToken(tokenString string) (*interfaces.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check if token is blacklisted
		if j.blacklistService != nil {
			// Check specific token blacklist
			isBlacklisted, err := j.blacklistService.IsTokenBlacklisted(context.Background(), tokenString)
			if err != nil {
				return nil, fmt.Errorf("failed to check token blacklist: %w", err)
			}
			if isBlacklisted {
				return nil, fmt.Errorf("token has been revoked")
			}

			// Check user-wide blacklist
			isUserBlacklisted, err := j.blacklistService.IsUserTokensBlacklisted(context.Background(), claims.UserID, claims.IssuedAt.Time)
			if err != nil {
				return nil, fmt.Errorf("failed to check user token blacklist: %w", err)
			}
			if isUserBlacklisted {
				return nil, fmt.Errorf("all user tokens have been revoked")
			}
		}

		return &interfaces.Claims{
			UserID:    claims.UserID,
			Email:     claims.Email,
			Username:  claims.Username,
			Provider:  claims.Provider,
			Role:      claims.Role,
			Status:    claims.Status,
			IssuedAt:  claims.IssuedAt.Time,
			ExpiresAt: claims.ExpiresAt.Time,
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken generates a new token based on an existing valid token
func (j *JWTService) RefreshToken(tokenString string) (*interfaces.TokenResponse, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token for refresh: %w", err)
	}

	// Check if token is close to expiry (within 1 hour)
	if time.Until(claims.ExpiresAt) > time.Hour {
		return nil, fmt.Errorf("token is not close to expiry, no need to refresh")
	}

	// Create new token with updated expiration
	newExpirationTime := time.Now().Add(time.Duration(j.cfg.JWT.ExpireHours) * time.Hour)
	newClaims := &Claims{
		UserID:   claims.UserID,
		Email:    claims.Email,
		Username: claims.Username,
		Provider: claims.Provider,
		Role:     claims.Role,
		Status:   claims.Status,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(newExpirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenString, err := token.SignedString([]byte(j.cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &interfaces.TokenResponse{
		AccessToken: newTokenString,
		TokenType:   "Bearer",
		ExpiresIn:   j.cfg.JWT.ExpireHours * 3600,
		ExpiresAt:   newExpirationTime,
	}, nil
}

// RevokeToken adds a token to the blacklist
func (j *JWTService) RevokeToken(tokenString string, userID *uint, reason string) error {
	if j.blacklistService == nil {
		return fmt.Errorf("blacklist service not available")
	}

	// Parse token to get expiration time
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.cfg.JWT.Secret), nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse token for revocation: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return fmt.Errorf("invalid token claims")
	}

	// Add to blacklist until token expires
	return j.blacklistService.BlacklistToken(context.Background(), tokenString, userID, reason, claims.ExpiresAt.Time)
}

// RevokeAllUserTokens revokes all tokens for a specific user
func (j *JWTService) RevokeAllUserTokens(userID uint, reason string) error {
	if j.blacklistService == nil {
		return fmt.Errorf("blacklist service not available")
	}

	// Set expiration far in the future to cover all possible tokens
	expiresAt := time.Now().Add(time.Duration(j.cfg.JWT.ExpireHours) * time.Hour)

	return j.blacklistService.BlacklistAllUserTokens(context.Background(), userID, reason, expiresAt)
}
