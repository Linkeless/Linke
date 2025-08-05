package implementations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/logger"
)

// paymentMethodService implements the PaymentMethodService interface
type paymentMethodService struct {
	paymentMethodRepo interfaces.PaymentMethodRepository
	paymentService    interfaces.PaymentService
	logger            logger.Logger
}

// NewPaymentMethodService creates a new payment method service instance
func NewPaymentMethodService(
	paymentMethodRepo interfaces.PaymentMethodRepository,
	paymentService interfaces.PaymentService,
	logger logger.Logger,
) interfaces.PaymentMethodService {
	return &paymentMethodService{
		paymentMethodRepo: paymentMethodRepo,
		paymentService:    paymentService,
		logger:            logger,
	}
}

// Configuration constants
const (
	MaxPaymentMethodsPerUser = 10
	MinValidationInterval    = 30 * 24 * time.Hour // 30 days
	HighFailureRateThreshold = 0.5                 // 50% failure rate
)

// CreatePaymentMethod creates a new payment method for a user
func (s *paymentMethodService) CreatePaymentMethod(ctx context.Context, userID uint, req *entities.CreatePaymentMethodRequest) (*entities.PaymentMethodResponse, error) {
	s.logger.Info("Creating payment method", logger.Uint("user_id", userID), logger.String("gateway", req.Gateway), logger.String("method", req.Method))

	// Check if user has reached the payment method limit
	limitReached, err := s.IsPaymentMethodLimitReached(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check payment method limit: %w", err)
	}
	if limitReached {
		return nil, fmt.Errorf("maximum number of payment methods reached (%d)", MaxPaymentMethodsPerUser)
	}

	// Validate payment token with gateway (mock implementation - replace with actual gateway validation)
	if err := s.validatePaymentTokenWithGateway(req.Gateway, req.PaymentToken); err != nil {
		return nil, fmt.Errorf("payment token validation failed: %w", err)
	}

	// Check if payment token already exists for this gateway
	existingMethod, err := s.paymentMethodRepo.GetByPaymentToken(ctx, req.Gateway, req.PaymentToken)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing payment token: %w", err)
	}
	if existingMethod != nil {
		return nil, fmt.Errorf("payment method already exists")
	}

	// Create payment method entity
	paymentMethod := &entities.PaymentMethod{
		UserID:            userID,
		Type:              req.Type,
		Gateway:           req.Gateway,
		Method:            req.Method,
		DisplayName:       req.DisplayName,
		PaymentToken:      req.PaymentToken,
		GatewayCustomerID: req.GatewayCustomerID,
		MaskedInfo:        req.MaskedInfo,
		Brand:             req.Brand,
		ExpiryMonth:       req.ExpiryMonth,
		ExpiryYear:        req.ExpiryYear,
		IsDefault:         false, // Will be set later if requested
		Active:            true,
		Status:            entities.PaymentMethodStatusActive,
		BillingCountry:    req.BillingCountry,
		BillingPostcode:   req.BillingPostcode,
		LastValidatedAt:   timePtr(time.Now()),
	}

	// Create the payment method
	if err := s.paymentMethodRepo.Create(ctx, paymentMethod); err != nil {
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}

	// Set as default if requested
	if req.SetAsDefault {
		if err := s.paymentMethodRepo.SetAsDefault(ctx, userID, paymentMethod.ID); err != nil {
			s.logger.Error("Failed to set payment method as default", logger.ErrorField(err), logger.Uint("payment_method_id", paymentMethod.ID))
			// Don't fail the creation, just log the error
		} else {
			paymentMethod.IsDefault = true
		}
	}

	s.logger.Info("Payment method created successfully", logger.Uint("payment_method_id", paymentMethod.ID), logger.Uint("user_id", userID))
	return paymentMethod.ToResponse(), nil
}

