package implementations

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/user/entities"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/logger"
)

type userAccountBindingService struct {
	bindingRepo interfaces.UserAccountBindingRepository
	userRepo    interfaces.UserRepository
}

// NewUserAccountBindingService creates a new user account binding service
func NewUserAccountBindingService(
	bindingRepo interfaces.UserAccountBindingRepository,
	userRepo interfaces.UserRepository,
) interfaces.UserAccountBindingService {
	return &userAccountBindingService{
		bindingRepo: bindingRepo,
		userRepo:    userRepo,
	}
}

// CreateBinding creates a new account binding for a user
func (s *userAccountBindingService) CreateBinding(ctx context.Context, userID uint, req *entities.CreateBindingRequest) (*entities.UserAccountBinding, error) {
	// Validate request
	if err := s.validateBindingRequest(req); err != nil {
		return nil, fmt.Errorf("invalid binding request: %w", err)
	}

	// Check if user can bind this provider
	if err := s.canBindProvider(ctx, userID, req.Provider); err != nil {
		return nil, err
	}

	// Check if provider account is already bound to another user
	exists, err := s.isProviderAccountBound(ctx, req.Provider, req.ProviderUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check provider account binding: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("provider account %s:%s is already bound to another user", req.Provider, req.ProviderUserID)
	}

	// Create binding
	binding := &entities.UserAccountBinding{
		UserID:           userID,
		Provider:         req.Provider,
		ProviderUserID:   req.ProviderUserID,
		ProviderEmail:    req.ProviderEmail,
		ProviderUsername: req.ProviderUsername,
		ProviderName:     req.ProviderName,
		ProviderAvatar:   req.ProviderAvatar,
		ProviderData:     req.ProviderData,
		BoundAt:          time.Now(),
	}

	// Set as primary if requested or if it's the first binding
	if req.IsPrimary != nil && *req.IsPrimary {
		binding.IsPrimary = true
	} else {
		// Check if user has any existing bindings
		existingBindings, err := s.bindingRepo.GetByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing bindings: %w", err)
		}
		// Set as primary if it's the first binding
		binding.IsPrimary = len(existingBindings) == 0
	}

	if err := s.bindingRepo.Create(ctx, binding); err != nil {
		return nil, fmt.Errorf("failed to create binding: %w", err)
	}

	// If this binding was set as primary, ensure it's the only primary binding
	if binding.IsPrimary {
		if err := s.bindingRepo.SetPrimaryBinding(ctx, userID, binding.ID); err != nil {
			logger.Error("Failed to set primary binding",
				logger.Uint("user_id", userID),
				logger.Uint("binding_id", binding.ID),
				logger.ErrorField(err))
			// Don't fail the operation, just log the error
		}
	}

	logger.Info("Account binding created successfully",
		logger.Uint("user_id", userID),
		logger.String("provider", req.Provider),
		logger.Bool("is_primary", binding.IsPrimary))

	return binding, nil
}

// GetUserBindings retrieves all bindings for a user
func (s *userAccountBindingService) GetUserBindings(ctx context.Context, userID uint) ([]*entities.UserAccountBinding, error) {
	bindings, err := s.bindingRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bindings: %w", err)
	}

	return bindings, nil
}

// UpdateBinding updates an existing binding
func (s *userAccountBindingService) UpdateBinding(ctx context.Context, userID uint, provider string, req *entities.UpdateBindingRequest) (*entities.UserAccountBinding, error) {
	// Get existing binding
	binding, err := s.getUserBinding(ctx, userID, provider)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.ProviderEmail != nil {
		binding.ProviderEmail = req.ProviderEmail
	}
	if req.ProviderUsername != nil {
		binding.ProviderUsername = req.ProviderUsername
	}
	if req.ProviderName != nil {
		binding.ProviderName = req.ProviderName
	}
	if req.ProviderAvatar != nil {
		binding.ProviderAvatar = req.ProviderAvatar
	}
	if req.ProviderData != nil {
		binding.ProviderData = req.ProviderData
	}

	// Handle primary status change
	if req.IsPrimary != nil && *req.IsPrimary != binding.IsPrimary {
		if *req.IsPrimary {
			// Set as primary
			if err := s.bindingRepo.SetPrimaryBinding(ctx, userID, binding.ID); err != nil {
				return nil, fmt.Errorf("failed to set primary binding: %w", err)
			}
			binding.IsPrimary = true
		} else {
			// Unset as primary (but ensure at least one primary exists)
			bindings, err := s.bindingRepo.GetByUserID(ctx, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to check existing bindings: %w", err)
			}

			primaryCount := 0
			for _, b := range bindings {
				if b.IsPrimary {
					primaryCount++
				}
			}

			if primaryCount <= 1 {
				return nil, fmt.Errorf("cannot unset primary binding when it's the only primary binding")
			}

			binding.IsPrimary = false
		}
	}

	if err := s.bindingRepo.Update(ctx, binding); err != nil {
		return nil, fmt.Errorf("failed to update binding: %w", err)
	}

	logger.Info("Account binding updated successfully",
		logger.Uint("user_id", userID),
		logger.String("provider", provider))

	return binding, nil
}

