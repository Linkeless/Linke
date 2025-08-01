package persistence

import (
	"time"

	"gorm.io/gorm"
)

// PaymentConfigPO represents the payment config persistent object for database storage
type PaymentConfigPO struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Gateway Information
	Gateway string `json:"gateway" gorm:"unique;size:50;not null"`
	Name    string `json:"name" gorm:"size:100;not null"`

	// Configuration
	Config string `json:"config" gorm:"type:json;not null"`

	// Status and Settings
	IsEnabled bool `json:"is_enabled" gorm:"not null;default:true;index"`
	SortOrder int  `json:"sort_order" gorm:"default:0;index"`

	// Currency and Methods
	SupportedCurrencies string `json:"supported_currencies" gorm:"type:text"`
	SupportedMethods    string `json:"supported_methods" gorm:"type:json"`

	// Limits
	MinAmount float64 `json:"min_amount" gorm:"type:decimal(10,2);default:0.01"`
	MaxAmount float64 `json:"max_amount" gorm:"type:decimal(10,2);default:99999.99"`

	// Fee Information
	FixedFee      float64 `json:"fixed_fee" gorm:"type:decimal(10,2);default:0.00"`
	PercentageFee float64 `json:"percentage_fee" gorm:"type:decimal(5,4);default:0.0000"`

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for PaymentConfigPO
func (PaymentConfigPO) TableName() string {
	return "payment_configs"
}