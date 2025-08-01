package command

import (
	"context"
	"fmt"
	"time"
	
	"linke/internal/coupon/domain/aggregate"
	"linke/internal/coupon/domain/entity"
	"linke/internal/coupon/domain/repository"
	"linke/internal/coupon/domain/service"
	"linke/internal/coupon/domain/valueobject"
	"linke/internal/shared/domain"
	sharedvo "linke/internal/shared/valueobject"
)

// CouponCommandHandler handles coupon-related commands
type CouponCommandHandler struct {
	couponRepository repository.CouponRepository
	domainService    *service.CouponDomainService
	eventPublisher   domain.EventPublisher
	transactionManager domain.TransactionManager
}

// NewCouponCommandHandler creates a new coupon command handler
func NewCouponCommandHandler(
	couponRepository repository.CouponRepository,
	domainService *service.CouponDomainService,
	eventPublisher domain.EventPublisher,
	transactionManager domain.TransactionManager,
) *CouponCommandHandler {
	return &CouponCommandHandler{
		couponRepository:   couponRepository,
		domainService:      domainService,
		eventPublisher:     eventPublisher,
		transactionManager: transactionManager,
	}
}

// CreateCouponResult represents the result of creating a coupon
type CreateCouponResult struct {
	Coupon *aggregate.Coupon `json:"coupon"`
}

