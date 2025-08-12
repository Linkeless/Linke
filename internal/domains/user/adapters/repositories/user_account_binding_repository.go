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
			logger.ErrorField(err))
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
			logger.ErrorField(err))
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
			logger.ErrorField(err))
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
			logger.ErrorField(err))
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
			logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get user account bindings: %w", err)
	}

	return bindings, nil
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
			logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get user account binding: %w", err)
	}

	return &binding, nil
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

// UpdateLastUsed updates the last used timestamp
func (r *userAccountBindingRepository) UpdateLastUsed(ctx context.Context, id uint) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&entities.UserAccountBinding{}).
		Where("id = ?", id).
		Update("last_used_at", now).Error; err != nil {
		logger.Error("Failed to update last used timestamp",
			logger.Uint("binding_id", id),
			logger.ErrorField(err))
		return fmt.Errorf("failed to update last used: %w", err)
	}

	return nil
}

