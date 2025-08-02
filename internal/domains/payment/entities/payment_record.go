package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/shared/dto"
)

// PaymentRecord represents a payment transaction record
type PaymentRecord struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	UserID              uint  `json:"user_id" gorm:"not null;index"`
	SubscriptionOrderID *uint `json:"subscription_order_id,omitempty" gorm:"index"` // 关联的订阅订单
	InvoiceID           *uint `json:"invoice_id,omitempty" gorm:"index"`             // 关联的发票

	// Payment Information
	PaymentNo     string `json:"payment_no" gorm:"uniqueIndex;size:100;not null"`   // 支付单号（系统生成）
	OutTradeNo    string `json:"out_trade_no" gorm:"uniqueIndex;size:100;not null"` // 商户订单号
	TransactionID string `json:"transaction_id" gorm:"size:100;index"`              // 第三方交易号
	Gateway       string `json:"gateway" gorm:"size:50;not null;index"`             // 支付网关 (epay, epusdt)
	PaymentMethod string `json:"payment_method" gorm:"size:50;not null;index"`      // 支付方式 (alipay, wechat, usdt, etc.)

	// Amount Information
	Amount       float64 `json:"amount" gorm:"type:decimal(10,2);not null"`         // 支付金额
	Currency     string  `json:"currency" gorm:"size:10;not null;default:'CNY'"`    // 货币类型
	ExchangeRate float64 `json:"exchange_rate" gorm:"type:decimal(10,4);default:1"` // 汇率（相对于基础货币）

	// Status Information
	Status        string `json:"status" gorm:"size:20;not null;default:'pending';index"` // pending, processing, completed, failed, cancelled, refunded
	PaymentStatus string `json:"payment_status" gorm:"size:20;index"`                    // 第三方支付状态

	// Gateway Response Data
	GatewayResponse string `json:"gateway_response,omitempty" gorm:"type:text"` // 网关返回的原始数据
	PaymentURL      string `json:"payment_url,omitempty" gorm:"size:500"`       // 支付链接
	QRCodeURL       string `json:"qr_code_url,omitempty" gorm:"size:500"`       // 二维码链接

	// Timing Information
	ExpiredAt  *time.Time `json:"expired_at,omitempty" gorm:"index"`  // 支付过期时间
	PaidAt     *time.Time `json:"paid_at,omitempty" gorm:"index"`     // 支付完成时间
	NotifiedAt *time.Time `json:"notified_at,omitempty" gorm:"index"` // 回调通知时间

	// Idempotency Protection
	LastNotifyHash string `json:"last_notify_hash,omitempty" gorm:"size:64;index"` // 最后一次通知的hash，用于防重放
	NotifyCount    int    `json:"notify_count" gorm:"default:0"`                   // 通知次数，用于检测重复通知

	// Refund Information
	RefundAmount float64    `json:"refund_amount" gorm:"type:decimal(10,2);default:0"` // 退款金额
	RefundStatus string     `json:"refund_status,omitempty" gorm:"size:20;index"`      // none, processing, completed, failed
	RefundedAt   *time.Time `json:"refunded_at,omitempty" gorm:"index"`                // 退款时间
	RefundReason string     `json:"refund_reason,omitempty" gorm:"size:255"`           // 退款原因

	// Additional Information
	ClientIP  string `json:"client_ip,omitempty" gorm:"size:45"`   // 客户端IP
	UserAgent string `json:"user_agent,omitempty" gorm:"size:500"` // 用户代理
	NotifyURL string `json:"notify_url,omitempty" gorm:"size:500"` // 异步通知地址
	ReturnURL string `json:"return_url,omitempty" gorm:"size:500"` // 同步返回地址

	// Metadata
	Metadata string `json:"metadata,omitempty" gorm:"type:text"` // 额外元数据(JSON)
	Remark   string `json:"remark,omitempty" gorm:"size:255"`    // 备注信息

	// Note: Relationships removed to avoid cross-domain dependencies
	// Related data should be fetched and assembled at the application layer

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for PaymentRecord model
func (PaymentRecord) TableName() string {
	return "payment_records"
}

