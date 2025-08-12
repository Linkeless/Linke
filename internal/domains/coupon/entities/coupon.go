package entities

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/coupon/constants"
)

// Coupon represents a discount coupon
type Coupon struct {
	// Primary Key
	ID uint64 `json:"id" gorm:"primaryKey"`

	// Core Fields
	Code        string  `json:"code" gorm:"uniqueIndex;size:50;not null"` // 优惠券代码
	Name        string  `json:"name" gorm:"size:100;not null;index"`      // 优惠券名称
	Description string  `json:"description" gorm:"type:text"`             // 描述
	Type        string  `json:"type" gorm:"size:20;not null;index"`       // percentage, fixed_amount
	Value       float64 `json:"value" gorm:"type:decimal(10,2);not null"` // 折扣值

	// Usage Limits
	MaxUses        int `json:"max_uses" gorm:"not null;default:1"`          // 最大使用次数 (0 = 无限制)
	UsedCount      int `json:"used_count" gorm:"not null;default:0"`        // 已使用次数
	MaxUsesPerUser int `json:"max_uses_per_user" gorm:"not null;default:1"` // 每用户最大使用次数

	// Minimum Order Requirements
	MinOrderAmount float64 `json:"min_order_amount" gorm:"type:decimal(10,2);default:0"` // 最小订单金额
	Currency       string  `json:"currency" gorm:"size:3;not null;default:'CNY'"`        // 货币

	// Validity Period
	ValidFrom  *time.Time `json:"valid_from,omitempty" gorm:"index"`  // 生效时间
	ValidUntil *time.Time `json:"valid_until,omitempty" gorm:"index"` // 失效时间

	// Applicable Plans (JSON array of plan IDs, empty means all plans)
	ApplicablePlans string `json:"applicable_plans,omitempty" gorm:"type:text"` // JSON格式的适用套餐ID列表

	// Status & Visibility
	Status    string `json:"status" gorm:"size:20;not null;default:'active';index"` // active, inactive, expired
	IsPublic  bool   `json:"is_public" gorm:"not null;default:false"`               // 是否允许在用户界面显示（仍需通过安全渠道获取）
	CreatedBy uint64 `json:"created_by" gorm:"not null;index"`                      // 创建者ID

	// Metadata
	Metadata string `json:"metadata,omitempty" gorm:"type:text"` // 额外元数据(JSON)

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for Coupon model
func (Coupon) TableName() string {
	return "coupons"
}


// IsValid checks if the coupon is valid for use
func (c *Coupon) IsValid() bool {
	if c.Status != constants.CouponStatusActive {
		return false
	}

	now := time.Now()

	// Check validity period
	if c.ValidFrom != nil && now.Before(*c.ValidFrom) {
		return false
	}

	if c.ValidUntil != nil && now.After(*c.ValidUntil) {
		return false
	}

	// Check usage limits
	if c.MaxUses > 0 && c.UsedCount >= c.MaxUses {
		return false
	}

	return !c.IsDeleted()
}

// IsDeleted checks if the coupon is soft deleted
func (c *Coupon) IsDeleted() bool {
	return c.DeletedAt.Valid
}

// CanBeUsedBy checks if the coupon can be used by a specific user for a specific order amount
func (c *Coupon) CanBeUsedBy(userID uint64, orderAmount float64, planID uint64, db *gorm.DB) (bool, string) {
	if !c.IsValid() {
		return false, "Coupon is not valid or has expired"
	}

	// Check minimum order amount
	if orderAmount < c.MinOrderAmount {
		return false, fmt.Sprintf("Minimum order amount is %.2f %s", c.MinOrderAmount, c.Currency)
	}

	// Check per-user usage limit
	if c.MaxUsesPerUser > 0 {
		var userUsageCount int64
		if err := db.Model(&CouponUsage{}).
			Where("coupon_id = ? AND user_id = ?", c.ID, userID).
			Count(&userUsageCount).Error; err == nil {
			if int(userUsageCount) >= c.MaxUsesPerUser {
				return false, "You have already used this coupon the maximum number of times"
			}
		}
	}

	// Check applicable plans (if specified)
	if c.ApplicablePlans != "" {
		// TODO: Implement plan ID checking logic
		// For now, assume all plans are applicable if ApplicablePlans is not empty
	}

	return true, ""
}

// CalculateDiscount calculates the discount amount for a given order
func (c *Coupon) CalculateDiscount(orderAmount float64) float64 {
	if !c.IsValid() {
		return 0
	}

	var discount float64

	switch c.Type {
	case constants.CouponTypePercentage:
		discount = orderAmount * (c.Value / 100)
	case constants.CouponTypeFixedAmount:
		discount = c.Value
	default:
		return 0
	}

	// Discount cannot exceed order amount
	if discount > orderAmount {
		discount = orderAmount
	}

	return discount
}

// CouponUsage represents a record of coupon usage
type CouponUsage struct {
	// Primary Key
	ID uint64 `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	CouponID            uint64 `json:"coupon_id" gorm:"not null;index"`
	UserID              uint64 `json:"user_id" gorm:"not null;index"`
	SubscriptionOrderID uint64 `json:"subscription_order_id" gorm:"not null;index"`

	// Usage Details
	DiscountAmount float64 `json:"discount_amount" gorm:"type:decimal(10,2);not null"` // 实际折扣金额
	OrderAmount    float64 `json:"order_amount" gorm:"type:decimal(10,2);not null"`    // 订单原金额
	Currency       string  `json:"currency" gorm:"size:3;not null"`                    // 货币

	// Note: Relationships removed to avoid cross-domain dependencies
	// Related data should be fetched and assembled at the application layer

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for CouponUsage model
func (CouponUsage) TableName() string {
	return "coupon_usages"
}