// GetPaymentMethod retrieves a payment method by ID for a specific user
func (s *paymentMethodService) GetPaymentMethod(ctx context.Context, userID, paymentMethodID uint) (*entities.PaymentMethodResponse, error) {
	// Validate ownership
	isOwner, err := s.paymentMethodRepo.ValidateOwnership(ctx, paymentMethodID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ownership: %w", err)
	}
	if !isOwner {
		return nil, fmt.Errorf("payment method not found or access denied")
	}

	paymentMethod, err := s.paymentMethodRepo.GetByID(ctx, paymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}
	if paymentMethod == nil {
		return nil, fmt.Errorf("payment method not found")
	}

	return paymentMethod.ToResponse(), nil
}

// ListPaymentMethods retrieves all payment methods for a user
func (s *paymentMethodService) ListPaymentMethods(ctx context.Context, userID uint) (*entities.PaymentMethodListResponse, error) {
	paymentMethods, err := s.paymentMethodRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list payment methods: %w", err)
	}

	responses := make([]entities.PaymentMethodResponse, len(paymentMethods))
	var defaultMethod *entities.PaymentMethodResponse

	for i, pm := range paymentMethods {
		responses[i] = *pm.ToResponse()
		if pm.IsDefault {
			resp := pm.ToResponse()
			defaultMethod = resp
		}
	}

	return &entities.PaymentMethodListResponse{
		PaymentMethods: responses,
		Total:          len(responses),
		DefaultMethod:  defaultMethod,
	}, nil
}

// ListActivePaymentMethods retrieves all active payment methods for a user
func (s *paymentMethodService) ListActivePaymentMethods(ctx context.Context, userID uint) (*entities.PaymentMethodListResponse, error) {
	paymentMethods, err := s.paymentMethodRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active payment methods: %w", err)
	}

	responses := make([]entities.PaymentMethodResponse, len(paymentMethods))
	var defaultMethod *entities.PaymentMethodResponse

	for i, pm := range paymentMethods {
		responses[i] = *pm.ToResponse()
		if pm.IsDefault {
			resp := pm.ToResponse()
			defaultMethod = resp
		}
	}

	return &entities.PaymentMethodListResponse{
		PaymentMethods: responses,
		Total:          len(responses),
		DefaultMethod:  defaultMethod,
	}, nil
}

// ListPaymentMethodsByGateway retrieves payment methods for a user by gateway
func (s *paymentMethodService) ListPaymentMethodsByGateway(ctx context.Context, userID uint, gateway string) (*entities.PaymentMethodListResponse, error) {
	paymentMethods, err := s.paymentMethodRepo.GetByUserIDAndGateway(ctx, userID, gateway)
	if err != nil {
		return nil, fmt.Errorf("failed to list payment methods by gateway: %w", err)
	}

	responses := make([]entities.PaymentMethodResponse, len(paymentMethods))
	var defaultMethod *entities.PaymentMethodResponse

	for i, pm := range paymentMethods {
		responses[i] = *pm.ToResponse()
		if pm.IsDefault {
			resp := pm.ToResponse()
			defaultMethod = resp
		}
	}

	return &entities.PaymentMethodListResponse{
		PaymentMethods: responses,
		Total:          len(responses),
		DefaultMethod:  defaultMethod,
	}, nil
}