// Payment status constants
const (
	PaymentRecordStatusPending    = "pending"
	PaymentRecordStatusProcessing = "processing"
	PaymentRecordStatusCompleted  = "completed"
	PaymentRecordStatusFailed     = "failed"
	PaymentRecordStatusCancelled  = "cancelled"
	PaymentRecordStatusRefunded   = "refunded"
)

// Payment gateway constants
const (
	PaymentGatewayEpay   = "epay"
	PaymentGatewayEPUSDT = "epusdt"
)

// Payment method constants
const (
	PaymentMethodAlipay   = "alipay"
	PaymentMethodWechat   = "wechat"
	PaymentMethodQQ       = "qqpay"
	PaymentMethodUnionPay = "unionpay"
	PaymentMethodUSDT     = "usdt"
	PaymentMethodBTC      = "btc"
	PaymentMethodETH      = "eth"
)

// Refund status constants
const (
	RefundStatusNone       = "none"
	RefundStatusProcessing = "processing"
	RefundStatusCompleted  = "completed"
	RefundStatusFailed     = "failed"
)

// Currency constants
const (
	CurrencyCNY  = "CNY"
	CurrencyUSD  = "USD"
	CurrencyUSDT = "USDT"
)

// IsPending checks if the payment is pending
func (pr *PaymentRecord) IsPending() bool {
	return pr.Status == PaymentRecordStatusPending
}

// IsProcessing checks if the payment is processing
func (pr *PaymentRecord) IsProcessing() bool {
	return pr.Status == PaymentRecordStatusProcessing
}

// IsCompleted checks if the payment is completed
func (pr *PaymentRecord) IsCompleted() bool {
	return pr.Status == PaymentRecordStatusCompleted
}

// IsFailed checks if the payment has failed
func (pr *PaymentRecord) IsFailed() bool {
	return pr.Status == PaymentRecordStatusFailed
}

// IsCancelled checks if the payment is cancelled
func (pr *PaymentRecord) IsCancelled() bool {
	return pr.Status == PaymentRecordStatusCancelled
}

// IsRefunded checks if the payment is refunded
func (pr *PaymentRecord) IsRefunded() bool {
	return pr.Status == PaymentRecordStatusRefunded
}

// IsExpired checks if the payment has expired
func (pr *PaymentRecord) IsExpired() bool {
	return pr.ExpiredAt != nil && pr.ExpiredAt.Before(time.Now())
}

// CanBeRefunded checks if the payment can be refunded
func (pr *PaymentRecord) CanBeRefunded() bool {
	return pr.IsCompleted() && pr.RefundStatus != RefundStatusCompleted
}

// GetRefundableAmount returns the amount that can be refunded
func (pr *PaymentRecord) GetRefundableAmount() float64 {
	if !pr.CanBeRefunded() {
		return 0
	}
	return pr.Amount - pr.RefundAmount
}

// IsDeleted checks if the payment record is soft deleted
func (pr *PaymentRecord) IsDeleted() bool {
	return pr.DeletedAt.Valid
}

