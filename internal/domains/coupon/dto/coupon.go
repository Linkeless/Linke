package dto

import (
	"sync"
	"time"

	"linke/internal/domains/coupon/entities"
	"linke/internal/shared/dto"
)

// Request DTOs

// CreateCouponRequest represents the request to create a coupon
type CreateCouponRequest struct {
	Code            string     `json:"code" binding:"required,min=3,max=50" example:"SAVE20"`
	Name            string     `json:"name" binding:"required,min=1,max=100" example:"20% Off All Plans"`
	Description     string     `json:"description,omitempty" binding:"max=1000" example:"Save 20% on any subscription plan"`
	Type            string     `json:"type" binding:"required,oneof=percentage fixed_amount" example:"percentage"`
	Value           float64    `json:"value" binding:"required,min=0" example:"20"`
	MaxUses         int        `json:"max_uses,omitempty" binding:"min=0" example:"100"`
	MaxUsesPerUser  int        `json:"max_uses_per_user,omitempty" binding:"min=1" example:"1"`
	MinOrderAmount  float64    `json:"min_order_amount,omitempty" binding:"min=0" example:"10"`
	Currency        string     `json:"currency,omitempty" binding:"omitempty,len=3" example:"CNY"`
	ValidFrom       *time.Time `json:"valid_from,omitempty" example:"2024-01-01T00:00:00Z"`
	ValidUntil      *time.Time `json:"valid_until,omitempty" example:"2024-12-31T23:59:59Z"`
	ApplicablePlans string     `json:"applicable_plans,omitempty" example:"[1,2,3]"`
	IsPublic        *bool      `json:"is_public,omitempty" example:"true"`
}

// UpdateCouponRequest represents the request to update a coupon
type UpdateCouponRequest struct {
	Name            *string    `json:"name,omitempty" binding:"omitempty,min=1,max=100" example:"Updated Coupon Name"`
	Description     *string    `json:"description,omitempty" binding:"omitempty,max=1000" example:"Updated description"`
	Type            *string    `json:"type,omitempty" binding:"omitempty,oneof=percentage fixed_amount" example:"percentage"`
	Value           *float64   `json:"value,omitempty" binding:"omitempty,min=0" example:"25"`
	MaxUses         *int       `json:"max_uses,omitempty" binding:"omitempty,min=0" example:"200"`
	MaxUsesPerUser  *int       `json:"max_uses_per_user,omitempty" binding:"omitempty,min=1" example:"2"`
	MinOrderAmount  *float64   `json:"min_order_amount,omitempty" binding:"omitempty,min=0" example:"15"`
	ValidFrom       *time.Time `json:"valid_from,omitempty" example:"2024-02-01T00:00:00Z"`
	ValidUntil      *time.Time `json:"valid_until,omitempty" example:"2024-11-30T23:59:59Z"`
	ApplicablePlans *string    `json:"applicable_plans,omitempty" example:"[1,2,4]"`
	Status          *string    `json:"status,omitempty" binding:"omitempty,oneof=active inactive expired" example:"active"`
	IsPublic        *bool      `json:"is_public,omitempty" example:"false"`
}

// GetCouponsRequest represents the request to get coupons
type GetCouponsRequest struct {
	Status   string `form:"status,omitempty" binding:"omitempty,oneof=active inactive expired" example:"active"`
	Type     string `form:"type,omitempty" binding:"omitempty,oneof=percentage fixed_amount" example:"percentage"`
	IsPublic *bool  `form:"is_public,omitempty" example:"true"`
	Limit    int    `form:"limit,omitempty" binding:"omitempty,min=1,max=100" example:"10"`
	Offset   int    `form:"offset,omitempty" binding:"omitempty,min=0" example:"0"`
}

// ValidateCouponRequest represents the request to validate a coupon
type ValidateCouponRequest struct {
	Code        string  `json:"code" binding:"required" example:"SAVE20"`
	UserID      uint64  `json:"user_id" binding:"required" example:"1"`
	OrderAmount float64 `json:"order_amount" binding:"required,min=0" example:"29.99"`
	PlanID      uint64  `json:"plan_id" binding:"required" example:"1"`
	Currency    string  `json:"currency" binding:"required,len=3" example:"CNY"`
}

// ToggleStatusRequest represents the request for toggling coupon status
type ToggleStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive" example:"active"`
}

// ExtendExpiryRequest represents the request for extending coupon expiry
type ExtendExpiryRequest struct {
	ExtendDays int        `json:"extend_days,omitempty" binding:"omitempty,min=1" example:"30"`
	NewExpiry  *time.Time `json:"new_expiry,omitempty" example:"2024-12-31T23:59:59Z"`
	ExtendType string     `json:"extend_type" binding:"required,oneof=days specific" example:"days"`
}

