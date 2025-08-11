package constants

// Coupon type constants
const (
	CouponTypePercentage  = "percentage"
	CouponTypeFixedAmount = "fixed_amount"
)

// Coupon status constants
const (
	CouponStatusActive   = "active"
	CouponStatusInactive = "inactive"
	CouponStatusExpired  = "expired"
)

// Default values
const (
	DefaultCurrency        = "USD"
	DefaultMaxUsesPerUser  = 1
	DefaultMaxDiscountRate = 100.0 // Maximum percentage discount
)

// Validation limits
const (
	MinCouponCodeLength = 3
	MaxCouponCodeLength = 50
	MinCouponNameLength = 1
	MaxCouponNameLength = 100
	MaxDescriptionLength = 1000
	MaxCurrencyLength = 3
)

// Search and pagination limits
const (
	DefaultPageSize     = 10
	MaxPageSize         = 100
	MaxBulkOperations   = 1000
	DefaultSearchLimit  = 50
)