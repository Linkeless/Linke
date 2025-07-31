package shared

import (
	"errors"
	"strconv"
	"strings"

	"linke/internal/model"
	"linke/internal/response"

	"github.com/gin-gonic/gin"
)

// CouponValidator provides validation utilities for coupon operations
type CouponValidator struct{}

// NewCouponValidator creates a new coupon validator
func NewCouponValidator() *CouponValidator {
	return &CouponValidator{}
}

// ValidateCouponID validates and parses coupon ID from URL parameter
func (v *CouponValidator) ValidateCouponID(c *gin.Context) (uint64, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid coupon ID")
		return 0, errors.New("invalid coupon ID")
	}
	return id, nil
}

// ValidateCouponType validates coupon type value
func (v *CouponValidator) ValidateCouponType(couponType string) error {
	if couponType != model.CouponTypePercentage && couponType != model.CouponTypeFixedAmount {
		return errors.New("invalid type value, must be 'percentage' or 'fixed_amount'")
	}
	return nil
}

// ValidateCouponStatus validates coupon status value
func (v *CouponValidator) ValidateCouponStatus(status string) error {
	if status != model.CouponStatusActive && status != model.CouponStatusInactive && status != model.CouponStatusExpired {
		return errors.New("invalid status value, must be 'active', 'inactive', or 'expired'")
	}
	return nil
}

// ValidateCouponCode validates coupon code format
func (v *CouponValidator) ValidateCouponCode(code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return errors.New("coupon code is required")
	}
	
	if len(code) < 3 {
		return errors.New("coupon code must be at least 3 characters long")
	}
	
	if len(code) > 50 {
		return errors.New("coupon code must be less than 50 characters")
	}
	
	return nil
}

// ValidateCouponValue validates coupon value based on type
func (v *CouponValidator) ValidateCouponValue(couponType string, value float64) error {
	if value < 0 {
		return errors.New("coupon value must be non-negative")
	}
	
	if couponType == model.CouponTypePercentage && value > 100 {
		return errors.New("percentage discount cannot exceed 100%")
	}
	
	return nil
}

// ValidatePaginationParams validates and returns pagination parameters
func (v *CouponValidator) ValidatePaginationParams(c *gin.Context) (int, int) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return limit, offset
}

// ValidateFilterParams extracts and validates filter parameters
func (v *CouponValidator) ValidateFilterParams(c *gin.Context) (map[string]interface{}, error) {
	filters := make(map[string]interface{})
	
	if status := c.Query("status"); status != "" {
		if err := v.ValidateCouponStatus(status); err != nil {
			return nil, err
		}
		filters["status"] = status
	}
	
	if couponType := c.Query("type"); couponType != "" {
		if err := v.ValidateCouponType(couponType); err != nil {
			return nil, err
		}
		filters["type"] = couponType
	}
	
	if isPublicStr := c.Query("is_public"); isPublicStr != "" {
		if isPublic, err := strconv.ParseBool(isPublicStr); err == nil {
			filters["is_public"] = isPublic
		}
	}
	
	return filters, nil
}

// ValidateCreateCouponRequest validates coupon creation data
func (v *CouponValidator) ValidateCreateCouponRequest(c *gin.Context, req interface{}) error {
	// The request binding is already handled by gin, so we just need to validate business rules
	// This can be extended for additional validation logic if needed
	return nil
}

// ValidateUpdateCouponRequest validates coupon update data
func (v *CouponValidator) ValidateUpdateCouponRequest(c *gin.Context, req interface{}) error {
	// The request binding is already handled by gin, so we just need to validate business rules
	// This can be extended for additional validation logic if needed
	return nil
}

// ValidateUserID gets and validates user ID from context
func (v *CouponValidator) ValidateUserID(c *gin.Context) (uint64, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return 0, errors.New("user not authenticated")
	}
	
	if id, ok := userID.(uint); ok {
		return uint64(id), nil
	}
	
	response.InternalServerError(c, "Invalid user context")
	return 0, errors.New("invalid user context")
}