// Handle processes a create coupon command
func (h *CouponCommandHandler) Handle(ctx context.Context, cmd *CreateCouponCommand) (*CreateCouponResult, error) {
	var result *CreateCouponResult
	
	err := h.transactionManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Create value objects
		couponCode, err := valueobject.NewCouponCode(cmd.Code)
		if err != nil {
			return fmt.Errorf("invalid coupon code: %w", err)
		}
		
		// Validate unique code
		if err := h.domainService.ValidateUniqueCouponCode(txCtx, couponCode); err != nil {
			return err
		}
		
		couponType, err := valueobject.NewCouponType(cmd.Type)
		if err != nil {
			return fmt.Errorf("invalid coupon type: %w", err)
		}
		
		discountValue, err := valueobject.NewDiscountValue(cmd.Value, couponType)
		if err != nil {
			return fmt.Errorf("invalid discount value: %w", err)
		}
		
		// Set default currency if not provided
		currency := "USD"
		if cmd.Currency != "" {
			currency = cmd.Currency
		}
		
		minOrderAmount, err := valueobject.NewMoney(cmd.MinOrderAmount, currency)
		if err != nil {
			return fmt.Errorf("invalid minimum order amount: %w", err)
		}
		
		validityPeriod, err := valueobject.NewValidityPeriod(cmd.ValidFrom, cmd.ValidUntil)
		if err != nil {
			return fmt.Errorf("invalid validity period: %w", err)
		}
		
		// Set default max uses per user if not provided
		maxUsesPerUser := 1
		if cmd.MaxUsesPerUser > 0 {
			maxUsesPerUser = cmd.MaxUsesPerUser
		}
		
		usageLimits, err := valueobject.NewUsageLimits(cmd.MaxUses, 0, maxUsesPerUser)
		if err != nil {
			return fmt.Errorf("invalid usage limits: %w", err)
		}
		
		domainCreatedBy, err := sharedvo.NewUserIDFromUint64(cmd.CreatedBy)
		if err != nil {
			return fmt.Errorf("invalid created by user ID: %w", err)
		}
		
		// Convert to shared types for domain service validation
		sharedMinOrderAmount, err := valueobject.ConvertToSharedMoney(minOrderAmount)
		if err != nil {
			return fmt.Errorf("failed to convert min order amount: %w", err)
		}
		
		// Validate business rules
		if err := h.domainService.ValidateCouponBusinessRules(
			couponType, discountValue, sharedMinOrderAmount, validityPeriod, usageLimits,
		); err != nil {
			return fmt.Errorf("business rule validation failed: %w", err)
		}
		
		// Convert to shared types for aggregate creation
		createdBy := domainCreatedBy
		if err != nil {
			return fmt.Errorf("failed to convert created by: %w", err)
		}
		
		// Generate new ID (this would typically come from your ID generation strategy)
		now := time.Now()
		couponID := valueobject.NewCouponID(uint64(now.UnixNano()) % 1000000000) // Simple ID generation
		
		// Create coupon aggregate
		coupon, err := aggregate.NewCoupon(
			couponID,
			couponCode,
			cmd.Name,
			cmd.Description,
			discountValue,
			sharedMinOrderAmount,
			validityPeriod,
			usageLimits,
			cmd.ApplicablePlans,
			cmd.IsPublic,
			createdBy,
		)
		if err != nil {
			return fmt.Errorf("failed to create coupon: %w", err)
		}
		
		// Save coupon
		if err := h.couponRepository.Save(txCtx, coupon); err != nil {
			return fmt.Errorf("failed to save coupon: %w", err)
		}
		
		// Publish domain events
		for _, event := range coupon.DomainEvents() {
			if err := h.eventPublisher.Publish(txCtx, event); err != nil {
				return fmt.Errorf("failed to publish event: %w", err)
			}
		}
		coupon.ClearDomainEvents()
		
		result = &CreateCouponResult{Coupon: coupon}
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateCouponResult represents the result of updating a coupon
type UpdateCouponResult struct {
	Coupon *aggregate.Coupon `json:"coupon"`
}

// HandleUpdate processes an update coupon command
func (h *CouponCommandHandler) HandleUpdate(ctx context.Context, cmd *UpdateCouponCommand) (*UpdateCouponResult, error) {
	var result *UpdateCouponResult
	
	err := h.transactionManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Find existing coupon
		couponID := valueobject.NewCouponID(cmd.CouponID)
		coupon, err := h.couponRepository.FindByID(txCtx, couponID)
		if err != nil {
			return fmt.Errorf("coupon not found: %w", err)
		}
		
		// Update basic info if provided
		if cmd.Name != nil || cmd.Description != nil {
			name := coupon.Name()
			description := coupon.Description()
			
			if cmd.Name != nil {
				name = *cmd.Name
			}
			if cmd.Description != nil {
				description = *cmd.Description
			}
			
			if err := coupon.UpdateBasicInfo(name, description); err != nil {
				return fmt.Errorf("failed to update basic info: %w", err)
			}
		}
		
		// Update discount value if provided
		if cmd.Type != nil || cmd.Value != nil {
			couponType := coupon.DiscountValue().Type()
			value := coupon.DiscountValue().Value()
			
			if cmd.Type != nil {
				newType, err := valueobject.NewCouponType(*cmd.Type)
				if err != nil {
					return fmt.Errorf("invalid coupon type: %w", err)
				}
				couponType = newType
			}
			if cmd.Value != nil {
				value = *cmd.Value
			}
			
			newDiscountValue, err := valueobject.NewDiscountValue(value, couponType)
			if err != nil {
				return fmt.Errorf("invalid discount value: %w", err)
			}
			
			coupon.UpdateDiscountValue(newDiscountValue)
		}
		
		// Update minimum order amount if provided
		if cmd.MinOrderAmount != nil {
			domainNewMinOrderAmount, err := valueobject.NewMoney(*cmd.MinOrderAmount, coupon.MinOrderAmount().Currency().Code())
			if err != nil {
				return fmt.Errorf("invalid minimum order amount: %w", err)
			}
			
			// Convert to shared type for aggregate method
			newMinOrderAmount, err := valueobject.ConvertToSharedMoney(domainNewMinOrderAmount)
			if err != nil {
				return fmt.Errorf("failed to convert min order amount: %w", err)
			}
			
			coupon.UpdateMinOrderAmount(newMinOrderAmount)
		}
		
		// Update validity period if provided
		if cmd.ValidFrom != nil || cmd.ValidUntil != nil {
			validFrom := coupon.ValidityPeriod().ValidFrom()
			validUntil := coupon.ValidityPeriod().ValidUntil()
			
			if cmd.ValidFrom != nil {
				validFrom = cmd.ValidFrom
			}
			if cmd.ValidUntil != nil {
				validUntil = cmd.ValidUntil
			}
			
			newValidityPeriod, err := valueobject.NewValidityPeriod(validFrom, validUntil)
			if err != nil {
				return fmt.Errorf("invalid validity period: %w", err)
			}
			
			coupon.UpdateValidityPeriod(newValidityPeriod)
		}
		
		// Update applicable plans if provided
		if cmd.ApplicablePlans != nil {
			coupon.UpdateApplicablePlans(cmd.ApplicablePlans)
		}
		
		// Update visibility if provided
		if cmd.IsPublic != nil {
			coupon.UpdateVisibility(*cmd.IsPublic)
		}
		
		// Update status if provided
		if cmd.Status != nil {
			newStatus, err := valueobject.NewCouponStatus(*cmd.Status)
			if err != nil {
				return fmt.Errorf("invalid status: %w", err)
			}
			
			domainUpdatedBy, err := sharedvo.NewUserIDFromUint64(cmd.UpdatedBy)
			if err != nil {
				return fmt.Errorf("invalid updated by user ID: %w", err)
			}
			// Convert to shared type for aggregate method
			updatedBy := domainUpdatedBy
			if err != nil {
				return fmt.Errorf("failed to convert updated by: %w", err)
			}
			
			if err := coupon.ChangeStatus(newStatus, updatedBy, "Updated via API"); err != nil {
				return fmt.Errorf("failed to change status: %w", err)
			}
		}
		
		// Save updated coupon
		if err := h.couponRepository.Save(txCtx, coupon); err != nil {
			return fmt.Errorf("failed to save updated coupon: %w", err)
		}
		
		// Publish domain events
		for _, event := range coupon.DomainEvents() {
			if err := h.eventPublisher.Publish(txCtx, event); err != nil {
				return fmt.Errorf("failed to publish event: %w", err)
			}
		}
		coupon.ClearDomainEvents()
		
		result = &UpdateCouponResult{Coupon: coupon}
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteCouponResult represents the result of deleting a coupon
type DeleteCouponResult struct {
	Success bool `json:"success"`
}

// HandleDelete processes a delete coupon command
func (h *CouponCommandHandler) HandleDelete(ctx context.Context, cmd *DeleteCouponCommand) (*DeleteCouponResult, error) {
	var result *DeleteCouponResult
	
	err := h.transactionManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Find existing coupon
		couponID := valueobject.NewCouponID(cmd.CouponID)
		coupon, err := h.couponRepository.FindByID(txCtx, couponID)
		if err != nil {
			return fmt.Errorf("coupon not found: %w", err)
		}
		
		// Soft delete the coupon
		coupon.SoftDelete()
		
		// Save updated coupon
		if err := h.couponRepository.Save(txCtx, coupon); err != nil {
			return fmt.Errorf("failed to delete coupon: %w", err)
		}
		
		result = &DeleteCouponResult{Success: true}
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UseCouponResult represents the result of using a coupon
type UseCouponResult struct {
	DiscountAmount float64 `json:"discount_amount"`
	FinalAmount    float64 `json:"final_amount"`
	Success        bool    `json:"success"`
	Message        string  `json:"message"`
}

// HandleUse processes a use coupon command
func (h *CouponCommandHandler) HandleUse(ctx context.Context, cmd *UseCouponCommand) (*UseCouponResult, error) {
	var result *UseCouponResult
	
	err := h.transactionManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Validate and calculate discount
		couponCode, err := valueobject.NewCouponCode(cmd.CouponCode)
		if err != nil {
			return fmt.Errorf("invalid coupon code: %w", err)
		}
		
		domainUserID, err := sharedvo.NewUserIDFromUint64(cmd.UserID)
		if err != nil {
			return fmt.Errorf("invalid user ID: %w", err)
		}
		domainOrderAmount, err := valueobject.NewMoney(cmd.OrderAmount, cmd.Currency)
		if err != nil {
			return fmt.Errorf("invalid order amount: %w", err)
		}
		
		// Convert to shared types for domain service
		userID := domainUserID
		if err != nil {
			return fmt.Errorf("failed to convert user ID: %w", err)
		}
		
		orderAmount, err := valueobject.ConvertToSharedMoney(domainOrderAmount)
		if err != nil {
			return fmt.Errorf("failed to convert order amount: %w", err)
		}
		
		discountCalc, err := h.domainService.ValidateAndCalculateDiscount(
			txCtx, couponCode, userID, orderAmount, cmd.PlanID,
		)
		if err != nil {
			return fmt.Errorf("failed to validate coupon: %w", err)
		}
		
		if !discountCalc.IsValid {
			result = &UseCouponResult{
				Success: false,
				Message: discountCalc.ValidationMessage,
			}
			return nil
		}
		
		// Generate usage ID
		usageID := entity.NewCouponUsageID(uint64(time.Now().UnixNano()) % 1000000000)
		
		// Use the coupon
		if err := discountCalc.Coupon.Use(usageID, userID, cmd.OrderID, orderAmount); err != nil {
			return fmt.Errorf("failed to use coupon: %w", err)
		}
		
		// Save updated coupon
		if err := h.couponRepository.Save(txCtx, discountCalc.Coupon); err != nil {
			return fmt.Errorf("failed to save coupon usage: %w", err)
		}
		
		// Publish domain events
		for _, event := range discountCalc.Coupon.DomainEvents() {
			if err := h.eventPublisher.Publish(txCtx, event); err != nil {
				return fmt.Errorf("failed to publish event: %w", err)
			}
		}
		discountCalc.Coupon.ClearDomainEvents()
		
		result = &UseCouponResult{
			DiscountAmount: discountCalc.DiscountAmount.Amount(),
			FinalAmount:    discountCalc.FinalAmount.Amount(),
			Success:        true,
			Message:        "Coupon applied successfully",
		}
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ChangeCouponStatusResult represents the result of changing coupon status
type ChangeCouponStatusResult struct {
	Coupon *aggregate.Coupon `json:"coupon"`
}

// HandleChangeStatus processes a change coupon status command
func (h *CouponCommandHandler) HandleChangeStatus(ctx context.Context, cmd *ChangeCouponStatusCommand) (*ChangeCouponStatusResult, error) {
	var result *ChangeCouponStatusResult
	
	err := h.transactionManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Find existing coupon
		couponID := valueobject.NewCouponID(cmd.CouponID)
		coupon, err := h.couponRepository.FindByID(txCtx, couponID)
		if err != nil {
			return fmt.Errorf("coupon not found: %w", err)
		}
		
		// Change status
		newStatus, err := valueobject.NewCouponStatus(cmd.NewStatus)
		if err != nil {
			return fmt.Errorf("invalid status: %w", err)
		}
		
		domainChangedBy, err := sharedvo.NewUserIDFromUint64(cmd.ChangedBy)
		if err != nil {
			return fmt.Errorf("invalid changed by user ID: %w", err)
		}
		// Convert to shared type for aggregate method
		changedBy := domainChangedBy
		if err != nil {
			return fmt.Errorf("failed to convert changed by: %w", err)
		}
		
		if err := coupon.ChangeStatus(newStatus, changedBy, cmd.Reason); err != nil {
			return fmt.Errorf("failed to change status: %w", err)
		}
		
		// Save updated coupon
		if err := h.couponRepository.Save(txCtx, coupon); err != nil {
			return fmt.Errorf("failed to save coupon: %w", err)
		}
		
		// Publish domain events
		for _, event := range coupon.DomainEvents() {
			if err := h.eventPublisher.Publish(txCtx, event); err != nil {
				return fmt.Errorf("failed to publish event: %w", err)
			}
		}
		coupon.ClearDomainEvents()
		
		result = &ChangeCouponStatusResult{Coupon: coupon}
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ExpireExpiredCouponsResult represents the result of expiring expired coupons
type ExpireExpiredCouponsResult struct {
	ExpiredCount int `json:"expired_count"`
}

// HandleExpireExpired processes an expire expired coupons command
func (h *CouponCommandHandler) HandleExpireExpired(ctx context.Context, cmd *ExpireExpiredCouponsCommand) (*ExpireExpiredCouponsResult, error) {
	expiredCount, err := h.domainService.ExpireCoupons(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to expire coupons: %w", err)
	}
	
	return &ExpireExpiredCouponsResult{ExpiredCount: expiredCount}, nil
}