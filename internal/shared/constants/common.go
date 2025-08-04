package constants

const (
	// 分页相关常量
	DefaultPageLimit = 10
	MaxPageLimit     = 100

	// 用户状态
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusBanned   = "banned"

	// 订阅状态
	SubscriptionStatusActive    = "active"
	SubscriptionStatusExpired   = "expired"
	SubscriptionStatusCancelled = "cancelled"
	SubscriptionStatusPaused    = "paused"

	// 支付状态
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"

	// 通用错误消息
	ErrInvalidRequestData     = "Invalid request data"
	ErrAuthenticationRequired = "Authentication required"
	ErrInvalidUserContext     = "Invalid user context"
	ErrResourceNotFound       = "Resource not found"
	ErrPermissionDenied       = "Permission denied"
	ErrInternalServerError    = "Internal server error"

	// HTTP参数名
	ParamID      = "id"
	ParamUserID  = "user_id"
	ParamOrderID = "order_id"
	ParamPlanID  = "plan_id"

	// 查询参数
	QueryLimit  = "limit"
	QueryOffset = "offset"
	QueryStatus = "status"
)
