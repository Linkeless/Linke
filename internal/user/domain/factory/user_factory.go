package factory

import (
	"linke/internal/user/domain/aggregate"
)

// UserFactory provides factory methods for creating User aggregates
type UserFactory struct{}

// NewUserFactory creates a new UserFactory
func NewUserFactory() *UserFactory {
	return &UserFactory{}
}

// CreateLocalUser creates a new local user account
func (f *UserFactory) CreateLocalUser(email, password string) (*aggregate.User, error) {
	return aggregate.NewUser(email, password)
}

// CreateOAuthUser creates a new OAuth user account with comprehensive validation
func (f *UserFactory) CreateOAuthUser(req CreateOAuthUserRequest) (*aggregate.User, error) {
	return aggregate.NewUserFromOAuth(
		req.Email,
		req.Name,
		req.Username,
		req.Avatar,
		req.Provider,
		req.ProviderID,
		req.ProviderData,
	)
}

// CreateOAuthUserRequest represents the parameters for creating an OAuth user
type CreateOAuthUserRequest struct {
	Email        string
	Name         string
	Username     string
	Avatar       string
	Provider     string
	ProviderID   string
	ProviderData *string
}

// CreateUserFromRegistration creates a user from a registration request
type RegistrationRequest struct {
	Email           string
	Password        string
	Username        string
	Name            string
	InviteCodeID    *uint
	InviteCodeStr   *string
}

// CreateUserFromRegistration creates a user from registration data
func (f *UserFactory) CreateUserFromRegistration(req RegistrationRequest) (*aggregate.User, error) {
	// Create the basic user
	user, err := aggregate.NewUser(req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	
	// Update profile if custom values provided
	if req.Username != "" || req.Name != "" {
		err = user.UpdateProfile(req.Username, req.Name, "")
		if err != nil {
			return nil, err
		}
	}
	
	// Set invite code if provided
	if req.InviteCodeID != nil && req.InviteCodeStr != nil {
		err = user.SetInviteCode(*req.InviteCodeID, *req.InviteCodeStr)
		if err != nil {
			return nil, err
		}
	}
	
	return user, nil
}