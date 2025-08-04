package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
)

// paymentMethodRepository implements the PaymentMethodRepository interface
type paymentMethodRepository struct {
	db *gorm.DB
}

// NewPaymentMethodRepository creates a new payment method repository instance
func NewPaymentMethodRepository(db *gorm.DB) interfaces.PaymentMethodRepository {
	return &paymentMethodRepository{
		db: db,
	}
}

// Create creates a new payment method for a user
func (r *paymentMethodRepository) Create(ctx context.Context, paymentMethod *entities.PaymentMethod) error {
	if err := paymentMethod.GenerateValidationHash(); err != nil {
		return fmt.Errorf("failed to generate validation hash: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(paymentMethod).Error; err != nil {
		return fmt.Errorf("failed to create payment method: %w", err)
	}

	return nil
}

// GetByID retrieves a payment method by its ID
func (r *paymentMethodRepository) GetByID(ctx context.Context, id uint) (*entities.PaymentMethod, error) {
	var paymentMethod entities.PaymentMethod
	if err := r.db.WithContext(ctx).First(&paymentMethod, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment method by ID: %w", err)
	}

	return &paymentMethod, nil
}

// GetByUserID retrieves all payment methods for a user
func (r *paymentMethodRepository) GetByUserID(ctx context.Context, userID uint) ([]*entities.PaymentMethod, error) {
	var paymentMethods []*entities.PaymentMethod
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&paymentMethods).Error; err != nil {
		return nil, fmt.Errorf("failed to get payment methods by user ID: %w", err)
	}

	return paymentMethods, nil
}

// GetActiveByUserID retrieves all active payment methods for a user
func (r *paymentMethodRepository) GetActiveByUserID(ctx context.Context, userID uint) ([]*entities.PaymentMethod, error) {
	var paymentMethods []*entities.PaymentMethod
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ? AND status = ?", userID, true, entities.PaymentMethodStatusActive).
		Order("is_default DESC, created_at DESC").
		Find(&paymentMethods).Error; err != nil {
		return nil, fmt.Errorf("failed to get active payment methods by user ID: %w", err)
	}

	return paymentMethods, nil
}

// GetByUserIDAndGateway retrieves payment methods for a user by gateway
func (r *paymentMethodRepository) GetByUserIDAndGateway(ctx context.Context, userID uint, gateway string) ([]*entities.PaymentMethod, error) {
	var paymentMethods []*entities.PaymentMethod
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND gateway = ?", userID, gateway).
		Order("is_default DESC, created_at DESC").
		Find(&paymentMethods).Error; err != nil {
		return nil, fmt.Errorf("failed to get payment methods by user ID and gateway: %w", err)
	}

	return paymentMethods, nil
}

// GetDefaultByUserID retrieves the default payment method for a user
func (r *paymentMethodRepository) GetDefaultByUserID(ctx context.Context, userID uint) (*entities.PaymentMethod, error) {
	var paymentMethod entities.PaymentMethod
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_default = ? AND is_active = ? AND status = ?",
			userID, true, true, entities.PaymentMethodStatusActive).
		First(&paymentMethod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get default payment method: %w", err)
	}

	return &paymentMethod, nil
}

// GetDefaultByUserIDAndGateway retrieves the default payment method for a user and gateway
func (r *paymentMethodRepository) GetDefaultByUserIDAndGateway(ctx context.Context, userID uint, gateway string) (*entities.PaymentMethod, error) {
	var paymentMethod entities.PaymentMethod
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND gateway = ? AND is_default = ? AND is_active = ? AND status = ?",
			userID, gateway, true, true, entities.PaymentMethodStatusActive).
		First(&paymentMethod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get default payment method by gateway: %w", err)
	}

	return &paymentMethod, nil
}

// GetByPaymentToken retrieves a payment method by its payment token and gateway
func (r *paymentMethodRepository) GetByPaymentToken(ctx context.Context, gateway, token string) (*entities.PaymentMethod, error) {
	var paymentMethod entities.PaymentMethod
	if err := r.db.WithContext(ctx).
		Where("gateway = ? AND payment_token = ? AND status = ?",
			gateway, token, entities.PaymentMethodStatusActive).
		First(&paymentMethod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment method by token: %w", err)
	}

	return &paymentMethod, nil
}

// Update updates an existing payment method
func (r *paymentMethodRepository) Update(ctx context.Context, paymentMethod *entities.PaymentMethod) error {
	if err := r.db.WithContext(ctx).Save(paymentMethod).Error; err != nil {
		return fmt.Errorf("failed to update payment method: %w", err)
	}

	return nil
}

