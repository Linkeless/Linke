package service

import (
	"context"
	"fmt"
	"time"
	
	"linke/internal/coupon/domain/aggregate"
	"linke/internal/coupon/domain/repository"
	"linke/internal/coupon/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// CouponDomainService provides domain-specific business logic
type CouponDomainService struct {
	couponRepository repository.CouponRepository
}

// NewCouponDomainService creates a new coupon domain service
func NewCouponDomainService(couponRepository repository.CouponRepository) *CouponDomainService {
	return &CouponDomainService{
		couponRepository: couponRepository,
	}
}

// ValidateUniqueCouponCode ensures a coupon code is unique across the system
func (s *CouponDomainService) ValidateUniqueCouponCode(ctx context.Context, code valueobject.CouponCode) error {
	exists, err := s.couponRepository.ExistsByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to check coupon code uniqueness: %w", err)
	}
	
	if exists {
		return fmt.Errorf("coupon with code '%s' already exists", code.String())
	}
	
	return nil
}

// ValidateAndCalculateDiscount validates a coupon and calculates discount for an order
type DiscountCalculation struct {
	Coupon           *aggregate.Coupon
	DiscountAmount   sharedvo.Money
	FinalAmount      sharedvo.Money
	IsValid          bool
	ValidationMessage string
}

func (s *CouponDomainService) ValidateAndCalculateDiscount(
	ctx context.Context,
	code valueobject.CouponCode,
	userID sharedvo.UserID,
	orderAmount sharedvo.Money,
	planID uint64,
) (*DiscountCalculation, error) {
	// Find coupon by code
	coupon, err := s.couponRepository.FindActiveByCode(ctx, code)
	if err != nil {
		return &DiscountCalculation{
			IsValid:           false,
			ValidationMessage: "Coupon not found",
		}, nil
	}
	
	// Check if coupon can be used by this user
	canUse, reason := coupon.CanBeUsedBy(userID, orderAmount, planID)
	if !canUse {
		return &DiscountCalculation{
			Coupon:            coupon,
			IsValid:           false,
			ValidationMessage: reason,
		}, nil
	}
	
	// Check user's usage count from repository (convert to domain type for repository)
	domainUserID := userID
	userUsageCount, err := s.couponRepository.CountUserUsage(ctx, coupon.ID(), domainUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to count user usage: %w", err)
	}
	
	if userUsageCount >= coupon.UsageLimits().MaxUsesPerUser() {
		return &DiscountCalculation{
			Coupon:            coupon,
			IsValid:           false,
			ValidationMessage: "You have already used this coupon the maximum number of times",
		}, nil
	}
	
	// Calculate discount
	discountAmount, err := coupon.CalculateDiscount(orderAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate discount: %w", err)
	}
	
	// Calculate final amount
	finalAmount, err := orderAmount.Subtract(discountAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate final amount: %w", err)
	}
	
	return &DiscountCalculation{
		Coupon:            coupon,
		DiscountAmount:    discountAmount,
		FinalAmount:       finalAmount,
		IsValid:           true,
		ValidationMessage: "Coupon is valid",
	}, nil
}

// ExpireCoupons finds and expires coupons that have passed their validity period
func (s *CouponDomainService) ExpireCoupons(ctx context.Context) (int, error) {
	// Find expired coupons
	expiredCoupons, err := s.couponRepository.FindExpiredCoupons(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to find expired coupons: %w", err)
	}
	
	expiredCount := 0
	for _, coupon := range expiredCoupons {
		// Expire the coupon
		if err := coupon.Expire(); err != nil {
			// Log error but continue with other coupons
			continue
		}
		
		// Save the updated coupon
		if err := s.couponRepository.Save(ctx, coupon); err != nil {
			// Log error but continue with other coupons
			continue
		}
		
		expiredCount++
	}
	
	return expiredCount, nil
}

