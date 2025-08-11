package constants

// =============================================================================
// System-level Constants (truly generic across all domains)
// =============================================================================

const (
	// 分页相关常量
	DefaultPageLimit = 10
	MaxPageLimit     = 100

	// 通用错误消息
	ErrInvalidRequestData     = "Invalid request data"
	ErrAuthenticationRequired = "Authentication required"
	ErrInvalidUserContext     = "Invalid user context"
	ErrResourceNotFound       = "Resource not found"
	ErrPermissionDenied       = "Permission denied"
	ErrInternalServerError    = "Internal server error"

	// 通用HTTP参数名
	ParamID      = "id"
	ParamUserID  = "user_id"
	ParamOrderID = "order_id"

	// 通用查询参数
	QueryLimit  = "limit"
	QueryOffset = "offset"
	QueryStatus = "status"
)
