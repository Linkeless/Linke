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
	PaymentGatewayEpay   = "epay"
	PaymentGatewayEPUSDT = "epusdt"
	PaymentGatewayCrypto = "crypto" // 加密货币直接收款网关
)

// Payment Method Constants
const (
	// 加密货币支付方式
	PaymentMethodTRCUSDT     = "trc_usdt"      // TRC链的USDT
	PaymentMethodPolygonUSDT = "polygon_usdt"  // Polygon链的USDT
	PaymentMethodUSDT        = "usdt"          // 通用USDT
	PaymentMethodBTC         = "btc"           // 比特币
	PaymentMethodETH         = "eth"           // 以太坊
)

// Refund Status Constants
const (
	RefundStatusNone       = "none"
	RefundStatusProcessing = "processing"
	RefundStatusCompleted  = "completed"
	RefundStatusFailed     = "failed"
)

// Currency Constants
const (
	CurrencyCNY  = "CNY"
	CurrencyUSD  = "USD"
	CurrencyUSDT = "USDT"
)

// Payment Method Type Constants
const (
	PaymentMethodTypeCard          = "card"
	PaymentMethodTypeBankAccount   = "bank_account"
	PaymentMethodTypeDigitalWallet = "digital_wallet"
	PaymentMethodTypeCrypto        = "crypto"
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


// Payment Retry Status Constants
const (
	PaymentRetryStatusPending    = "pending"
	PaymentRetryStatusInProgress = "in_progress"
	PaymentRetryStatusCompleted  = "completed"
	PaymentRetryStatusFailed     = "failed"
	PaymentRetryStatusCancelled  = "cancelled"
)

// Retry Strategy Constants
const (
	RetryStrategyExponential = "exponential"
	RetryStrategyLinear      = "linear"
	RetryStrategyCustom      = "custom"
)

// Failure Type Constants
const (
	FailureTypeTemporary = "temporary"
	FailureTypePermanent = "permanent"
	FailureTypeNetwork   = "network"
	FailureTypeGateway   = "gateway"
	FailureTypeBusiness  = "business"
)

// Attempt Status Constants (for retry history)
const (
	AttemptStatusSuccess = "success"
	AttemptStatusFailed  = "failed"
	AttemptStatusTimeout = "timeout"
	AttemptStatusError   = "error"
)