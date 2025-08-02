package entities

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

// JWTBlacklist represents a blacklisted JWT token
type JWTBlacklist struct {
	// Primary key is the token hash
	TokenHash string    `json:"token_hash" gorm:"primaryKey;size:64"`
	UserID    *uint     `json:"user_id" gorm:"index"`
	Reason    string    `json:"reason" gorm:"size:100;not null"` // logout, security_breach, admin_revoke
	ExpiresAt time.Time `json:"expires_at" gorm:"index;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;index"`
}

// TableName returns the table name for JWTBlacklist model
func (JWTBlacklist) TableName() string {
	return "jwt_blacklist"
}

// Blacklist reason constants
const (
	BlacklistReasonLogout          = "logout"
	BlacklistReasonSecurityBreach  = "security_breach"
	BlacklistReasonAdminRevoke     = "admin_revoke"
	BlacklistReasonPasswordChange  = "password_change"
	BlacklistReasonAccountLocked   = "account_locked"
)

// HashToken creates a SHA256 hash of the JWT token for storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// IsExpired checks if the blacklist entry is expired
func (jb *JWTBlacklist) IsExpired() bool {
	return time.Now().After(jb.ExpiresAt)
}

// BeforeCreate hook to ensure token hash is set
func (jb *JWTBlacklist) BeforeCreate(tx *gorm.DB) error {
	if jb.CreatedAt.IsZero() {
		jb.CreatedAt = time.Now()
	}
	return nil
}