// ValidateCouponBusinessRules validates complex business rules for coupon creation/updates
func (s *CouponDomainService) ValidateCouponBusinessRules(
	couponType valueobject.CouponType,
	discountValue valueobject.DiscountValue,
	minOrderAmount sharedvo.Money,
	validityPeriod valueobject.ValidityPeriod,
	usageLimits valueobject.UsageLimits,
) error {
	// Validate discount value makes sense with type
	if !discountValue.Type().Equals(couponType) {
		return fmt.Errorf("discount value type %s does not match coupon type %s", 
			discountValue.Type().String(), couponType.String())
	}
	
	// For percentage discounts, ensure reasonable limits
	if couponType.IsPercentage() && discountValue.Value() > 90 {
		return fmt.Errorf("percentage discount cannot exceed 90%% for business policy reasons")
	}
	
	// For fixed amount discounts, ensure it's reasonable compared to minimum order
	if couponType.IsFixedAmount() && !minOrderAmount.IsZero() {
		fixedDiscount, err := sharedvo.NewMoney(discountValue.Value(), minOrderAmount.Currency())
		if err != nil {
			return fmt.Errorf("failed to create fixed discount money: %w", err)
		}
		
		if isGreater, err := fixedDiscount.GreaterThanOrEqual(minOrderAmount); err != nil {
			return fmt.Errorf("failed to compare discount and minimum order amounts: %w", err)
		} else if isGreater {
			return fmt.Errorf("fixed discount amount cannot be greater than or equal to minimum order amount")
		}
	}
	
	// Validate validity period makes business sense
	if validityPeriod.HasStartTime() && validityPeriod.HasEndTime() {
		duration := validityPeriod.Duration()
		if duration != nil && *duration > 365*24*time.Hour {
			return fmt.Errorf("coupon validity period cannot exceed 1 year")
		}
	}
	
	// Validate usage limits make sense
	if !usageLimits.IsUnlimited() && usageLimits.MaxUses() > 10000 {
		return fmt.Errorf("maximum uses cannot exceed 10,000 for system performance reasons")
	}
	
	if usageLimits.MaxUsesPerUser() > 100 {
		return fmt.Errorf("maximum uses per user cannot exceed 100")
	}
	
	return nil
}

// CalculateSystemDiscountImpact calculates the potential financial impact of a coupon
type DiscountImpact struct {
	EstimatedMaxDiscount sharedvo.Money
	RiskLevel           string // "low", "medium", "high"
	Recommendations     []string
}

func (s *CouponDomainService) CalculateSystemDiscountImpact(
	couponType valueobject.CouponType,
	discountValue valueobject.DiscountValue,
	minOrderAmount sharedvo.Money,
	usageLimits valueobject.UsageLimits,
	averageOrderAmount sharedvo.Money,
) (*DiscountImpact, error) {
	// Calculate potential discount per use
	var discountPerUse sharedvo.Money
	var err error
	
	if couponType.IsPercentage() {
		discountPerUse, err = averageOrderAmount.MultiplyByPercentage(discountValue.Value())
		if err != nil {
			return nil, fmt.Errorf("failed to calculate percentage discount: %w", err)
		}
	} else {
		currency, currErr := sharedvo.NewCurrency(averageOrderAmount.Currency().Code())
		if currErr != nil {
			return nil, fmt.Errorf("failed to create currency: %w", currErr)
		}
		discountPerUse, err = sharedvo.NewMoney(discountValue.Value(), currency)
		if err != nil {
			return nil, fmt.Errorf("failed to create fixed discount: %w", err)
		}
	}
	
	// Calculate maximum potential impact
	var maxDiscount sharedvo.Money
	if usageLimits.IsUnlimited() {
		// For unlimited coupons, estimate based on a reasonable cap
		cappedUses := 1000 // Business assumption
		for i := 0; i < cappedUses; i++ {
			maxDiscount, err = maxDiscount.Add(discountPerUse)
			if err != nil {
				return nil, fmt.Errorf("failed to add discount amounts: %w", err)
			}
		}
	} else {
		for i := 0; i < usageLimits.MaxUses(); i++ {
			maxDiscount, err = maxDiscount.Add(discountPerUse)
			if err != nil {
				return nil, fmt.Errorf("failed to add discount amounts: %w", err)
			}
		}
	}
	
	// Determine risk level
	riskLevel := "low"
	recommendations := make([]string, 0)
	
	// Risk based on total potential discount
	highRiskThreshold, _ := sharedvo.NewMoney(10000, maxDiscount.Currency())
	mediumRiskThreshold, _ := sharedvo.NewMoney(1000, maxDiscount.Currency())
	
	if maxDiscountGreater, _ := maxDiscount.GreaterThanOrEqual(highRiskThreshold); maxDiscountGreater {
		riskLevel = "high"
		recommendations = append(recommendations, "Consider reducing usage limits or discount value")
		recommendations = append(recommendations, "Implement additional approval processes")
	} else if mediumDiscountGreater, _ := maxDiscount.GreaterThanOrEqual(mediumRiskThreshold); mediumDiscountGreater {
		riskLevel = "medium"
		recommendations = append(recommendations, "Monitor usage closely")
		recommendations = append(recommendations, "Consider setting expiration date")
	}
	
	// Additional recommendations based on coupon characteristics
	if usageLimits.IsUnlimited() {
		recommendations = append(recommendations, "Consider setting usage limits for better control")
	}
	
	if couponType.IsPercentage() && discountValue.Value() > 50 {
		recommendations = append(recommendations, "High percentage discounts may impact profit margins significantly")
	}
	
	return &DiscountImpact{
		EstimatedMaxDiscount: maxDiscount,
		RiskLevel:           riskLevel,
		Recommendations:     recommendations,
	}, nil
}