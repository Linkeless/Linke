package repositories

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/user/entities"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type userAccountBindingRepository struct {
	db *gorm.DB
}

// NewUserAccountBindingRepository creates a new user account binding repository
func NewUserAccountBindingRepository(db *gorm.DB) interfaces.UserAccountBindingRepository {
	return &userAccountBindingRepository{
		db: db,
	}
}

// Create creates a new user account binding
func (r *userAccountBindingRepository) Create(ctx context.Context, binding *entities.UserAccountBinding) error {
	if err := r.db.WithContext(ctx).Create(binding).Error; err != nil {
		logger.Error("Failed to create user account binding",
			logger.Uint("user_id", binding.UserID),
			logger.String("provider", binding.Provider),
			logger.Error2("error", err))
		return fmt.Errorf("failed to create user account binding: %w", err)
	}

	logger.Info("User account binding created successfully",
		logger.Uint("binding_id", binding.ID),
		logger.Uint("user_id", binding.UserID),
		logger.String("provider", binding.Provider))

	return nil
}

// GetByID retrieves a user account binding by ID
func (r *userAccountBindingRepository) GetByID(ctx context.Context, id uint) (*entities.UserAccountBinding, error) {
	var binding entities.UserAccountBinding
	if err := r.db.WithContext(ctx).First(&binding, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user account binding with ID %d not found", id)
		}
		logger.Error("Failed to get user account binding by ID",
			logger.Uint("binding_id", id),
			logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get user account binding: %w", err)
	}

	return &binding, nil
}

// Update updates a user account binding
func (r *userAccountBindingRepository) Update(ctx context.Context, binding *entities.UserAccountBinding) error {
	if err := r.db.WithContext(ctx).Save(binding).Error; err != nil {
		logger.Error("Failed to update user account binding",
			logger.Uint("binding_id", binding.ID),
			logger.Uint("user_id", binding.UserID),
			logger.Error2("error", err))
		return fmt.Errorf("failed to update user account binding: %w", err)
	}

	logger.Info("User account binding updated successfully",
		logger.Uint("binding_id", binding.ID),
		logger.Uint("user_id", binding.UserID),
		logger.String("provider", binding.Provider))

	return nil
}

// Delete deletes a user account binding (soft delete)
func (r *userAccountBindingRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entities.UserAccountBinding{}, id).Error; err != nil {
		logger.Error("Failed to delete user account binding",
			logger.Uint("binding_id", id),
			logger.Error2("error", err))
		return fmt.Errorf("failed to delete user account binding: %w", err)
	}

	logger.Info("User account binding deleted successfully",
		logger.Uint("binding_id", id))

	return nil
}

// GetByUserID retrieves all bindings for a user
func (r *userAccountBindingRepository) GetByUserID(ctx context.Context, userID uint) ([]*entities.UserAccountBinding, error) {
	var bindings []*entities.UserAccountBinding
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_primary DESC, created_at ASC").
		Find(&bindings).Error; err != nil {
		logger.Error("Failed to get user account bindings by user ID",
			logger.Uint("user_id", userID),
			logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get user account bindings: %w", err)
	}

	return bindings, nil
}

// GetByUserIDAndProvider retrieves a binding by user ID and provider
func (r *userAccountBindingRepository) GetByUserIDAndProvider(ctx context.Context, userID uint, provider string) (*entities.UserAccountBinding, error) {
	var binding entities.UserAccountBinding
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		First(&binding).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user account binding not found for user %d and provider %s", userID, provider)
		}
		logger.Error("Failed to get user account binding by user ID and provider",
			logger.Uint("user_id", userID),
			logger.String("provider", provider),
			logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get user account binding: %w", err)
	}

	return &binding, nil
}

// GetByProviderAndProviderUserID retrieves a binding by provider and provider user ID
func (r *userAccountBindingRepository) GetByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (*entities.UserAccountBinding, error) {
	var binding entities.UserAccountBinding
	if err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_user_id = ?", provider, providerUserID).
		First(&binding).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user account binding not found for provider %s and provider user ID %s", provider, providerUserID)
		}
		logger.Error("Failed to get user account binding by provider and provider user ID",
			logger.String("provider", provider),
			logger.String("provider_user_id", providerUserID),
			logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get user account binding: %w", err)
	}

	return &binding, nil
}

// GetByProviderAndEmail retrieves a binding by provider and email
func (r *userAccountBindingRepository) GetByProviderAndEmail(ctx context.Context, provider, email string) (*entities.UserAccountBinding, error) {
	var binding entities.UserAccountBinding
	if err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_email = ?", provider, email).
		First(&binding).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user account binding not found for provider %s and email %s", provider, email)
		}
		logger.Error("Failed to get user account binding by provider and email",
			logger.String("provider", provider),
			logger.String("provider_email", email),
			logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get user account binding: %w", err)
	}

	return &binding, nil
}

// GetPrimaryBindingByUserID retrieves the primary binding for a user
func (r *userAccountBindingRepository) GetPrimaryBindingByUserID(ctx context.Context, userID uint) (*entities.UserAccountBinding, error) {
	var binding entities.UserAccountBinding
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_primary = ?", userID, true).
		First(&binding).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no primary binding found for user %d", userID)
		}
		logger.Error("Failed to get primary binding by user ID",
			logger.Uint("user_id", userID),
			logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get primary binding: %w", err)
	}

	return &binding, nil
}