// BulkCreateCouponRequest represents the request for bulk coupon creation
type BulkCreateCouponRequest struct {
	CodePrefix      string     `json:"code_prefix" binding:"required,min=2,max=20" example:"BULK"`
	Count           int        `json:"count" binding:"required,min=1,max=1000" example:"100"`
	Name            string     `json:"name" binding:"required,min=1,max=100" example:"Bulk Generated Coupons"`
	Description     string     `json:"description,omitempty" binding:"max=1000" example:"Bulk generated discount coupons"`
	Type            string     `json:"type" binding:"required,oneof=percentage fixed_amount" example:"percentage"`
	Value           float64    `json:"value" binding:"required,min=0" example:"20"`
	MaxUses         int        `json:"max_uses,omitempty" binding:"min=0" example:"1"`
	MaxUsesPerUser  int        `json:"max_uses_per_user,omitempty" binding:"min=1" example:"1"`
	MinOrderAmount  float64    `json:"min_order_amount,omitempty" binding:"min=0" example:"10"`
	Currency        string     `json:"currency,omitempty" binding:"omitempty,len=3" example:"CNY"`
	ValidFrom       *time.Time `json:"valid_from,omitempty" example:"2024-01-01T00:00:00Z"`
	ValidUntil      *time.Time `json:"valid_until,omitempty" example:"2024-12-31T23:59:59Z"`
	ApplicablePlans string     `json:"applicable_plans,omitempty" example:"[1,2,3]"`
	IsPublic        *bool      `json:"is_public,omitempty" example:"true"`
}

// BulkUpdateRequest represents the request for bulk operations
type BulkUpdateRequest struct {
	IDs    []uint64 `json:"ids" binding:"required,min=1,max=100"`
	Status *string  `json:"status,omitempty" binding:"omitempty,oneof=active inactive expired"`
}

// SearchCouponsRequest represents the search request
type SearchCouponsRequest struct {
	Query         string     `form:"q" binding:"omitempty,min=1,max=100"`
	Status        string     `form:"status,omitempty" binding:"omitempty,oneof=active inactive expired"`
	Type          string     `form:"type,omitempty" binding:"omitempty,oneof=percentage fixed_amount"`
	IsPublic      *bool      `form:"is_public,omitempty"`
	CreatedAfter  *time.Time `form:"created_after,omitempty"`
	CreatedBefore *time.Time `form:"created_before,omitempty"`
	ExpiresAfter  *time.Time `form:"expires_after,omitempty"`
	ExpiresBefore *time.Time `form:"expires_before,omitempty"`
	MinValue      *float64   `form:"min_value,omitempty" binding:"omitempty,min=0"`
	MaxValue      *float64   `form:"max_value,omitempty" binding:"omitempty,min=0"`
	MinUsed       *int       `form:"min_used,omitempty" binding:"omitempty,min=0"`
	MaxUsed       *int       `form:"max_used,omitempty" binding:"omitempty,min=0"`
	Page          int        `form:"page,omitempty" binding:"omitempty,min=1" example:"1"`
	Limit         int        `form:"limit,omitempty" binding:"omitempty,min=1,max=100" example:"10"`
}

// Response DTOs

// ValidateCouponResponse represents the response of coupon validation
type ValidateCouponResponse struct {
	Valid          bool            `json:"valid" example:"true"`
	Message        string          `json:"message" example:"Coupon is valid"`
	DiscountAmount float64         `json:"discount_amount" example:"5.99"`
	FinalAmount    float64         `json:"final_amount" example:"24.00"`
	Coupon         *CouponResponse `json:"coupon,omitempty"`
}

