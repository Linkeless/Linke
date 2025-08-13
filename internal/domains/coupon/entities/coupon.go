package entities

import (
	"fmt"
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/coupon/constants"
)

// Coupon represents a discount coupon
// Field ordering optimized for memory alignment to minimize padding
type Coupon struct {
	// 8-byte aligned fields (largest first)
	ID             uint64         `json:"id" gorm:"primaryKey"`                                 // Primary Key
	CreatedBy      uint64         `json:"created_by" gorm:"not null;index"`                     // 创建者ID
	Value          float64        `json:"value" gorm:"type:decimal(10,2);not null"`             // 折扣值
	MinOrderAmount float64        `json:"min_order_amount" gorm:"type:decimal(10,2);default:0"` // 最小订单金额
	ValidFrom      *time.Time     `json:"valid_from,omitempty" gorm:"index"`                    // 生效时间
	ValidUntil     *time.Time     `json:"valid_until,omitempty" gorm:"index"`                   // 失效时间
	CreatedAt      time.Time      `json:"created_at" gorm:"not null;index"`                     // Timestamp Fields
	UpdatedAt      time.Time      `json:"updated_at" gorm:"not null"`                           // Timestamp Fields
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`                    // Timestamp Fields

	// 4-byte aligned fields (int32 for atomic operations optimization)
	MaxUses        int32 `json:"max_uses" gorm:"not null;default:1"`          // 最大使用次数 (0 = 无限制) - atomic safe
	UsedCount      int32 `json:"used_count" gorm:"not null;default:0"`        // 已使用次数 - atomic safe
	MaxUsesPerUser int32 `json:"max_uses_per_user" gorm:"not null;default:1"` // 每用户最大使用次数

	// String fields (16 bytes each on 64-bit systems)
	Code            string `json:"code" gorm:"uniqueIndex;size:50;not null"`              // 优惠券代码
	Name            string `json:"name" gorm:"size:100;not null;index"`                   // 优惠券名称
	Description     string `json:"description" gorm:"type:text"`                          // 描述
	Type            string `json:"type" gorm:"size:20;not null;index"`                    // percentage, fixed_amount
	Currency        string `json:"currency" gorm:"size:3;not null;default:'CNY'"`         // 货币
	ApplicablePlans string `json:"applicable_plans,omitempty" gorm:"type:text"`           // JSON格式的适用套餐ID列表
	Status          string `json:"status" gorm:"size:20;not null;default:'active';index"` // active, inactive, expired
	Metadata        string `json:"metadata,omitempty" gorm:"type:text"`                   // 额外元数据(JSON)

	// 1-byte aligned fields (smallest last)
	IsPublic bool `json:"is_public" gorm:"not null;default:false"` // 是否允许在用户界面显示（仍需通过安全渠道获取）
}

// TableName returns the table name for Coupon model
func (Coupon) TableName() string {
	return "coupons"
}

// AtomicIncrementUsedCount atomically increments the used count
func (c *Coupon) AtomicIncrementUsedCount() int32 {
	return atomic.AddInt32(&c.UsedCount, 1)
}

// AtomicLoadUsedCount atomically loads the used count
func (c *Coupon) AtomicLoadUsedCount() int32 {
	return atomic.LoadInt32(&c.UsedCount)
}

// AtomicStoreUsedCount atomically stores the used count
func (c *Coupon) AtomicStoreUsedCount(count int32) {
	atomic.StoreInt32(&c.UsedCount, count)
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

	// Check usage limits (atomic-safe access)
	if c.MaxUses > 0 && atomic.LoadInt32(&c.UsedCount) >= c.MaxUses {
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
			if int32(userUsageCount) >= c.MaxUsesPerUser {
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
// Field ordering optimized for memory alignment to minimize padding
type CouponUsage struct {
	// 8-byte aligned fields (largest first)
	ID                  uint64         `json:"id" gorm:"primaryKey"`                               // Primary Key
	CouponID            uint64         `json:"coupon_id" gorm:"not null;index"`                    // Foreign Keys
	UserID              uint64         `json:"user_id" gorm:"not null;index"`                      // Foreign Keys
	SubscriptionOrderID uint64         `json:"subscription_order_id" gorm:"not null;index"`        // Foreign Keys
	DiscountAmount      float64        `json:"discount_amount" gorm:"type:decimal(10,2);not null"` // 实际折扣金额
	OrderAmount         float64        `json:"order_amount" gorm:"type:decimal(10,2);not null"`    // 订单原金额
	CreatedAt           time.Time      `json:"created_at" gorm:"not null;index"`                   // Timestamp Fields
	UpdatedAt           time.Time      `json:"updated_at" gorm:"not null"`                         // Timestamp Fields
	DeletedAt           gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`                  // Timestamp Fields

	// String fields (16 bytes each on 64-bit systems)
	Currency string `json:"currency" gorm:"size:3;not null"` // 货币

	// Note: Relationships removed to avoid cross-domain dependencies
	// Related data should be fetched and assembled at the application layer
}

// TableName returns the table name for CouponUsage model
func (CouponUsage) TableName() string {
	return "coupon_usages"
}