// UpdatePaymentMethod updates an existing payment method
func (s *paymentMethodService) UpdatePaymentMethod(ctx context.Context, userID, paymentMethodID uint, req *entities.UpdatePaymentMethodRequest) (*entities.PaymentMethodResponse, error) {
	// Validate ownership
	isOwner, err := s.paymentMethodRepo.ValidateOwnership(ctx, paymentMethodID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ownership: %w", err)
	}
	if !isOwner {
		return nil, fmt.Errorf("payment method not found or access denied")
	}

	paymentMethod, err := s.paymentMethodRepo.GetByID(ctx, paymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}
	if paymentMethod == nil {
		return nil, fmt.Errorf("payment method not found")
	}

	// Update fields if provided
	if req.DisplayName != nil {
		paymentMethod.DisplayName = *req.DisplayName
	}
	if req.ExpiryMonth != nil {
		paymentMethod.ExpiryMonth = req.ExpiryMonth
	}
	if req.ExpiryYear != nil {
		paymentMethod.ExpiryYear = req.ExpiryYear
	}
	if req.BillingCountry != nil {
		paymentMethod.BillingCountry = *req.BillingCountry
	}
	if req.BillingPostcode != nil {
		paymentMethod.BillingPostcode = *req.BillingPostcode
	}
	if req.Active != nil {
		paymentMethod.Active = *req.Active
		if !*req.Active {
			paymentMethod.Status = entities.PaymentMethodStatusInactive
		} else {
			paymentMethod.Status = entities.PaymentMethodStatusActive
		}
	}

	if err := s.paymentMethodRepo.Update(ctx, paymentMethod); err != nil {
		return nil, fmt.Errorf("failed to update payment method: %w", err)
	}

	s.logger.Info("Payment method updated", logger.Uint("payment_method_id", paymentMethodID), logger.Uint("user_id", userID))
	return paymentMethod.ToResponse(), nil
}

// SetDefaultPaymentMethod sets a payment method as the default for the user
func (s *paymentMethodService) SetDefaultPaymentMethod(ctx context.Context, userID, paymentMethodID uint) (*entities.PaymentMethodResponse, error) {
	// Validate ownership
	isOwner, err := s.paymentMethodRepo.ValidateOwnership(ctx, paymentMethodID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ownership: %w", err)
	}
	if !isOwner {
		return nil, fmt.Errorf("payment method not found or access denied")
	}

	paymentMethod, err := s.paymentMethodRepo.GetByID(ctx, paymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}
	if paymentMethod == nil {
		return nil, fmt.Errorf("payment method not found")
	}

	// Check if payment method can be used
	if !paymentMethod.CanBeUsedForPayment() {
		return nil, fmt.Errorf("payment method cannot be set as default (inactive or expired)")
	}

	if err := s.paymentMethodRepo.SetAsDefault(ctx, userID, paymentMethodID); err != nil {
		return nil, fmt.Errorf("failed to set default payment method: %w", err)
	}

	paymentMethod.IsDefault = true
	s.logger.Info("Payment method set as default", logger.Uint("payment_method_id", paymentMethodID), logger.Uint("user_id", userID))
	return paymentMethod.ToResponse(), nil
}

// DeletePaymentMethod soft deletes a payment method
func (s *paymentMethodService) DeletePaymentMethod(ctx context.Context, userID, paymentMethodID uint) error {
	// Validate ownership
	isOwner, err := s.paymentMethodRepo.ValidateOwnership(ctx, paymentMethodID, userID)
	if err != nil {
		return fmt.Errorf("failed to validate ownership: %w", err)
	}
	if !isOwner {
		return fmt.Errorf("payment method not found or access denied")
	}

	if err := s.paymentMethodRepo.Delete(ctx, paymentMethodID); err != nil {
		return fmt.Errorf("failed to delete payment method: %w", err)
	}

	s.logger.Info("Payment method deleted", logger.Uint("payment_method_id", paymentMethodID), logger.Uint("user_id", userID))
	return nil
}

// GetDefaultPaymentMethod retrieves the default payment method for a user
func (s *paymentMethodService) GetDefaultPaymentMethod(ctx context.Context, userID uint) (*entities.PaymentMethodResponse, error) {
	paymentMethod, err := s.paymentMethodRepo.GetDefaultByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get default payment method: %w", err)
	}
	if paymentMethod == nil {
		return nil, fmt.Errorf("no default payment method found")
	}

	return paymentMethod.ToResponse(), nil
}

// GetDefaultPaymentMethodByGateway retrieves the default payment method for a user and gateway
func (s *paymentMethodService) GetDefaultPaymentMethodByGateway(ctx context.Context, userID uint, gateway string) (*entities.PaymentMethodResponse, error) {
	paymentMethod, err := s.paymentMethodRepo.GetDefaultByUserIDAndGateway(ctx, userID, gateway)
	if err != nil {
		return nil, fmt.Errorf("failed to get default payment method by gateway: %w", err)
	}
	if paymentMethod == nil {
		return nil, fmt.Errorf("no default payment method found for gateway %s", gateway)
	}

	return paymentMethod.ToResponse(), nil
}

