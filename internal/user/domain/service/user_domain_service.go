package service

import (
	"context"
	"errors"
	"fmt"

	"linke/internal/user/domain/repository"
	"linke/internal/user/domain/valueobject"
)

// UserDomainService encapsulates domain business logic that doesn't belong to a single aggregate
type UserDomainService struct {
	userRepo repository.UserRepository
}

// NewUserDomainService creates a new UserDomainService
func NewUserDomainService(userRepo repository.UserRepository) *UserDomainService {
	return &UserDomainService{
		userRepo: userRepo,
	}
}

// ValidateUserCreation validates business rules for user creation
func (s *UserDomainService) ValidateUserCreation(ctx context.Context, email valueobject.Email) error {
	// Check if user with email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}

	if exists {
		return errors.New("user with this email already exists")
	}

	return nil
}

// ValidateOAuthUserCreation validates business rules for OAuth user creation
func (s *UserDomainService) ValidateOAuthUserCreation(ctx context.Context, email valueobject.Email, provider string, providerID valueobject.ProviderID) error {
	// Check if user with provider ID already exists
	exists, err := s.userRepo.ExistsByProviderID(ctx, provider, providerID)
	if err != nil {
		return fmt.Errorf("failed to check if OAuth user exists: %w", err)
	}

	if exists {
		return fmt.Errorf("user with %s ID already exists", provider)
	}

	// Check if email is already used by another account
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		// Email exists with different provider
		if !existingUser.Provider().Equals(valueobject.Provider{}) {
			currentProvider := existingUser.Provider().String()
			return fmt.Errorf("email is already associated with a %s account", currentProvider)
		}
	}

	return nil
}

// ValidateUsernameUniqueness ensures username is unique
func (s *UserDomainService) ValidateUsernameUniqueness(ctx context.Context, username string, excludeUserID *valueobject.UserID) error {
	exists, err := s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to check username uniqueness: %w", err)
	}

	if exists {
		// If excluding a user ID, check if the existing user is the same
		if excludeUserID != nil {
			existingUser, err := s.userRepo.FindByUsername(ctx, username)
			if err != nil {
				return fmt.Errorf("failed to find user by username: %w", err)
			}

			if existingUser != nil && existingUser.ID().Equals(*excludeUserID) {
				return nil // Username belongs to the same user
			}
		}

		return errors.New("username is already taken")
	}

	return nil
}

// GenerateUniqueUsername generates a unique username based on email
func (s *UserDomainService) GenerateUniqueUsername(ctx context.Context, baseUsername string) (string, error) {
	// Try base username first
	exists, err := s.userRepo.ExistsByUsername(ctx, baseUsername)
	if err != nil {
		return "", fmt.Errorf("failed to check username existence: %w", err)
	}

	if !exists {
		return baseUsername, nil
	}

	// Try with numbers
	for i := 1; i <= 999; i++ {
		candidate := fmt.Sprintf("%s%d", baseUsername, i)
		exists, err := s.userRepo.ExistsByUsername(ctx, candidate)
		if err != nil {
			return "", fmt.Errorf("failed to check username existence: %w", err)
		}

		if !exists {
			return candidate, nil
		}
	}

	// If all attempts failed, use timestamp-based approach
	return fmt.Sprintf("%s_%d", baseUsername, valueobject.NewUserID().ToUint()), nil
}

// IsEmailDomainAllowed checks if an email domain is allowed for registration
func (s *UserDomainService) IsEmailDomainAllowed(email valueobject.Email) bool {
	// Business rule: Block common disposable email domains
	blockedDomains := []string{
		"tempmail.org",
		"10minutemail.com",
		"mailinator.com",
		"guerrillamail.com",
	}

	domain := email.Domain()
	for _, blocked := range blockedDomains {
		if domain == blocked {
			return false
		}
	}

	return true
}