// CouponResponse represents the coupon data structure for API responses
type CouponResponse struct {
	ID              uint64     `json:"id" example:"1"`                                       // Coupon ID
	Code            string     `json:"code" example:"SAVE20"`                                // Coupon code
	Name            string     `json:"name" example:"20% Off"`                               // Coupon name
	Description     string     `json:"description" example:"Save 20% on any plan"`           // Description
	Type            string     `json:"type" example:"percentage"`                            // Discount type
	Value           float64    `json:"value" example:"20"`                                   // Discount value
	MaxUses         int        `json:"max_uses" example:"100"`                               // Maximum uses
	UsedCount       int        `json:"used_count" example:"15"`                              // Used count
	MaxUsesPerUser  int        `json:"max_uses_per_user" example:"1"`                        // Max uses per user
	MinOrderAmount  float64    `json:"min_order_amount" example:"10"`                        // Minimum order amount
	Currency        string     `json:"currency" example:"CNY"`                               // Currency
	ValidFrom       *time.Time `json:"valid_from,omitempty" example:"2024-01-01T00:00:00Z"`  // Valid from
	ValidUntil      *time.Time `json:"valid_until,omitempty" example:"2024-12-31T23:59:59Z"` // Valid until
	ApplicablePlans string     `json:"applicable_plans,omitempty"`                           // Applicable plans
	Status          string     `json:"status" example:"active"`                              // Status
	IsPublic        bool       `json:"is_public" example:"true"`                             // Public visibility
	CreatedAt       time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`            // Creation time
	UpdatedAt       time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`            // Update time
}

// CouponUsageResponse represents the coupon usage data structure for API responses
type CouponUsageResponse struct {
	ID                  uint64    `json:"id" example:"1"`                            // Usage ID
	CouponID            uint64    `json:"coupon_id" example:"1"`                     // Coupon ID
	UserID              uint64    `json:"user_id" example:"1"`                       // User ID
	SubscriptionOrderID uint64    `json:"subscription_order_id" example:"1"`         // Order ID
	DiscountAmount      float64   `json:"discount_amount" example:"5.99"`            // Discount amount
	OrderAmount         float64   `json:"order_amount" example:"29.99"`              // Original order amount
	Currency            string    `json:"currency" example:"CNY"`                    // Currency
	CreatedAt           time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"` // Creation time
	UpdatedAt           time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"` // Update time

	// Related data (to be populated at application layer)
	Coupon            *CouponResponse                `json:"coupon,omitempty"`             // Coupon info
	User              *dto.UserBasicDTO              `json:"user,omitempty"`               // User info
	SubscriptionOrder *dto.SubscriptionOrderBasicDTO `json:"subscription_order,omitempty"` // Order info
}

// Helper methods for conversion

// ToResponse converts Coupon entity to CouponResponse DTO
func ToResponse(coupon *entities.Coupon) *CouponResponse {
	return &CouponResponse{
		ID:              coupon.ID,
		Code:            coupon.Code,
		Name:            coupon.Name,
		Description:     coupon.Description,
		Type:            coupon.Type,
		Value:           coupon.Value,
		MaxUses:         int(coupon.MaxUses),        // Convert from int32
		UsedCount:       int(coupon.UsedCount),      // Convert from int32
		MaxUsesPerUser:  int(coupon.MaxUsesPerUser), // Convert from int32
		MinOrderAmount:  coupon.MinOrderAmount,
		Currency:        coupon.Currency,
		ValidFrom:       coupon.ValidFrom,
		ValidUntil:      coupon.ValidUntil,
		ApplicablePlans: coupon.ApplicablePlans,
		Status:          coupon.Status,
		IsPublic:        coupon.IsPublic,
		CreatedAt:       coupon.CreatedAt,
		UpdatedAt:       coupon.UpdatedAt,
	}
}

// ToPublicResponse converts Coupon entity to public CouponResponse DTO (limited information)
func ToPublicResponse(coupon *entities.Coupon) *CouponResponse {
	return &CouponResponse{
		ID:             coupon.ID,
		Code:           coupon.Code,
		Name:           coupon.Name,
		Description:    coupon.Description,
		Type:           coupon.Type,
		Value:          coupon.Value,
		MinOrderAmount: coupon.MinOrderAmount,
		Currency:       coupon.Currency,
		ValidFrom:      coupon.ValidFrom,
		ValidUntil:     coupon.ValidUntil,
		Status:         coupon.Status,
	}
}

// CouponUsageToResponse converts CouponUsage entity to CouponUsageResponse DTO
func CouponUsageToResponse(usage *entities.CouponUsage) *CouponUsageResponse {
	return &CouponUsageResponse{
		ID:                  usage.ID,
		CouponID:            usage.CouponID,
		UserID:              usage.UserID,
		SubscriptionOrderID: usage.SubscriptionOrderID,
		DiscountAmount:      usage.DiscountAmount,
		OrderAmount:         usage.OrderAmount,
		Currency:            usage.Currency,
		CreatedAt:           usage.CreatedAt,
		UpdatedAt:           usage.UpdatedAt,
		// Note: Related data should be populated at the application layer
		// to avoid cross-domain dependencies
	}
}

