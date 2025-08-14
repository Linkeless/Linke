package constants

// User Status Constants
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusBanned   = "banned"
)

// User Role Constants
const (
	UserRoleUser  = "user"
	UserRoleAdmin = "admin"
)

// Provider Constants
const (
	ProviderLocal    = "local"
	ProviderGoogle   = "google"
	ProviderGitHub   = "github"
	ProviderTelegram = "telegram"
)

// Account Binding Provider Constants
const (
	BindingProviderGoogle   = "google"
	BindingProviderGitHub   = "github"
	BindingProviderTelegram = "telegram"
)

// Security Event Type Constants
const (
	SecurityEventSuspiciousBinding  = "suspicious_binding"
	SecurityEventDuplicateAttempt   = "duplicate_attempt"
	SecurityEventRapidBinding       = "rapid_binding"
	SecurityEventUnusualProvider    = "unusual_provider"
	SecurityEventFailedValidation   = "failed_validation"
	SecurityEventMassBatchOperation = "mass_batch_operation"
)

// Security Severity Level Constants
const (
	SecuritySeverityLow      = "low"
	SecuritySeverityMedium   = "medium"
	SecuritySeverityHigh     = "high"
	SecuritySeverityCritical = "critical"
)

// Audit Operation Constants
const (
	AuditOperationCreate = "create"
	AuditOperationUpdate = "update"
	AuditOperationDelete = "delete"
	AuditOperationBind   = "bind"
	AuditOperationUnbind = "unbind"
	AuditOperationLogin  = "login"
)

// Audit Status Constants
const (
	AuditStatusSuccess = "success"
	AuditStatusFailure = "failure"
	AuditStatusWarning = "warning"
)

// Helper functions for validation

// ValidProviders returns a list of valid provider names
func ValidProviders() []string {
	return []string{
		ProviderLocal,
		ProviderGoogle,
		ProviderGitHub,
		ProviderTelegram,
	}
}

// ValidBindingProviders returns a list of valid binding provider names
func ValidBindingProviders() []string {
	return []string{
		BindingProviderGoogle,
		BindingProviderGitHub,
		BindingProviderTelegram,
	}
}

// ValidUserStatuses returns a list of valid user statuses
func ValidUserStatuses() []string {
	return []string{
		UserStatusActive,
		UserStatusInactive,
		UserStatusBanned,
	}
}

// ValidUserRoles returns a list of valid user roles
func ValidUserRoles() []string {
	return []string{
		UserRoleUser,
		UserRoleAdmin,
	}
}

// ValidSecurityEventTypes returns a list of valid security event types
func ValidSecurityEventTypes() []string {
	return []string{
		SecurityEventSuspiciousBinding,
		SecurityEventDuplicateAttempt,
		SecurityEventRapidBinding,
		SecurityEventUnusualProvider,
		SecurityEventFailedValidation,
		SecurityEventMassBatchOperation,
	}
}

// ValidSecuritySeverityLevels returns a list of valid security severity levels
func ValidSecuritySeverityLevels() []string {
	return []string{
		SecuritySeverityLow,
		SecuritySeverityMedium,
		SecuritySeverityHigh,
		SecuritySeverityCritical,
	}
}