// PaymentRecordResponse represents the payment record data structure for API responses
type PaymentRecordResponse struct {
	ID                  uint       `json:"id" example:"1"`                                          // Payment ID
	UserID              uint       `json:"user_id" example:"1"`                                     // User ID
	SubscriptionOrderID *uint      `json:"subscription_order_id,omitempty" example:"1"`             // Subscription order ID
	PaymentNo           string     `json:"payment_no" example:"PAY202401010001"`                    // Payment number
	OutTradeNo          string     `json:"out_trade_no" example:"ORDER202401010001"`                // Merchant order number
	TransactionID       string     `json:"transaction_id,omitempty" example:"TXN123456789"`         // Transaction ID
	Gateway             string     `json:"gateway" example:"epay"`                                  // Payment gateway
	PaymentMethod       string     `json:"payment_method" example:"alipay"`                         // Payment method
	Amount              float64    `json:"amount" example:"29.99"`                                  // Payment amount
	Currency            string     `json:"currency" example:"CNY"`                                  // Currency
	ExchangeRate        float64    `json:"exchange_rate" example:"1.0000"`                          // Exchange rate
	Status              string     `json:"status" example:"completed"`                              // Payment status
	PaymentStatus       string     `json:"payment_status,omitempty" example:"success"`              // Gateway payment status
	PaymentURL          string     `json:"payment_url,omitempty" example:"https://example.com/pay"` // Payment URL
	QRCodeURL           string     `json:"qr_code_url,omitempty" example:"https://example.com/qr"`  // QR code URL
	ExpiredAt           *time.Time `json:"expired_at,omitempty" example:"2024-01-01T01:00:00Z"`     // Expiration time
	PaidAt              *time.Time `json:"paid_at,omitempty" example:"2024-01-01T00:30:00Z"`        // Payment completion time
	NotifiedAt          *time.Time `json:"notified_at,omitempty" example:"2024-01-01T00:31:00Z"`    // Notification time
	RefundAmount        float64    `json:"refund_amount" example:"0"`                               // Refund amount
	RefundStatus        string     `json:"refund_status,omitempty" example:"none"`                  // Refund status
	RefundedAt          *time.Time `json:"refunded_at,omitempty" example:"2024-01-02T10:00:00Z"`    // Refund time
	RefundReason        string     `json:"refund_reason,omitempty" example:"User request"`          // Refund reason
	Remark              string     `json:"remark,omitempty" example:"Subscription payment"`         // Remark
	CreatedAt           time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`               // Creation time
	UpdatedAt           time.Time  `json:"updated_at" example:"2024-01-01T00:30:00Z"`               // Update time

	// Related data (to be populated at application layer)
	User              *dto.UserBasicDTO              `json:"user,omitempty"`               // User info
	SubscriptionOrder *dto.SubscriptionOrderBasicDTO `json:"subscription_order,omitempty"` // Subscription order info

	// Computed fields
	IsExpired        bool    `json:"is_expired"`        // Expiration status
	CanRefund        bool    `json:"can_refund"`        // Refundable status
	RefundableAmount float64 `json:"refundable_amount"` // Refundable amount
}

// ToResponse converts PaymentRecord to PaymentRecordResponse
func (pr *PaymentRecord) ToResponse() *PaymentRecordResponse {
	resp := &PaymentRecordResponse{
		ID:                  pr.ID,
		UserID:              pr.UserID,
		SubscriptionOrderID: pr.SubscriptionOrderID,
		PaymentNo:           pr.PaymentNo,
		OutTradeNo:          pr.OutTradeNo,
		TransactionID:       pr.TransactionID,
		Gateway:             pr.Gateway,
		PaymentMethod:       pr.PaymentMethod,
		Amount:              pr.Amount,
		Currency:            pr.Currency,
		ExchangeRate:        pr.ExchangeRate,
		Status:              pr.Status,
		PaymentStatus:       pr.PaymentStatus,
		PaymentURL:          pr.PaymentURL,
		QRCodeURL:           pr.QRCodeURL,
		ExpiredAt:           pr.ExpiredAt,
		PaidAt:              pr.PaidAt,
		NotifiedAt:          pr.NotifiedAt,
		RefundAmount:        pr.RefundAmount,
		RefundStatus:        pr.RefundStatus,
		RefundedAt:          pr.RefundedAt,
		RefundReason:        pr.RefundReason,
		Remark:              pr.Remark,
		CreatedAt:           pr.CreatedAt,
		UpdatedAt:           pr.UpdatedAt,

		// Computed fields
		IsExpired:        pr.IsExpired(),
		CanRefund:        pr.CanBeRefunded(),
		RefundableAmount: pr.GetRefundableAmount(),
	}

	// Note: Related data should be populated at the application layer
	// to avoid cross-domain dependencies

	return resp
}

// ToUserResponse converts PaymentRecord to a response suitable for the paying user
func (pr *PaymentRecord) ToUserResponse() *PaymentRecordResponse {
	resp := pr.ToResponse()

	// Remove sensitive information for user
	resp.User = nil
	resp.NotifiedAt = nil

	return resp
}