// ListByUserID retrieves bindings for a user with pagination
func (r *userAccountBindingRepository) ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]*entities.UserAccountBinding, int64, error) {
	var bindings []*entities.UserAccountBinding
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).
		Model(&entities.UserAccountBinding{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		logger.Error("Failed to count user account bindings",
			logger.Uint("user_id", userID),
			logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count bindings: %w", err)
	}

	// Get bindings with pagination
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_primary DESC, created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&bindings).Error; err != nil {
		logger.Error("Failed to list user account bindings",
			logger.Uint("user_id", userID),
			logger.Int("offset", offset),
			logger.Int("limit", limit),
			logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list bindings: %w", err)
	}

	return bindings, total, nil
}

// ExistsByProviderAndProviderUserID checks if a binding exists
func (r *userAccountBindingRepository) ExistsByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entities.UserAccountBinding{}).
		Where("provider = ? AND provider_user_id = ?", provider, providerUserID).
		Count(&count).Error; err != nil {
		logger.Error("Failed to check binding existence",
			logger.String("provider", provider),
			logger.String("provider_user_id", providerUserID),
			logger.Error2("error", err))
		return false, fmt.Errorf("failed to check binding existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByUserIDAndProvider checks if a binding exists for user and provider
func (r *userAccountBindingRepository) ExistsByUserIDAndProvider(ctx context.Context, userID uint, provider string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entities.UserAccountBinding{}).
		Where("user_id = ? AND provider = ?", userID, provider).
		Count(&count).Error; err != nil {
		logger.Error("Failed to check binding existence",
			logger.Uint("user_id", userID),
			logger.String("provider", provider),
			logger.Error2("error", err))
		return false, fmt.Errorf("failed to check binding existence: %w", err)
	}

	return count > 0, nil
}

// DeleteByUserID deletes all bindings for a user
func (r *userAccountBindingRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entities.UserAccountBinding{}).Error; err != nil {
		logger.Error("Failed to delete bindings by user ID",
			logger.Uint("user_id", userID),
			logger.Error2("error", err))
		return fmt.Errorf("failed to delete bindings: %w", err)
	}

	logger.Info("User account bindings deleted successfully",
		logger.Uint("user_id", userID))

	return nil
}

// CountByUserID counts bindings for a user
func (r *userAccountBindingRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entities.UserAccountBinding{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		logger.Error("Failed to count bindings by user ID",
			logger.Uint("user_id", userID),
			logger.Error2("error", err))
		return 0, fmt.Errorf("failed to count bindings: %w", err)
	}

	return count, nil
}

// CountByProvider counts bindings for a provider
func (r *userAccountBindingRepository) CountByProvider(ctx context.Context, provider string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entities.UserAccountBinding{}).
		Where("provider = ?", provider).
		Count(&count).Error; err != nil {
		logger.Error("Failed to count bindings by provider",
			logger.String("provider", provider),
			logger.Error2("error", err))
		return 0, fmt.Errorf("failed to count bindings: %w", err)
	}

	return count, nil
}

// SetPrimaryBinding sets a binding as primary and unsets others
func (r *userAccountBindingRepository) SetPrimaryBinding(ctx context.Context, userID uint, bindingID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First, unset all other bindings for this user
		if err := tx.Model(&entities.UserAccountBinding{}).
			Where("user_id = ? AND id != ?", userID, bindingID).
			Update("is_primary", false).Error; err != nil {
			return fmt.Errorf("failed to unset other primary bindings: %w", err)
		}

		// Then set this one as primary
		if err := tx.Model(&entities.UserAccountBinding{}).
			Where("id = ? AND user_id = ?", bindingID, userID).
			Update("is_primary", true).Error; err != nil {
			return fmt.Errorf("failed to set primary binding: %w", err)
		}

		logger.Info("Primary binding updated successfully",
			logger.Uint("user_id", userID),
			logger.Uint("binding_id", bindingID))

		return nil
	})
}

// UnsetAllPrimaryBindings unsets all primary bindings for a user
func (r *userAccountBindingRepository) UnsetAllPrimaryBindings(ctx context.Context, userID uint) error {
	if err := r.db.WithContext(ctx).
		Model(&entities.UserAccountBinding{}).
		Where("user_id = ?", userID).
		Update("is_primary", false).Error; err != nil {
		logger.Error("Failed to unset all primary bindings",
			logger.Uint("user_id", userID),
			logger.Error2("error", err))
		return fmt.Errorf("failed to unset primary bindings: %w", err)
	}

	logger.Info("All primary bindings unset successfully",
		logger.Uint("user_id", userID))

	return nil
}

// UpdateLastUsed updates the last used timestamp
func (r *userAccountBindingRepository) UpdateLastUsed(ctx context.Context, id uint) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&entities.UserAccountBinding{}).
		Where("id = ?", id).
		Update("last_used_at", now).Error; err != nil {
		logger.Error("Failed to update last used timestamp",
			logger.Uint("binding_id", id),
			logger.Error2("error", err))
		return fmt.Errorf("failed to update last used: %w", err)
	}

	return nil
}

// CleanupOldBindings removes bindings that haven't been used for specified days
func (r *userAccountBindingRepository) CleanupOldBindings(ctx context.Context, days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)

	result := r.db.WithContext(ctx).
		Where("last_used_at < ? OR (last_used_at IS NULL AND created_at < ?)", cutoffDate, cutoffDate).
		Delete(&entities.UserAccountBinding{})

	if result.Error != nil {
		logger.Error("Failed to cleanup old bindings",
			logger.Int("days", days),
			logger.Error2("error", result.Error))
		return fmt.Errorf("failed to cleanup old bindings: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		logger.Info("Old bindings cleaned up successfully",
			logger.Int("days", days),
			logger.Int64("rows_affected", result.RowsAffected))
	}

	return nil
}