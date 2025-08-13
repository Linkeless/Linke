package constants

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

// Account lockout reasons
const (
	LockReasonMultipleFailures   = "multiple_failed_attempts"
	LockReasonSuspiciousActivity = "suspicious_activity"
	LockReasonAdminAction        = "admin_action"
	LockReasonSecurityBreach     = "security_breach"
)
