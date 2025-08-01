package persistence

import (
	"time"
	
	"gorm.io/gorm"
)

// CouponPO represents the coupon persistent object for GORM
type CouponPO struct {
	// Primary Key
	ID uint64 `gorm:"primaryKey;column:id"`

	// Core Fields
	Code        string  `gorm:"uniqueIndex;size:50;not null;column:code"`
	Name        string  `gorm:"size:100;not null;index;column:name"`
	Description string  `gorm:"type:text;column:description"`
	Type        string  `gorm:"size:20;not null;index;column:type"`
	Value       float64 `gorm:"type:decimal(10,2);not null;column:value"`

	// Usage Limits
	MaxUses        int `gorm:"not null;default:1;column:max_uses"`
	UsedCount      int `gorm:"not null;default:0;column:used_count"`
	MaxUsesPerUser int `gorm:"not null;default:1;column:max_uses_per_user"`

	// Minimum Order Requirements
	MinOrderAmount float64 `gorm:"type:decimal(10,2);default:0;column:min_order_amount"`
	Currency       string  `gorm:"size:3;not null;default:'USD';column:currency"`

	// Validity Period
	ValidFrom  *time.Time `gorm:"index;column:valid_from"`
	ValidUntil *time.Time `gorm:"index;column:valid_until"`

	// Applicable Plans (JSON array of plan IDs)
	ApplicablePlans string `gorm:"type:text;column:applicable_plans"`

	// Status & Visibility
	Status    string `gorm:"size:20;not null;default:'active';index;column:status"`
	IsPublic  bool   `gorm:"not null;default:false;column:is_public"`
	CreatedBy uint64 `gorm:"not null;index;column:created_by"`

	// Metadata
	Metadata string `gorm:"type:text;column:metadata"`

	// Timestamp Fields
	CreatedAt time.Time      `gorm:"not null;index;column:created_at"`
	UpdatedAt time.Time      `gorm:"not null;column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at"`

	// Relationships
	Usages []CouponUsagePO `gorm:"foreignKey:CouponID;references:ID"`
}

// TableName returns the table name for CouponPO
func (CouponPO) TableName() string {
	return "coupons"
}

// CouponUsagePO represents the coupon usage persistent object for GORM
type CouponUsagePO struct {
	// Primary Key
	ID uint64 `gorm:"primaryKey;column:id"`

	// Foreign Keys
	CouponID uint64 `gorm:"not null;index;column:coupon_id"`
	UserID   uint64 `gorm:"not null;index;column:user_id"`
	OrderID  uint64 `gorm:"not null;index;column:order_id"`

	// Usage Details
	DiscountAmount float64 `gorm:"type:decimal(10,2);not null;column:discount_amount"`
	OrderAmount    float64 `gorm:"type:decimal(10,2);not null;column:order_amount"`
	Currency       string  `gorm:"size:3;not null;column:currency"`

	// Timestamp Fields
	CreatedAt time.Time      `gorm:"not null;index;column:created_at"`
	UpdatedAt time.Time      `gorm:"not null;column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at"`

	// Relationships
	Coupon *CouponPO `gorm:"foreignKey:CouponID;references:ID"`
}

// TableName returns the table name for CouponUsagePO
func (CouponUsagePO) TableName() string {
	return "coupon_usages"
}