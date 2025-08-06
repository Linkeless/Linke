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

	// 服务器状态
	ServerStatusActive      = "active"
	ServerStatusInactive    = "inactive"
	ServerStatusMaintenance = "maintenance"

	// 查询参数
	QueryLimit  = "limit"
	QueryOffset = "offset"
	QueryStatus = "status"

	// HTTP参数名 - 服务器相关
	ParamServerID = "server_id"
	ParamGroupID  = "group_id"

	// 服务器错误消息
	ErrServerNotFound       = "Server not found"
	ErrServerGroupNotFound  = "Server group not found"
	ErrInvalidServerID      = "Invalid server ID"
	ErrInvalidGroupID       = "Invalid group ID"
	ErrServerCreateFailed   = "Failed to create server"
	ErrServerUpdateFailed   = "Failed to update server"
	ErrServerDeleteFailed   = "Failed to delete server"
	ErrGroupCreateFailed    = "Failed to create server group"
	ErrGroupUpdateFailed    = "Failed to update server group"
	ErrGroupDeleteFailed    = "Failed to delete server group"
	ErrGroupHasServers      = "Cannot delete group that contains servers"
	ErrServerHealthCheck    = "Failed to check server health"
	ErrServerStatistics     = "Failed to get server statistics"
)
