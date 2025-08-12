package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/payment/constants"
)

// CryptoWalletConfig represents cryptocurrency wallet configuration for receiving payments
type CryptoWalletConfig struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Network and Currency Information
	Network  string `json:"network" gorm:"size:50;not null;index"`  // trc, polygon, etc.
	Currency string `json:"currency" gorm:"size:20;not null;index"` // USDT, BTC, ETH, etc.
	Symbol   string `json:"symbol" gorm:"size:10;not null"`         // USDT, BTC, ETH for display

	// Wallet Configuration
	WalletAddress    string `json:"wallet_address" gorm:"size:255;not null;unique"` // 收款钱包地址
	WalletName       string `json:"wallet_name" gorm:"size:100"`                    // 钱包显示名称
	ContractAddress  string `json:"contract_address,omitempty" gorm:"size:255"`     // 合约地址（如USDT代币合约）
	Decimals         int    `json:"decimals" gorm:"default:18"`                     // 代币精度
	MinConfirmations int    `json:"min_confirmations" gorm:"default:1"`             // 最小确认数

	// Display and Settings
	DisplayName string  `json:"display_name" gorm:"size:100;not null"` // 显示名称，如 "TRC-USDT"
	Description string  `json:"description" gorm:"type:text"`          // 描述信息
	Icon        string  `json:"icon,omitempty" gorm:"size:255"`        // 图标URL
	IsEnabled   bool    `json:"is_enabled" gorm:"default:true;index"`  // 是否启用
	SortOrder   int     `json:"sort_order" gorm:"default:0;index"`     // 排序顺序

	// Transaction Limits
	MinAmount float64 `json:"min_amount" gorm:"type:decimal(20,8);default:0.01"`      // 最小交易金额
	MaxAmount float64 `json:"max_amount" gorm:"type:decimal(20,8);default:100000.00"` // 最大交易金额

	// Fee Configuration
	NetworkFee    float64 `json:"network_fee" gorm:"type:decimal(20,8);default:0"`    // 网络手续费
	ProcessingFee float64 `json:"processing_fee" gorm:"type:decimal(5,4);default:0"`  // 处理费率(%)
	FixedFee      float64 `json:"fixed_fee" gorm:"type:decimal(10,2);default:0"`      // 固定手续费

	// API Configuration (for blockchain queries)
	APIEndpoint string `json:"api_endpoint,omitempty" gorm:"size:255"` // 区块链API端点
	APIKey      string `json:"api_key,omitempty" gorm:"size:255"`      // API密钥

	// Status and Monitoring
	LastCheckAt   *time.Time `json:"last_check_at,omitempty" gorm:"index"` // 最后检查时间
	LastTxHash    string     `json:"last_tx_hash,omitempty" gorm:"size:255"` // 最后交易哈希
	Balance       float64    `json:"balance" gorm:"type:decimal(20,8);default:0"` // 钱包余额
	Active        bool       `json:"is_active" gorm:"default:true"`         // 钱包是否活跃
	HealthStatus  string     `json:"health_status" gorm:"size:20;default:'healthy'"` // 健康状态: healthy, warning, error

	// Security and Validation
	AddressValidated bool       `json:"address_validated" gorm:"default:false"`     // 地址是否已验证
	ValidatedAt      *time.Time `json:"validated_at,omitempty"`                     // 验证时间
	ValidationHash   string     `json:"validation_hash,omitempty" gorm:"size:64"`   // 验证哈希

	// Metadata
	Metadata string `json:"metadata,omitempty" gorm:"type:json"` // 额外配置信息JSON

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for CryptoWalletConfig model
func (CryptoWalletConfig) TableName() string {
	return "crypto_wallet_configs"
}

// IsActive checks if the crypto wallet config is active and can be used
func (cwc *CryptoWalletConfig) IsActive() bool {
	return cwc.IsEnabled && cwc.Active && !cwc.IsDeleted() && cwc.AddressValidated
}

// IsDeleted checks if the crypto wallet config is soft deleted
func (cwc *CryptoWalletConfig) IsDeleted() bool {
	return cwc.DeletedAt.Valid
}

// CanAcceptPayment checks if this wallet can accept payments
func (cwc *CryptoWalletConfig) CanAcceptPayment(amount float64) bool {
	if !cwc.IsActive() {
		return false
	}
	return amount >= cwc.MinAmount && amount <= cwc.MaxAmount
}

// GetPaymentMethod returns the corresponding payment method constant
func (cwc *CryptoWalletConfig) GetPaymentMethod() string {
	networkCurrency := cwc.Network + "_" + cwc.Currency
	switch networkCurrency {
	case "trc_USDT":
		return constants.PaymentMethodTRCUSDT
	case "polygon_USDT":
		return constants.PaymentMethodPolygonUSDT
	default:
		// 通用格式：network_currency 转为小写
		return networkCurrency
	}
}

// GetGatewayType returns the gateway type for this crypto payment
func (cwc *CryptoWalletConfig) GetGatewayType() string {
	return constants.PaymentGatewayCrypto
}

// IsHealthy checks if the wallet configuration is healthy
func (cwc *CryptoWalletConfig) IsHealthy() bool {
	return cwc.HealthStatus == "healthy"
}

// NeedsValidation checks if the wallet address needs revalidation
func (cwc *CryptoWalletConfig) NeedsValidation() bool {
	if !cwc.AddressValidated {
		return true
	}
	if cwc.ValidatedAt == nil {
		return true
	}
	// Revalidate every 7 days
	return time.Since(*cwc.ValidatedAt) > 7*24*time.Hour
}

// CalculateProcessingFee calculates the processing fee for a given amount
func (cwc *CryptoWalletConfig) CalculateProcessingFee(amount float64) float64 {
	return amount * cwc.ProcessingFee / 100
}

// CalculateTotalFee calculates the total fee (fixed + processing + network) for a given amount
func (cwc *CryptoWalletConfig) CalculateTotalFee(amount float64) float64 {
	return cwc.FixedFee + cwc.CalculateProcessingFee(amount) + cwc.NetworkFee
}

// GetDisplayInfo returns formatted display information for this wallet
func (cwc *CryptoWalletConfig) GetDisplayInfo() map[string]interface{} {
	return map[string]interface{}{
		"display_name": cwc.DisplayName,
		"network":      cwc.Network,
		"currency":     cwc.Currency,
		"symbol":       cwc.Symbol,
		"icon":         cwc.Icon,
		"min_amount":   cwc.MinAmount,
		"max_amount":   cwc.MaxAmount,
		"description":  cwc.Description,
	}
}