// UpdateStatus updates the status of a payment method
func (r *paymentMethodRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	if err := r.db.WithContext(ctx).
		Model(&entities.PaymentMethod{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update payment method status: %w", err)
	}

	return nil
}

// SetAsDefault sets a payment method as the default for the user and gateway
func (r *paymentMethodRepository) SetAsDefault(ctx context.Context, userID, paymentMethodID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First, get the payment method to find its gateway
		var paymentMethod entities.PaymentMethod
		if err := tx.First(&paymentMethod, paymentMethodID).Error; err != nil {
			return fmt.Errorf("failed to find payment method: %w", err)
		}

		// Verify ownership
		if paymentMethod.UserID != userID {
			return fmt.Errorf("payment method does not belong to user")
		}

		// Unset default for all other methods of this user and gateway
		if err := tx.Model(&entities.PaymentMethod{}).
			Where("user_id = ? AND gateway = ? AND id != ?", userID, paymentMethod.Gateway, paymentMethodID).
			Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to unset other default methods: %w", err)
		}

		// Set this method as default
		if err := tx.Model(&entities.PaymentMethod{}).
			Where("id = ?", paymentMethodID).
			Update("is_default", true).Error; err != nil {
			return fmt.Errorf("failed to set method as default: %w", err)
		}

		return nil
	})
}

// UnsetDefault removes default status from a payment method
func (r *paymentMethodRepository) UnsetDefault(ctx context.Context, userID uint, gateway string) error {
	if err := r.db.WithContext(ctx).
		Model(&entities.PaymentMethod{}).
		Where("user_id = ? AND gateway = ?", userID, gateway).
		Update("is_default", false).Error; err != nil {
		return fmt.Errorf("failed to unset default payment method: %w", err)
	}

	return nil
}

// Delete soft deletes a payment method
func (r *paymentMethodRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entities.PaymentMethod{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete payment method: %w", err)
	}

	return nil
}

// HardDelete permanently deletes a payment method
func (r *paymentMethodRepository) HardDelete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Unscoped().Delete(&entities.PaymentMethod{}, id).Error; err != nil {
		return fmt.Errorf("failed to hard delete payment method: %w", err)
	}

	return nil
}

// GetExpiredMethods retrieves payment methods that have expired
func (r *paymentMethodRepository) GetExpiredMethods(ctx context.Context) ([]*entities.PaymentMethod, error) {
	var paymentMethods []*entities.PaymentMethod
	now := time.Now()

	if err := r.db.WithContext(ctx).
		Where("expiry_year IS NOT NULL AND expiry_month IS NOT NULL").
		Where("(expiry_year < ? OR (expiry_year = ? AND expiry_month < ?))",
			now.Year(), now.Year(), int(now.Month())).
		Where("status != ?", entities.PaymentMethodStatusExpired).
		Find(&paymentMethods).Error; err != nil {
		return nil, fmt.Errorf("failed to get expired payment methods: %w", err)
	}

	return paymentMethods, nil
}

// GetMethodsNeedingValidation retrieves payment methods that need revalidation
func (r *paymentMethodRepository) GetMethodsNeedingValidation(ctx context.Context) ([]*entities.PaymentMethod, error) {
	var paymentMethods []*entities.PaymentMethod
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	if err := r.db.WithContext(ctx).
		Where("(last_validated_at IS NULL OR last_validated_at < ?) AND status = ?",
			thirtyDaysAgo, entities.PaymentMethodStatusActive).
		Find(&paymentMethods).Error; err != nil {
		return nil, fmt.Errorf("failed to get payment methods needing validation: %w", err)
	}

	return paymentMethods, nil
}

// UpdateLastUsed updates the last used timestamp and usage statistics
func (r *paymentMethodRepository) UpdateLastUsed(ctx context.Context, id uint, successful bool) error {
	updates := map[string]interface{}{
		"last_used_at": time.Now(),
	}

	if successful {
		if err := r.db.WithContext(ctx).Model(&entities.PaymentMethod{}).
			Where("id = ?", id).
			Update("successful_uses", gorm.Expr("successful_uses + 1")).Error; err != nil {
			return fmt.Errorf("failed to increment successful uses: %w", err)
		}
	} else {
		if err := r.db.WithContext(ctx).Model(&entities.PaymentMethod{}).
			Where("id = ?", id).
			Update("failed_uses", gorm.Expr("failed_uses + 1")).Error; err != nil {
			return fmt.Errorf("failed to increment failed uses: %w", err)
		}
	}

	if err := r.db.WithContext(ctx).Model(&entities.PaymentMethod{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}

	return nil
}

// GetUserPaymentMethodCount returns the count of payment methods for a user
func (r *paymentMethodRepository) GetUserPaymentMethodCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entities.PaymentMethod{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count user payment methods: %w", err)
	}

	return count, nil
}

// GetHighFailureRateMethods returns payment methods with high failure rates
func (r *paymentMethodRepository) GetHighFailureRateMethods(ctx context.Context, threshold float64) ([]*entities.PaymentMethod, error) {
	var paymentMethods []*entities.PaymentMethod

	// Use raw SQL to calculate failure rate
	if err := r.db.WithContext(ctx).
		Where("(successful_uses + failed_uses) > 0").
		Where("CAST(failed_uses AS FLOAT) / CAST((successful_uses + failed_uses) AS FLOAT) >= ?", threshold).
		Find(&paymentMethods).Error; err != nil {
		return nil, fmt.Errorf("failed to get high failure rate methods: %w", err)
	}

	return paymentMethods, nil
}

// ValidateOwnership checks if a payment method belongs to a specific user
func (r *paymentMethodRepository) ValidateOwnership(ctx context.Context, paymentMethodID, userID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entities.PaymentMethod{}).
		Where("id = ? AND user_id = ?", paymentMethodID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to validate payment method ownership: %w", err)
	}

	return count > 0, nil
}
