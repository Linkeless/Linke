package service

import (
	"fmt"
	"time"

	"linke/config"
	"linke/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	cfg *config.Config
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Provider string `json:"provider"`
	jwt.RegisteredClaims
}

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		cfg: cfg,
	}
}

// GenerateToken generates a JWT token for the given user
func (j *JWTService) GenerateToken(user *model.User) (*TokenResponse, error) {
	expirationTime := time.Now().Add(time.Duration(j.cfg.JWT.ExpireHours) * time.Hour)

	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Provider: user.Provider,
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

	return &TokenResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   j.cfg.JWT.ExpireHours * 3600, // Convert hours to seconds
		ExpiresAt:   expirationTime,
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
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
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken generates a new token based on an existing valid token
func (j *JWTService) RefreshToken(tokenString string) (*TokenResponse, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token for refresh: %w", err)
	}

	// Check if token is close to expiry (within 1 hour)
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return nil, fmt.Errorf("token is not close to expiry, no need to refresh")
	}

	// Create new token with updated expiration
	newExpirationTime := time.Now().Add(time.Duration(j.cfg.JWT.ExpireHours) * time.Hour)
	claims.ExpiresAt = jwt.NewNumericDate(newExpirationTime)
	claims.IssuedAt = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newTokenString, err := token.SignedString([]byte(j.cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken: newTokenString,
		TokenType:   "Bearer",
		ExpiresIn:   j.cfg.JWT.ExpireHours * 3600,
		ExpiresAt:   newExpirationTime,
	}, nil
}