// ValidatePaymentMethod validates a payment method with the gateway
func (s *paymentMethodService) ValidatePaymentMethod(ctx context.Context, userID, paymentMethodID uint) (*entities.PaymentMethodResponse, error) {
	// Validate ownership
	isOwner, err := s.paymentMethodRepo.ValidateOwnership(ctx, paymentMethodID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ownership: %w", err)
	}
	if !isOwner {
		return nil, fmt.Errorf("payment method not found or access denied")
	}

	paymentMethod, err := s.paymentMethodRepo.GetByID(ctx, paymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}
	if paymentMethod == nil {
		return nil, fmt.Errorf("payment method not found")
	}

	// Validate with gateway (mock implementation)
	if err := s.validatePaymentTokenWithGateway(paymentMethod.Gateway, paymentMethod.PaymentToken); err != nil {
		// Mark as invalid if validation fails
		paymentMethod.Status = entities.PaymentMethodStatusInvalid
		if updateErr := s.paymentMethodRepo.Update(ctx, paymentMethod); updateErr != nil {
			s.logger.Error("Failed to update payment method status", logger.ErrorField(updateErr))
		}
		return nil, fmt.Errorf("payment method validation failed: %w", err)
	}

	// Update validation timestamp
	now := time.Now()
	paymentMethod.LastValidatedAt = &now
	if err := s.paymentMethodRepo.Update(ctx, paymentMethod); err != nil {
		s.logger.Error("Failed to update validation timestamp", logger.ErrorField(err))
	}

	s.logger.Info("Payment method validated", logger.Uint("payment_method_id", paymentMethodID), logger.Uint("user_id", userID))
	return paymentMethod.ToResponse(), nil
}

// ProcessPaymentWithMethod processes a payment using a specific payment method
func (s *paymentMethodService) ProcessPaymentWithMethod(ctx context.Context, userID, paymentMethodID uint, amount float64, currency string) (*entities.PaymentRecord, error) {
	// Validate ownership
	isOwner, err := s.paymentMethodRepo.ValidateOwnership(ctx, paymentMethodID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ownership: %w", err)
	}
	if !isOwner {
		return nil, fmt.Errorf("payment method not found or access denied")
	}

	paymentMethod, err := s.paymentMethodRepo.GetByID(ctx, paymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}
	if paymentMethod == nil {
		return nil, fmt.Errorf("payment method not found")
	}

	// Check if payment method can be used
	if !paymentMethod.CanBeUsedForPayment() {
		return nil, fmt.Errorf("payment method cannot be used for payment")
	}

	// TODO: Implement actual payment processing logic
	// This would integrate with the existing payment service
	s.logger.Info("Processing payment with method", logger.Uint("payment_method_id", paymentMethodID), logger.Float64("amount", amount), logger.String("currency", currency))

	// Update usage statistics
	if err := s.paymentMethodRepo.UpdateLastUsed(ctx, paymentMethodID, true); err != nil {
		s.logger.Error("Failed to update payment method usage", logger.ErrorField(err))
	}

	// Return mock payment record for now
	return &entities.PaymentRecord{
		UserID:        userID,
		PaymentMethod: paymentMethod.Method,
		Gateway:       paymentMethod.Gateway,
		Amount:        amount,
		Currency:      currency,
		Status:        entities.PaymentRecordStatusPending,
	}, nil
}

// ProcessPaymentWithDefaultMethod processes a payment using the default payment method
func (s *paymentMethodService) ProcessPaymentWithDefaultMethod(ctx context.Context, userID uint, gateway string, amount float64, currency string) (*entities.PaymentRecord, error) {
	defaultMethod, err := s.paymentMethodRepo.GetDefaultByUserIDAndGateway(ctx, userID, gateway)
	if err != nil {
		return nil, fmt.Errorf("failed to get default payment method: %w", err)
	}
	if defaultMethod == nil {
		return nil, fmt.Errorf("no default payment method found for gateway %s", gateway)
	}

	return s.ProcessPaymentWithMethod(ctx, userID, defaultMethod.ID, amount, currency)
}