// DeleteBinding deletes a binding
func (s *userAccountBindingService) DeleteBinding(ctx context.Context, userID uint, provider string) error {
	// Get existing binding
	binding, err := s.getUserBinding(ctx, userID, provider)
	if err != nil {
		return err
	}

	// Check if this is the primary binding and if there are other bindings
	if binding.IsPrimary {
		bindings, err := s.bindingRepo.GetByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to check existing bindings: %w", err)
		}

		// If there are other bindings, set one of them as primary
		if len(bindings) > 1 {
			for _, b := range bindings {
				if b.ID != binding.ID {
					if err := s.bindingRepo.SetPrimaryBinding(ctx, userID, b.ID); err != nil {
						logger.Error("Failed to set new primary binding during deletion",
							logger.Uint("user_id", userID),
							logger.Uint("new_primary_id", b.ID),
							logger.ErrorField(err))
					}
					break
				}
			}
		}
	}

	if err := s.bindingRepo.Delete(ctx, binding.ID); err != nil {
		return fmt.Errorf("failed to delete binding: %w", err)
	}

	logger.Info("Account binding deleted successfully",
		logger.Uint("user_id", userID),
		logger.String("provider", provider))

	return nil
}

// FindUserByProviderAccount finds a user by provider account
func (s *userAccountBindingService) FindUserByProviderAccount(ctx context.Context, provider, providerUserID string) (*entities.UserAccountBinding, error) {
	if !entities.IsValidProvider(provider) {
		return nil, fmt.Errorf("invalid provider: %s", provider)
	}

	binding, err := s.bindingRepo.GetByProviderAndProviderUserID(ctx, provider, providerUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by provider account: %w", err)
	}

	// Update last used timestamp
	if err := s.bindingRepo.UpdateLastUsed(ctx, binding.ID); err != nil {
		logger.Error("Failed to update last used timestamp",
			logger.Uint("binding_id", binding.ID),
			logger.ErrorField(err))
		// Don't fail the operation
	}

	return binding, nil
}

// SetPrimaryBinding sets a binding as primary
func (s *userAccountBindingService) SetPrimaryBinding(ctx context.Context, userID uint, provider string) error {
	if !entities.IsValidProvider(provider) {
		return fmt.Errorf("invalid provider: %s", provider)
	}

	// Get the binding
	binding, err := s.getUserBinding(ctx, userID, provider)
	if err != nil {
		return fmt.Errorf("failed to get binding: %w", err)
	}

	// Set as primary
	if err := s.bindingRepo.SetPrimaryBinding(ctx, userID, binding.ID); err != nil {
		return fmt.Errorf("failed to set primary binding: %w", err)
	}

	logger.Info("Primary binding set successfully",
		logger.Uint("user_id", userID),
		logger.String("provider", provider))

	return nil
}

// Helper methods - not exposed in interface

// getUserBinding retrieves a specific binding for a user and provider
func (s *userAccountBindingService) getUserBinding(ctx context.Context, userID uint, provider string) (*entities.UserAccountBinding, error) {
	if !entities.IsValidProvider(provider) {
		return nil, fmt.Errorf("invalid provider: %s", provider)
	}

	// Since we removed GetByUserIDAndProvider from repository, we need to get all user bindings
	// and find the one with matching provider
	bindings, err := s.bindingRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bindings: %w", err)
	}

	for _, binding := range bindings {
		if binding.Provider == provider {
			return binding, nil
		}
	}

	return nil, fmt.Errorf("user account binding not found for user %d and provider %s", userID, provider)
}

// validateBindingRequest validates a binding request
func (s *userAccountBindingService) validateBindingRequest(req *entities.CreateBindingRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	if !entities.IsValidProvider(req.Provider) {
		return fmt.Errorf("invalid provider: %s", req.Provider)
	}

	if req.ProviderUserID == "" {
		return fmt.Errorf("provider user ID is required")
	}

	return nil
}

// canBindProvider checks if a user can bind a provider
func (s *userAccountBindingService) canBindProvider(ctx context.Context, userID uint, provider string) error {
	if !entities.IsValidProvider(provider) {
		return fmt.Errorf("invalid provider: %s", provider)
	}

	// Check if user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if user already has a binding for this provider
	bindings, err := s.bindingRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check existing bindings: %w", err)
	}

	for _, binding := range bindings {
		if binding.Provider == provider {
			return fmt.Errorf("user already has a binding for provider: %s", provider)
		}
	}

	return nil
}

// isProviderAccountBound checks if a provider account is already bound
func (s *userAccountBindingService) isProviderAccountBound(ctx context.Context, provider, providerUserID string) (bool, error) {
	_, err := s.bindingRepo.GetByProviderAndProviderUserID(ctx, provider, providerUserID)
	if err != nil {
		// If error is "not found", return false; otherwise return the error
		if err.Error() == fmt.Sprintf("user account binding not found for provider %s and provider user ID %s", provider, providerUserID) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
