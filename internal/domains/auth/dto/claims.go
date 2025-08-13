package dto

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT token claims - public interface
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

// JWTClaims represents the internal JWT claims structure used with jwt.RegisteredClaims
// This is used internally by the JWT service for token parsing and generation
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Provider string `json:"provider"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	jwt.RegisteredClaims
}

// ToPublicClaims converts JWTClaims to public Claims
func (jc *JWTClaims) ToPublicClaims() *Claims {
	return &Claims{
		UserID:    jc.UserID,
		Email:     jc.Email,
		Username:  jc.Username,
		Provider:  jc.Provider,
		Role:      jc.Role,
		Status:    jc.Status,
		IssuedAt:  jc.IssuedAt.Time,
		ExpiresAt: jc.ExpiresAt.Time,
	}
}

// ToJWTClaims converts public Claims to JWTClaims for token generation
func (c *Claims) ToJWTClaims() *JWTClaims {
	return &JWTClaims{
		UserID:   c.UserID,
		Email:    c.Email,
		Username: c.Username,
		Provider: c.Provider,
		Role:     c.Role,
		Status:   c.Status,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(c.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(c.IssuedAt),
			NotBefore: jwt.NewNumericDate(c.IssuedAt),
			Issuer:    "linke-api",
		},
	}
}