// RetryFailedPayment retries a failed payment with alternative payment methods
func (s *paymentMethodService) RetryFailedPayment(ctx context.Context, userID uint, paymentRecordID uint) (*entities.PaymentRecord, error) {
	// TODO: Implement retry logic with alternative payment methods
	// This would:
	// 1. Get the failed payment record
	// 2. Find alternative payment methods for the user
	// 3. Try payment with methods with good success rates
	// 4. Update payment record with new attempt

	s.logger.Info("Retrying failed payment", logger.Uint("user_id", userID), logger.Uint("payment_record_id", paymentRecordID))
	return nil, fmt.Errorf("retry failed payment not implemented yet")
}

// GetPaymentMethodUsageStats retrieves usage statistics for a payment method
func (s *paymentMethodService) GetPaymentMethodUsageStats(ctx context.Context, userID, paymentMethodID uint) (*interfaces.PaymentMethodUsageStats, error) {
	// Validate ownership
	isOwner, err := s.paymentMethodRepo.ValidateOwnership(ctx, paymentMethodID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ownership: %w", err)
	}
	if !isOwner {
		return nil, fmt.Errorf("payment method not found or access denied")
	}

	paymentMethod, err := s.paymentMethodRepo.GetByID(ctx, paymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}
	if paymentMethod == nil {
		return nil, fmt.Errorf("payment method not found")
	}

	totalUses := paymentMethod.SuccessfulUses + paymentMethod.FailedUses
	var successRate, failureRate float64
	if totalUses > 0 {
		successRate = float64(paymentMethod.SuccessfulUses) / float64(totalUses)
		failureRate = float64(paymentMethod.FailedUses) / float64(totalUses)
	}

	var lastUsed *string
	if paymentMethod.LastUsedAt != nil {
		lastUsedStr := paymentMethod.LastUsedAt.Format(time.RFC3339)
		lastUsed = &lastUsedStr
	}

	return &interfaces.PaymentMethodUsageStats{
		PaymentMethodID:  paymentMethodID,
		TotalUses:        totalUses,
		SuccessfulUses:   paymentMethod.SuccessfulUses,
		FailedUses:       paymentMethod.FailedUses,
		SuccessRate:      successRate,
		FailureRate:      failureRate,
		LastUsed:         lastUsed,
		AverageAmount:    0, // TODO: Calculate from payment records
		TotalAmount:      0, // TODO: Calculate from payment records
		RecentUses30Days: 0, // TODO: Calculate from payment records
	}, nil
}

// RefreshExpiredMethods marks expired payment methods as expired
func (s *paymentMethodService) RefreshExpiredMethods(ctx context.Context) error {
	expiredMethods, err := s.paymentMethodRepo.GetExpiredMethods(ctx)
	if err != nil {
		return fmt.Errorf("failed to get expired methods: %w", err)
	}

	for _, method := range expiredMethods {
		if err := s.paymentMethodRepo.UpdateStatus(ctx, method.ID, entities.PaymentMethodStatusExpired); err != nil {
			s.logger.Error("Failed to update expired method status", logger.ErrorField(err), logger.Uint("payment_method_id", method.ID))
		}
	}

	s.logger.Info("Refreshed expired payment methods", logger.Int("count", len(expiredMethods)))
	return nil
}

