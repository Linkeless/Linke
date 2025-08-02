package entities

import (
	"time"

	"gorm.io/gorm"
)

// LoginAttempt represents a login attempt (successful or failed)
type LoginAttempt struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"index;size:255;not null"`
	IP        string    `json:"ip" gorm:"index;size:45;not null"` // IPv6 support
	UserAgent string    `json:"user_agent" gorm:"size:500"`
	Success   bool      `json:"success" gorm:"index;not null"`
	Reason    string    `json:"reason" gorm:"size:200"` // failure reason or success details
	UserID    *uint     `json:"user_id" gorm:"index"`   // null for failed attempts on non-existent users
	CreatedAt time.Time `json:"created_at" gorm:"not null;index"`
}

// TableName returns the table name for LoginAttempt model
func (LoginAttempt) TableName() string {
	return "login_attempts"
}

// Login attempt failure reasons
const (
	LoginFailureInvalidCredentials = "invalid_credentials"
	LoginFailureAccountLocked      = "account_locked"
	LoginFailureAccountInactive    = "account_inactive"
	LoginFailureAccountBanned      = "account_banned"
	LoginFailureUserNotFound       = "user_not_found"
	LoginFailureOAuthMismatch      = "oauth_mismatch"
	LoginFailureRateLimit          = "rate_limit"
	LoginSuccessLocal              = "local_login"
	LoginSuccessOAuth              = "oauth_login"
)

// BeforeCreate hook to set creation time
func (la *LoginAttempt) BeforeCreate(tx *gorm.DB) error {
	if la.CreatedAt.IsZero() {
		la.CreatedAt = time.Now()
	}
	return nil
}

// AccountLockout represents account lockout information
type AccountLockout struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"uniqueIndex;size:255;not null"`
	UserID       *uint     `json:"user_id" gorm:"index"`
	FailedCount  int       `json:"failed_count" gorm:"not null;default:0"`
	LastFailure  time.Time `json:"last_failure" gorm:"index"`
	LockedUntil  *time.Time `json:"locked_until" gorm:"index"`
	LockReason   string    `json:"lock_reason" gorm:"size:200"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null;index"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"not null"`
}

// TableName returns the table name for AccountLockout model
func (AccountLockout) TableName() string {
	return "account_lockouts"
}

// Account lockout reasons
const (
	LockReasonMultipleFailures = "multiple_failed_attempts"
	LockReasonSuspiciousActivity = "suspicious_activity"
	LockReasonAdminAction      = "admin_action"
	LockReasonSecurityBreach   = "security_breach"
)

// IsLocked checks if the account is currently locked
func (al *AccountLockout) IsLocked() bool {
	if al.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*al.LockedUntil)
}

// GetRemainingLockTime returns the remaining lock duration
func (al *AccountLockout) GetRemainingLockTime() time.Duration {
	if !al.IsLocked() {
		return 0
	}
	return al.LockedUntil.Sub(time.Now())
}

// BeforeCreate hook for AccountLockout
func (al *AccountLockout) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if al.CreatedAt.IsZero() {
		al.CreatedAt = now
	}
	if al.UpdatedAt.IsZero() {
		al.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate hook for AccountLockout
func (al *AccountLockout) BeforeUpdate(tx *gorm.DB) error {
	al.UpdatedAt = time.Now()
	return nil
}