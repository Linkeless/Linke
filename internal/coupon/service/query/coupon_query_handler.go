package query

import (
	"context"
	"fmt"
	
	"linke/internal/coupon/domain/aggregate"
	"linke/internal/coupon/domain/repository"
	"linke/internal/coupon/domain/service"
	"linke/internal/coupon/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// CouponQueryHandler handles coupon-related queries
type CouponQueryHandler struct {
	couponRepository repository.CouponRepository
	domainService    *service.CouponDomainService
}

// NewCouponQueryHandler creates a new coupon query handler
func NewCouponQueryHandler(
	couponRepository repository.CouponRepository,
	domainService *service.CouponDomainService,
) *CouponQueryHandler {
	return &CouponQueryHandler{
		couponRepository: couponRepository,
		domainService:    domainService,
	}
}

// GetCouponByIDResult represents the result of getting a coupon by ID
type GetCouponByIDResult struct {
	Coupon *aggregate.Coupon `json:"coupon"`
}

// Handle processes a get coupon by ID query
func (h *CouponQueryHandler) Handle(ctx context.Context, query *GetCouponByIDQuery) (*GetCouponByIDResult, error) {
	couponID := valueobject.NewCouponID(query.CouponID)
	coupon, err := h.couponRepository.FindByID(ctx, couponID)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}
	
	return &GetCouponByIDResult{Coupon: coupon}, nil
}

// GetCouponByCodeResult represents the result of getting a coupon by code
type GetCouponByCodeResult struct {
	Coupon *aggregate.Coupon `json:"coupon"`
}

// HandleGetByCode processes a get coupon by code query
func (h *CouponQueryHandler) HandleGetByCode(ctx context.Context, query *GetCouponByCodeQuery) (*GetCouponByCodeResult, error) {
	couponCode, err := valueobject.NewCouponCode(query.Code)
	if err != nil {
		return nil, fmt.Errorf("invalid coupon code: %w", err)
	}
	
	coupon, err := h.couponRepository.FindByCode(ctx, couponCode)
	if err != nil {
		return nil, fmt.Errorf("coupon not found: %w", err)
	}
	
	return &GetCouponByCodeResult{Coupon: coupon}, nil
}

// GetCouponsResult represents the result of getting coupons
type GetCouponsResult struct {
	Coupons    []*aggregate.Coupon `json:"coupons"`
	TotalCount int64              `json:"total_count"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
}

// HandleGetCoupons processes a get coupons query
func (h *CouponQueryHandler) HandleGetCoupons(ctx context.Context, query *GetCouponsQuery) (*GetCouponsResult, error) {
	// Build filter
	filter := repository.NewCouponFilter()
	
	// Apply status filter
	if query.Status != nil {
		status, err := valueobject.NewCouponStatus(*query.Status)
		if err != nil {
			return nil, fmt.Errorf("invalid status: %w", err)
		}
		filter = filter.WithStatus(status)
	}
	
	// Apply type filter
	if query.Type != nil {
		couponType, err := valueobject.NewCouponType(*query.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid type: %w", err)
		}
		filter = filter.WithType(couponType)
	}
	
	// Apply public visibility filter
	if query.IsPublic != nil {
		filter = filter.WithPublicVisibility(*query.IsPublic)
	}
	
	// Apply created by filter
	if query.CreatedBy != nil {
		createdBy, err := sharedvo.NewUserIDFromUint64(*query.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("invalid created by user ID: %w", err)
		}
		filter = filter.WithCreatedBy(createdBy)
	}
	
	// Apply search filters
	if query.CodeSearch != "" {
		filter = filter.WithCodeSearch(query.CodeSearch)
	}
	if query.NameSearch != "" {
		filter = filter.WithNameSearch(query.NameSearch)
	}
	
	// Apply options
	filter = filter.WithIncludeDeleted(query.IncludeDeleted)
	
	// Apply pagination
	limit := 50 // default
	if query.Limit > 0 {
		limit = query.Limit
	}
	offset := 0
	if query.Offset > 0 {
		offset = query.Offset
	}
	filter = filter.WithPagination(limit, offset)
	
	// Apply sorting
	sortBy := "created_at" // default
	if query.SortBy != "" {
		sortBy = query.SortBy
	}
	sortOrder := "desc" // default
	if query.SortOrder != "" {
		sortOrder = query.SortOrder
	}
	filter = filter.WithSorting(sortBy, sortOrder)
	
	// Execute query
	coupons, totalCount, err := h.couponRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupons: %w", err)
	}
	
	return &GetCouponsResult{
		Coupons:    coupons,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

// GetPublicCouponsResult represents the result of getting public coupons
type GetPublicCouponsResult struct {
	Coupons []*aggregate.Coupon `json:"coupons"`
}

// HandleGetPublicCoupons processes a get public coupons query
func (h *CouponQueryHandler) HandleGetPublicCoupons(ctx context.Context, query *GetPublicCouponsQuery) (*GetPublicCouponsResult, error) {
	coupons, err := h.couponRepository.FindPublicCoupons(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get public coupons: %w", err)
	}
	
	return &GetPublicCouponsResult{Coupons: coupons}, nil
}

// ValidateCouponResult represents the result of validating a coupon
type ValidateCouponResult struct {
	Valid          bool                `json:"valid"`
	Message        string              `json:"message"`
	DiscountAmount float64             `json:"discount_amount"`
	FinalAmount    float64             `json:"final_amount"`
	Currency       string              `json:"currency"`
	Coupon         *aggregate.Coupon   `json:"coupon,omitempty"`
}

// HandleValidateCoupon processes a validate coupon query
func (h *CouponQueryHandler) HandleValidateCoupon(ctx context.Context, query *ValidateCouponQuery) (*ValidateCouponResult, error) {
	// Create value objects
	couponCode, err := valueobject.NewCouponCode(query.Code)
	if err != nil {
		return &ValidateCouponResult{
			Valid:   false,
			Message: fmt.Sprintf("Invalid coupon code: %v", err),
		}, nil
	}
	
	domainUserID, err := sharedvo.NewUserIDFromUint64(query.UserID)
	if err != nil {
		return &ValidateCouponResult{
			Valid:   false,
			Message: fmt.Sprintf("Invalid user ID: %v", err),
		}, nil
	}
	domainOrderAmount, err := valueobject.NewMoney(query.OrderAmount, query.Currency)
	if err != nil {
		return &ValidateCouponResult{
			Valid:   false,
			Message: fmt.Sprintf("Invalid order amount: %v", err),
		}, nil
	}
	
	// Convert to shared types for domain service
	userID := domainUserID
	if err != nil {
		return nil, fmt.Errorf("failed to convert user ID: %w", err)
	}
	
	orderAmount, err := valueobject.ConvertToSharedMoney(domainOrderAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert order amount: %w", err)
	}
	
	// Validate and calculate discount
	discountCalc, err := h.domainService.ValidateAndCalculateDiscount(
		ctx, couponCode, userID, orderAmount, query.PlanID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to validate coupon: %w", err)
	}
	
	result := &ValidateCouponResult{
		Valid:    discountCalc.IsValid,
		Message:  discountCalc.ValidationMessage,
		Currency: query.Currency,
	}
	
	if discountCalc.IsValid {
		result.DiscountAmount = discountCalc.DiscountAmount.Amount()
		result.FinalAmount = discountCalc.FinalAmount.Amount()
		result.Coupon = discountCalc.Coupon
	}
	
	return result, nil
}