package service

import (
	"errors"

	"linke/internal/user/domain/valueobject"
)

// AuthenticationService handles authentication-related domain logic
type AuthenticationService struct{}

// NewAuthenticationService creates a new authentication service
func NewAuthenticationService() *AuthenticationService {
	return &AuthenticationService{}
}

// ValidateCredentials validates user credentials for login
func (s *AuthenticationService) ValidateCredentials(
	password valueobject.Password,
	candidatePassword string,
	status valueobject.UserStatus,
	provider valueobject.Provider,
) error {
	// OAuth accounts cannot authenticate with password
	if !provider.IsLocal() {
		return errors.New("password authentication not supported for OAuth accounts")
	}
	
	// Suspended users cannot authenticate
	if status.IsSuspended() {
		return errors.New("user account is suspended")
	}
	
	// Inactive users cannot authenticate
	if !status.IsActive() {
		return errors.New("user account is not active")
	}
	
	// Verify password
	if !password.Verify(candidatePassword) {
		return errors.New("invalid credentials")
	}
	
	return nil
}

// CanChangePassword checks if a user can change their password
func (s *AuthenticationService) CanChangePassword(provider valueobject.Provider) error {
	if !provider.IsLocal() {
		return errors.New("password change not supported for OAuth accounts")
	}
	return nil
}

// ValidatePasswordChange validates a password change request
func (s *AuthenticationService) ValidatePasswordChange(
	currentPassword valueobject.Password,
	oldPassword string,
	newPassword string,
) (valueobject.Password, error) {
	// Verify current password
	if !currentPassword.Verify(oldPassword) {
		return valueobject.Password{}, errors.New("current password is incorrect")
	}
	
	// Create new password value object (this will validate the new password)
	newPasswordVO, err := valueobject.NewPassword(newPassword)
	if err != nil {
		return valueobject.Password{}, err
	}
	
	return newPasswordVO, nil
}

// CanUserLogin checks if a user can login based on their status
func (s *AuthenticationService) CanUserLogin(status valueobject.UserStatus) bool {
	return status.IsActive() && !status.IsSuspended()
}