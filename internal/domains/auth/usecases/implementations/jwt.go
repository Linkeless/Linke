package implementations

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/auth/dto"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/config"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	cfg              *config.Config
	blacklistService *JWTBlacklistService
	currentSecret    []byte
	previousSecret   []byte // For supporting secret rotation
}

// JWTClaims struct for internal JWT parsing (includes jwt.RegisteredClaims)
// This is different from the public dto.Claims as it includes jwt.RegisteredClaims
type JWTClaims struct {
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
		currentSecret:    []byte(cfg.JWT.Secret),
		previousSecret:   nil, // Will be set during secret rotation
	}
}

// GenerateToken generates a JWT token for the given user
func (j *JWTService) GenerateToken(user *userEntities.User) (*dto.TokenResponse, error) {
	expirationTime := time.Now().Add(time.Duration(j.cfg.JWT.ExpireHours) * time.Hour)

	claims := &JWTClaims{
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
	tokenString, err := token.SignedString(j.currentSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Use object pool to reduce memory allocations
	tokenResponse := dto.GetTokenResponse()
	tokenResponse.AccessToken = tokenString
	tokenResponse.TokenType = "Bearer"
	tokenResponse.ExpiresIn = j.cfg.JWT.ExpireHours * 3600 // Convert hours to seconds
	tokenResponse.ExpiresAt = expirationTime

	return tokenResponse, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTService) ValidateToken(tokenString string) (*dto.Claims, error) {
	// Try current secret first
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.currentSecret, nil
	})

	// If parsing with current secret fails and we have a previous secret, try it
	if err != nil && j.previousSecret != nil {
		token, err = jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return j.previousSecret, nil
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Check if token is blacklisted
		if j.blacklistService != nil {
			// Check specific token blacklist
			isBlacklisted, err := j.blacklistService.IsTokenBlacklisted(context.Background(), tokenString)
			if err != nil {
				return nil, dto.NewAuthErrorWithCause(dto.ErrorTypeBlacklistFailure, 
					"Failed to check token blacklist", err).WithTokenHash(tokenString[:8])
			}
			if isBlacklisted {
				return nil, dto.ErrRevokedToken().WithTokenHash(tokenString[:8])
			}

			// Check user-wide blacklist
			isUserBlacklisted, err := j.blacklistService.IsUserTokensBlacklisted(context.Background(), claims.UserID, claims.IssuedAt.Time)
			if err != nil {
				return nil, dto.NewAuthErrorWithCause(dto.ErrorTypeBlacklistFailure,
					"Failed to check user token blacklist", err).WithUserID(claims.UserID)
			}
			if isUserBlacklisted {
				return nil, dto.ErrRevokedToken().WithUserID(claims.UserID).WithDetails("all user tokens revoked")
			}
		}

		// Use object pool to reduce memory allocations
		dtoClaims := dto.GetClaims()
		dtoClaims.UserID = claims.UserID
		dtoClaims.Email = claims.Email
		dtoClaims.Username = claims.Username
		dtoClaims.Provider = claims.Provider
		dtoClaims.Role = claims.Role
		dtoClaims.Status = claims.Status
		dtoClaims.IssuedAt = claims.IssuedAt.Time
		dtoClaims.ExpiresAt = claims.ExpiresAt.Time
		
		return dtoClaims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken generates a new token based on an existing valid token
func (j *JWTService) RefreshToken(tokenString string) (*dto.TokenResponse, error) {
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
	newClaims := &JWTClaims{
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
	newTokenString, err := token.SignedString(j.currentSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Use object pool to reduce memory allocations
	tokenResponse := dto.GetTokenResponse()
	tokenResponse.AccessToken = newTokenString
	tokenResponse.TokenType = "Bearer"
	tokenResponse.ExpiresIn = j.cfg.JWT.ExpireHours * 3600
	tokenResponse.ExpiresAt = newExpirationTime

	return tokenResponse, nil
}

// RevokeToken adds a token to the blacklist
func (j *JWTService) RevokeToken(tokenString string, userID *uint, reason string) error {
	if j.blacklistService == nil {
		return fmt.Errorf("blacklist service not available")
	}

	// Parse token to get expiration time - try current secret first
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		return j.currentSecret, nil
	})

	// If parsing with current secret fails and we have a previous secret, try it
	if err != nil && j.previousSecret != nil {
		token, err = jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
			return j.previousSecret, nil
		})
	}

	if err != nil {
		return fmt.Errorf("failed to parse token for revocation: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
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

// RotateSecret rotates the JWT signing secret while maintaining backward compatibility
func (j *JWTService) RotateSecret(newSecret string) {
	j.previousSecret = j.currentSecret
	j.currentSecret = []byte(newSecret)
}
