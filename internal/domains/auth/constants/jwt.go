package constants

// JWT Blacklist reason constants
const (
	BlacklistReasonLogout         = "logout"
	BlacklistReasonSecurityBreach = "security_breach"
	BlacklistReasonAdminRevoke    = "admin_revoke"
	BlacklistReasonPasswordChange = "password_change"
	BlacklistReasonAccountLocked  = "account_locked"
)