// Object pools for frequently used DTOs to reduce GC pressure
var (
	// CouponResponsePool reduces allocations for API responses
	couponResponsePool = sync.Pool{
		New: func() any {
			return &CouponResponse{}
		},
	}

	// CreateCouponRequestPool reduces allocations for creation requests
	createCouponRequestPool = sync.Pool{
		New: func() any {
			return &CreateCouponRequest{}
		},
	}

	// ValidateCouponResponsePool reduces allocations for validation responses
	validateCouponResponsePool = sync.Pool{
		New: func() any {
			return &ValidateCouponResponse{}
		},
	}

	// CouponUsageResponsePool reduces allocations for usage response DTOs
	couponUsageResponsePool = sync.Pool{
		New: func() any {
			return &CouponUsageResponse{}
		},
	}
)

// Object pool management functions

// GetCouponResponse retrieves a CouponResponse from the pool
func GetCouponResponse() *CouponResponse {
	return couponResponsePool.Get().(*CouponResponse)
}

// PutCouponResponse returns a CouponResponse to the pool after resetting it
func PutCouponResponse(resp *CouponResponse) {
	// Reset the response to prevent data leakage
	*resp = CouponResponse{}
	couponResponsePool.Put(resp)
}

// GetCreateCouponRequest retrieves a CreateCouponRequest from the pool
func GetCreateCouponRequest() *CreateCouponRequest {
	return createCouponRequestPool.Get().(*CreateCouponRequest)
}

// PutCreateCouponRequest returns a CreateCouponRequest to the pool after resetting it
func PutCreateCouponRequest(req *CreateCouponRequest) {
	// Reset the request to prevent data leakage
	*req = CreateCouponRequest{}
	createCouponRequestPool.Put(req)
}

// GetValidateCouponResponse retrieves a ValidateCouponResponse from the pool
func GetValidateCouponResponse() *ValidateCouponResponse {
	return validateCouponResponsePool.Get().(*ValidateCouponResponse)
}

// PutValidateCouponResponse returns a ValidateCouponResponse to the pool after resetting it
func PutValidateCouponResponse(resp *ValidateCouponResponse) {
	// Reset the response to prevent data leakage
	*resp = ValidateCouponResponse{}
	validateCouponResponsePool.Put(resp)
}

// GetCouponUsageResponse retrieves a CouponUsageResponse from the pool
func GetCouponUsageResponse() *CouponUsageResponse {
	return couponUsageResponsePool.Get().(*CouponUsageResponse)
}

// PutCouponUsageResponse returns a CouponUsageResponse to the pool after resetting it
func PutCouponUsageResponse(resp *CouponUsageResponse) {
	// Reset the response to prevent data leakage
	*resp = CouponUsageResponse{}
	couponUsageResponsePool.Put(resp)
}

// ToResponsePooled converts Coupon entity to CouponResponse DTO using object pool
func ToResponsePooled(coupon *entities.Coupon) *CouponResponse {
	resp := GetCouponResponse()
	resp.ID = coupon.ID
	resp.Code = coupon.Code
	resp.Name = coupon.Name
	resp.Description = coupon.Description
	resp.Type = coupon.Type
	resp.Value = coupon.Value
	resp.MaxUses = int(coupon.MaxUses)               // Convert from int32
	resp.UsedCount = int(coupon.UsedCount)           // Convert from int32
	resp.MaxUsesPerUser = int(coupon.MaxUsesPerUser) // Convert from int32
	resp.MinOrderAmount = coupon.MinOrderAmount
	resp.Currency = coupon.Currency
	resp.ValidFrom = coupon.ValidFrom
	resp.ValidUntil = coupon.ValidUntil
	resp.ApplicablePlans = coupon.ApplicablePlans
	resp.Status = coupon.Status
	resp.IsPublic = coupon.IsPublic
	resp.CreatedAt = coupon.CreatedAt
	resp.UpdatedAt = coupon.UpdatedAt
	return resp
}

// CouponUsageToResponsePooled converts CouponUsage entity to CouponUsageResponse DTO using object pool
func CouponUsageToResponsePooled(usage *entities.CouponUsage) *CouponUsageResponse {
	resp := GetCouponUsageResponse()
	resp.ID = usage.ID
	resp.CouponID = usage.CouponID
	resp.UserID = usage.UserID
	resp.SubscriptionOrderID = usage.SubscriptionOrderID
	resp.DiscountAmount = usage.DiscountAmount
	resp.OrderAmount = usage.OrderAmount
	resp.Currency = usage.Currency
	resp.CreatedAt = usage.CreatedAt
	resp.UpdatedAt = usage.UpdatedAt
	// Note: Related data should be populated at the application layer
	return resp
}
