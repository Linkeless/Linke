package query

// GetCouponByIDQuery represents a query to get a coupon by ID
type GetCouponByIDQuery struct {
	CouponID uint64 `json:"coupon_id" validate:"required"`
}

// GetCouponByCodeQuery represents a query to get a coupon by code
type GetCouponByCodeQuery struct {
	Code string `json:"code" validate:"required"`
}

// GetCouponsQuery represents a query to get coupons with filtering
type GetCouponsQuery struct {
	// Filters
	Status     *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive expired"`
	Type       *string `json:"type,omitempty" validate:"omitempty,oneof=percentage fixed_amount"`
	IsPublic   *bool   `json:"is_public,omitempty"`
	CreatedBy  *uint64 `json:"created_by,omitempty"`
	CodeSearch string  `json:"code_search,omitempty"`
	NameSearch string  `json:"name_search,omitempty"`
	
	// Options
	IncludeDeleted bool `json:"include_deleted,omitempty"`
	
	// Pagination
	Limit  int `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset int `json:"offset,omitempty" validate:"omitempty,min=0"`
	
	// Sorting
	SortBy    string `json:"sort_by,omitempty" validate:"omitempty,oneof=created_at updated_at name code status"`
	SortOrder string `json:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// GetPublicCouponsQuery represents a query to get public coupons
type GetPublicCouponsQuery struct {
	// No parameters needed - returns all active public coupons
}

// ValidateCouponQuery represents a query to validate a coupon
type ValidateCouponQuery struct {
	Code        string  `json:"code" validate:"required"`
	UserID      uint64  `json:"user_id" validate:"required"`
	OrderAmount float64 `json:"order_amount" validate:"required,min=0"`
	Currency    string  `json:"currency" validate:"required,len=3"`
	PlanID      uint64  `json:"plan_id" validate:"required"`
}

// GetCouponUsagesQuery represents a query to get coupon usage records
type GetCouponUsagesQuery struct {
	// Filters
	CouponID *uint64 `json:"coupon_id,omitempty"`
	UserID   *uint64 `json:"user_id,omitempty"`
	
	// Pagination
	Limit  int `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset int `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// GetCouponStatsQuery represents a query to get coupon statistics
type GetCouponStatsQuery struct {
	CouponID *uint64 `json:"coupon_id,omitempty"`
	UserID   *uint64 `json:"user_id,omitempty"`
}