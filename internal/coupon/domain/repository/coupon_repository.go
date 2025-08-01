package repository

import (
	"context"
	
	"linke/internal/coupon/domain/aggregate"
	"linke/internal/coupon/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// CouponRepository defines the interface for coupon persistence
type CouponRepository interface {
	// Save saves or updates a coupon aggregate
	Save(ctx context.Context, coupon *aggregate.Coupon) error
	
	// FindByID finds a coupon by ID
	FindByID(ctx context.Context, id valueobject.CouponID) (*aggregate.Coupon, error)
	
	// FindByCode finds a coupon by code
	FindByCode(ctx context.Context, code valueobject.CouponCode) (*aggregate.Coupon, error)
	
	// FindActiveByCode finds an active coupon by code
	FindActiveByCode(ctx context.Context, code valueobject.CouponCode) (*aggregate.Coupon, error)
	
	// FindAll finds coupons with filtering and pagination
	FindAll(ctx context.Context, filter CouponFilter) ([]*aggregate.Coupon, int64, error)
	
	// FindPublicCoupons finds active and public coupons
	FindPublicCoupons(ctx context.Context) ([]*aggregate.Coupon, error)
	
	// Delete removes a coupon (hard delete)
	Delete(ctx context.Context, id valueobject.CouponID) error
	
	// ExistsByCode checks if a coupon with the given code exists
	ExistsByCode(ctx context.Context, code valueobject.CouponCode) (bool, error)
	
	// CountUserUsage counts how many times a user has used a specific coupon
	CountUserUsage(ctx context.Context, couponID valueobject.CouponID, userID sharedvo.UserID) (int, error)
	
	// FindExpiredCoupons finds coupons that have expired but status is still active
	FindExpiredCoupons(ctx context.Context) ([]*aggregate.Coupon, error)
}

// CouponFilter represents filtering criteria for coupon queries
type CouponFilter struct {
	// Status filter
	Status *valueobject.CouponStatus
	
	// Type filter
	Type *valueobject.CouponType
	
	// Public visibility filter
	IsPublic *bool
	
	// Created by filter
	CreatedBy *sharedvo.UserID
	
	// Code search (partial match)
	CodeSearch string
	
	// Name search (partial match)
	NameSearch string
	
	// Include deleted records
	IncludeDeleted bool
	
	// Pagination
	Limit  int
	Offset int
	
	// Sorting
	SortBy    string // "created_at", "updated_at", "name", "code", "status"
	SortOrder string // "asc", "desc"
}

// NewCouponFilter creates a new coupon filter with default values
func NewCouponFilter() CouponFilter {
	return CouponFilter{
		Limit:     50,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// WithStatus sets the status filter
func (f CouponFilter) WithStatus(status valueobject.CouponStatus) CouponFilter {
	f.Status = &status
	return f
}

// WithType sets the type filter
func (f CouponFilter) WithType(couponType valueobject.CouponType) CouponFilter {
	f.Type = &couponType
	return f
}

// WithPublicVisibility sets the public visibility filter
func (f CouponFilter) WithPublicVisibility(isPublic bool) CouponFilter {
	f.IsPublic = &isPublic
	return f
}

// WithCreatedBy sets the created by filter
func (f CouponFilter) WithCreatedBy(createdBy sharedvo.UserID) CouponFilter {
	f.CreatedBy = &createdBy
	return f
}

// WithCodeSearch sets the code search filter
func (f CouponFilter) WithCodeSearch(codeSearch string) CouponFilter {
	f.CodeSearch = codeSearch
	return f
}

// WithNameSearch sets the name search filter
func (f CouponFilter) WithNameSearch(nameSearch string) CouponFilter {
	f.NameSearch = nameSearch
	return f
}

// WithIncludeDeleted sets whether to include deleted records
func (f CouponFilter) WithIncludeDeleted(includeDeleted bool) CouponFilter {
	f.IncludeDeleted = includeDeleted
	return f
}

// WithPagination sets pagination parameters
func (f CouponFilter) WithPagination(limit, offset int) CouponFilter {
	f.Limit = limit
	f.Offset = offset
	return f
}

// WithSorting sets sorting parameters
func (f CouponFilter) WithSorting(sortBy, sortOrder string) CouponFilter {
	f.SortBy = sortBy
	f.SortOrder = sortOrder
	return f
}