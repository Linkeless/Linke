package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/payment/constants"
)

// PaymentRecord represents a payment transaction record
type PaymentRecord struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	UserID              uint  `json:"user_id" gorm:"not null;index"`
	SubscriptionOrderID *uint `json:"subscription_order_id,omitempty" gorm:"index"` // 关联的订阅订单
	InvoiceID           *uint `json:"invoice_id,omitempty" gorm:"index"`            // 关联的发票

	// Payment Information
	PaymentNo     string `json:"payment_no" gorm:"uniqueIndex;size:100;not null"`   // 支付单号（系统生成）
	OutTradeNo    string `json:"out_trade_no" gorm:"uniqueIndex;size:100;not null"` // 商户订单号
	TransactionID string `json:"transaction_id" gorm:"size:100;index"`              // 第三方交易号
	Gateway       string `json:"gateway" gorm:"size:50;not null;index"`             // 支付网关 (epay)
	PaymentMethod string `json:"payment_method" gorm:"size:50;not null;index"`      // 支付方式 (alipay, wechat, qqpay)

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

	// Security Enhancement
	LastNotifyTime *time.Time `json:"last_notify_time,omitempty" gorm:"index"` // 最后一次通知时间，用于时间窗口验证
	NotifySource   string     `json:"notify_source,omitempty" gorm:"size:45"`  // 通知来源IP，用于异常检测

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


// IsPending checks if the payment is pending
func (pr *PaymentRecord) IsPending() bool {
	return pr.Status == constants.PaymentRecordStatusPending
}

// IsProcessing checks if the payment is processing
func (pr *PaymentRecord) IsProcessing() bool {
	return pr.Status == constants.PaymentRecordStatusProcessing
}

// IsCompleted checks if the payment is completed
func (pr *PaymentRecord) IsCompleted() bool {
	return pr.Status == constants.PaymentRecordStatusCompleted
}

// IsFailed checks if the payment has failed
func (pr *PaymentRecord) IsFailed() bool {
	return pr.Status == constants.PaymentRecordStatusFailed
}

// IsCancelled checks if the payment is cancelled
func (pr *PaymentRecord) IsCancelled() bool {
	return pr.Status == constants.PaymentRecordStatusCancelled
}

// IsRefunded checks if the payment is refunded
func (pr *PaymentRecord) IsRefunded() bool {
	return pr.Status == constants.PaymentRecordStatusRefunded
}

// IsExpired checks if the payment has expired
func (pr *PaymentRecord) IsExpired() bool {
	return pr.ExpiredAt != nil && pr.ExpiredAt.Before(time.Now())
}

// CanBeRefunded checks if the payment can be refunded
func (pr *PaymentRecord) CanBeRefunded() bool {
	return pr.IsCompleted() && pr.RefundStatus != constants.RefundStatusCompleted
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

