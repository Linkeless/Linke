package constants

// Payment Record Status Constants
const (
	PaymentRecordStatusPending    = "pending"
	PaymentRecordStatusProcessing = "processing"
	PaymentRecordStatusCompleted  = "completed"
	PaymentRecordStatusFailed     = "failed"
	PaymentRecordStatusCancelled  = "cancelled"
	PaymentRecordStatusRefunded   = "refunded"
)

// Payment Gateway Constants
const (
	PaymentGatewayEpay = "epay" // 易支付网关
)

// Payment Method Constants - 支持的epay支付方式
const (
	PaymentMethodAlipay = "alipay" // 支付宝
	PaymentMethodWechat = "wechat" // 微信支付
	PaymentMethodQQ     = "qqpay"  // QQ钱包
)

// Refund Status Constants
const (
	RefundStatusNone       = "none"
	RefundStatusProcessing = "processing"
	RefundStatusCompleted  = "completed"
	RefundStatusFailed     = "failed"
)

// Currency Constants - 只保留系统支持的CNY货币
const (
	CurrencyCNY = "CNY"
)

// Payment Method Type Constants
const (
	PaymentMethodTypeCard          = "card"
	PaymentMethodTypeBankAccount   = "bank_account"
	PaymentMethodTypeDigitalWallet = "digital_wallet"
)

// Payment Method Status Constants
const (
	PaymentMethodStatusActive   = "active"
	PaymentMethodStatusInactive = "inactive"
	PaymentMethodStatusExpired  = "expired"
	PaymentMethodStatusInvalid  = "invalid"
)

// Fee Type Constants
const (
	FeeTypeNone       = "none"
	FeeTypeFixed      = "fixed"
	FeeTypePercentage = "percentage"
)