// RevalidatePaymentMethods revalidates payment methods that need validation
func (s *paymentMethodService) RevalidatePaymentMethods(ctx context.Context) error {
	methods, err := s.paymentMethodRepo.GetMethodsNeedingValidation(ctx)
	if err != nil {
		return fmt.Errorf("failed to get methods needing validation: %w", err)
	}

	revalidated := 0
	for _, method := range methods {
		if err := s.validatePaymentTokenWithGateway(method.Gateway, method.PaymentToken); err != nil {
			// Mark as invalid
			if updateErr := s.paymentMethodRepo.UpdateStatus(ctx, method.ID, entities.PaymentMethodStatusInvalid); updateErr != nil {
				s.logger.Error("Failed to mark method as invalid", logger.ErrorField(updateErr), logger.Uint("payment_method_id", method.ID))
			}
		} else {
			// Update validation timestamp
			now := time.Now()
			method.LastValidatedAt = &now
			if updateErr := s.paymentMethodRepo.Update(ctx, method); updateErr != nil {
				s.logger.Error("Failed to update validation timestamp", logger.ErrorField(updateErr), logger.Uint("payment_method_id", method.ID))
			} else {
				revalidated++
			}
		}
	}

	s.logger.Info("Revalidated payment methods", logger.Int("total", len(methods)), logger.Int("successful", revalidated))
	return nil
}

// GetUserPaymentMethodsCount returns the count of payment methods for a user
func (s *paymentMethodService) GetUserPaymentMethodsCount(ctx context.Context, userID uint) (int64, error) {
	return s.paymentMethodRepo.GetUserPaymentMethodCount(ctx, userID)
}

// IsPaymentMethodLimitReached checks if user has reached payment method limit
func (s *paymentMethodService) IsPaymentMethodLimitReached(ctx context.Context, userID uint) (bool, error) {
	count, err := s.GetUserPaymentMethodsCount(ctx, userID)
	if err != nil {
		return false, err
	}
	return count >= MaxPaymentMethodsPerUser, nil
}

// GetRecommendedPaymentMethod returns the best payment method for a user based on history
func (s *paymentMethodService) GetRecommendedPaymentMethod(ctx context.Context, userID uint, gateway string) (*entities.PaymentMethodResponse, error) {
	// First try to get default method for the gateway
	defaultMethod, err := s.paymentMethodRepo.GetDefaultByUserIDAndGateway(ctx, userID, gateway)
	if err == nil && defaultMethod != nil && defaultMethod.CanBeUsedForPayment() {
		return defaultMethod.ToResponse(), nil
	}

	// Get all active methods for the gateway
	methods, err := s.paymentMethodRepo.GetByUserIDAndGateway(ctx, userID, gateway)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment methods: %w", err)
	}

	// Find the method with the best success rate
	var bestMethod *entities.PaymentMethod
	var bestSuccessRate float64 = -1

	for _, method := range methods {
		if !method.CanBeUsedForPayment() {
			continue
		}

		successRate := 1.0 // Default for new methods
		totalUses := method.SuccessfulUses + method.FailedUses
		if totalUses > 0 {
			successRate = float64(method.SuccessfulUses) / float64(totalUses)
		}

		if successRate > bestSuccessRate {
			bestSuccessRate = successRate
			bestMethod = method
		}
	}

	if bestMethod == nil {
		return nil, fmt.Errorf("no usable payment method found for gateway %s", gateway)
	}

	return bestMethod.ToResponse(), nil
}

// Helper functions

// validatePaymentTokenWithGateway validates a payment token with the specified gateway
// This is a mock implementation - replace with actual gateway integration
func (s *paymentMethodService) validatePaymentTokenWithGateway(gateway, token string) error {
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("empty payment token")
	}

	// Mock validation logic
	switch gateway {
	case entities.PaymentGatewayEpay:
		if !strings.HasPrefix(token, "epay_") {
			return fmt.Errorf("invalid epay token format")
		}
	case entities.PaymentGatewayEPUSDT:
		if !strings.HasPrefix(token, "epusdt_") {
			return fmt.Errorf("invalid epusdt token format")
		}
	default:
		return fmt.Errorf("unsupported gateway: %s", gateway)
	}

	return nil
}

// timePtr returns a pointer to a time.Time value
func timePtr(t time.Time) *time.Time {
	return &t
}
