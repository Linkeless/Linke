package command

import (
	"context"
	"fmt"

	"linke/internal/shared/domain"
	"linke/internal/user/domain/aggregate"
	"linke/internal/user/domain/repository"
	"linke/internal/user/domain/service"
	"linke/internal/user/domain/valueobject"
)

// UserCommandHandler handles user-related commands
type UserCommandHandler struct {
	userRepo        repository.UserRepository
	userDomainSvc   *service.UserDomainService
	eventPublisher  domain.EventPublisher
	txManager       TransactionManager
}

// TransactionManager is imported from shared domain
type TransactionManager = domain.TransactionManager

// NewUserCommandHandler creates a new user command handler
func NewUserCommandHandler(
	userRepo repository.UserRepository,
	userDomainSvc *service.UserDomainService,
	eventPublisher domain.EventPublisher,
	txManager TransactionManager,
) *UserCommandHandler {
	return &UserCommandHandler{
		userRepo:       userRepo,
		userDomainSvc:  userDomainSvc,
		eventPublisher: eventPublisher,
		txManager:      txManager,
	}
}

// CreateUser creates a new user with email and password
func (h *UserCommandHandler) CreateUser(ctx context.Context, cmd CreateUserCommand) (*aggregate.User, error) {
	var createdUser *aggregate.User
	
	// Execute in transaction to ensure consistency
	err := h.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Create email value object
		email, err := valueobject.NewEmail(cmd.Email)
		if err != nil {
			return fmt.Errorf("invalid email: %w", err)
		}

		// Validate business rules
		if err := h.userDomainSvc.ValidateUserCreation(txCtx, email); err != nil {
			return err
		}

		// Validate email domain
		if !h.userDomainSvc.IsEmailDomainAllowed(email) {
			return fmt.Errorf("email domain is not allowed")
		}

		// Create user aggregate
		user, err := aggregate.NewUser(cmd.Email, cmd.Password)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Handle optional fields
		if cmd.Username != nil {
			if err := h.userDomainSvc.ValidateUsernameUniqueness(txCtx, *cmd.Username, nil); err != nil {
				return err
			}
			user.UpdateProfile(*cmd.Name, *cmd.Username, "")
		}

		if cmd.Name != nil {
			user.UpdateProfile(user.Username().String(), *cmd.Name, "")
		}

		// Save user within transaction
		if err := h.userRepo.Save(txCtx, user); err != nil {
			return fmt.Errorf("failed to save user: %w", err)
		}
		
		createdUser = user
		return nil
	})
	
	if err != nil {
		return nil, err
	}

	// Publish domain events after successful transaction
	h.publishDomainEvents(ctx, createdUser)

	return createdUser, nil
}

// CreateOAuthUser creates a new user from OAuth provider
func (h *UserCommandHandler) CreateOAuthUser(ctx context.Context, cmd CreateOAuthUserCommand) (*aggregate.User, error) {
	// Create email value object
	email, err := valueobject.NewEmail(cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// Create provider ID value object
	providerID, err := valueobject.NewProviderID(cmd.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("invalid provider ID: %w", err)
	}

	// Validate business rules
	if err := h.userDomainSvc.ValidateOAuthUserCreation(ctx, email, cmd.Provider, providerID); err != nil {
		return nil, err
	}

	// Prepare user data
	username := ""
	if cmd.Username != nil {
		username = *cmd.Username
	}

	avatar := ""
	if cmd.Avatar != nil {
		avatar = *cmd.Avatar
	}

	// Create user aggregate
	user, err := aggregate.NewUserFromOAuth(
		cmd.Email,
		cmd.Name,
		username,
		avatar,
		cmd.Provider,
		cmd.ProviderID,
		cmd.ProviderData,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth user: %w", err)
	}

	// Generate unique username if not provided or conflicts
	if username == "" || h.usernameExists(ctx, username) {
		baseUsername := email.LocalPart()
		uniqueUsername, err := h.userDomainSvc.GenerateUniqueUsername(ctx, baseUsername)
		if err != nil {
			return nil, fmt.Errorf("failed to generate unique username: %w", err)
		}
		user.UpdateProfile(uniqueUsername, user.Name().String(), user.Avatar().String())
	}

	// Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save OAuth user: %w", err)
	}

	// Publish domain events
	h.publishDomainEvents(ctx, user)

	return user, nil
}

// AuthenticateUser authenticates a user with email and password
func (h *UserCommandHandler) AuthenticateUser(ctx context.Context, cmd AuthenticateUserCommand) (*aggregate.User, error) {
	// Create email value object
	email, err := valueobject.NewEmail(cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// Find user by email
	user, err := h.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Authenticate user (this will raise domain events)
	if err := user.Authenticate(cmd.Password, cmd.IPAddress, cmd.UserAgent); err != nil {
		// Publish domain events even for failed authentication
		h.publishDomainEvents(ctx, user)
		return nil, err
	}

	// Update user to persist any changes (like failed login attempts)
	if err := h.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Publish domain events
	h.publishDomainEvents(ctx, user)

	return user, nil
}

// AuthenticateOAuthUser authenticates or creates a user via OAuth
func (h *UserCommandHandler) AuthenticateOAuthUser(ctx context.Context, cmd AuthenticateOAuthUserCommand) (*aggregate.User, error) {
	// Create provider ID value object
	providerID, err := valueobject.NewProviderID(cmd.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("invalid provider ID: %w", err)
	}

	// Try to find existing user by provider ID
	user, err := h.userRepo.FindByProviderID(ctx, cmd.Provider, providerID)
	if err == nil && user != nil {
		// User exists, update profile if needed
		h.updateOAuthUserProfile(user, cmd)

		// Save updates
		if err := h.userRepo.Update(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		// Publish domain events
		h.publishDomainEvents(ctx, user)

		return user, nil
	}

	// User doesn't exist, create new OAuth user
	createCmd := CreateOAuthUserCommand{
		Email:        cmd.Email,
		Name:         cmd.Name,
		Username:     cmd.Username,
		Avatar:       cmd.Avatar,
		Provider:     cmd.Provider,
		ProviderID:   cmd.ProviderID,
		ProviderData: cmd.ProviderData,
	}

	return h.CreateOAuthUser(ctx, createCmd)
}

// ChangePassword changes a user's password
func (h *UserCommandHandler) ChangePassword(ctx context.Context, cmd ChangePasswordCommand) error {
	// Find user
	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Change password (this will validate and raise domain events)
	if err := user.ChangePassword(cmd.OldPassword, cmd.NewPassword); err != nil {
		return err
	}

	// Save user
	if err := h.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Publish domain events
	h.publishDomainEvents(ctx, user)

	return nil
}

// UpdateUserProfile updates a user's profile
func (h *UserCommandHandler) UpdateUserProfile(ctx context.Context, cmd UpdateUserProfileCommand) error {
	// Find user
	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Validate username uniqueness if provided
	if cmd.Username != nil {
		if err := h.userDomainSvc.ValidateUsernameUniqueness(ctx, *cmd.Username, &cmd.UserID); err != nil {
			return err
		}
	}

	// Prepare values
	name := user.Name().String()
	if cmd.Name != nil {
		name = *cmd.Name
	}

	username := user.Username().String()
	if cmd.Username != nil {
		username = *cmd.Username
	}

	avatar := user.Avatar().String()
	if cmd.Avatar != nil {
		avatar = *cmd.Avatar
	}

	// Update profile (this will raise domain events if there are changes)
	if err := user.UpdateProfile(name, username, avatar); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	// Save user
	if err := h.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Publish domain events
	h.publishDomainEvents(ctx, user)

	return nil
}

// ChangeUserStatus changes a user's status
func (h *UserCommandHandler) ChangeUserStatus(ctx context.Context, cmd ChangeUserStatusCommand) error {
	// Find user
	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Change status (this will validate and raise domain events)
	if err := user.ChangeStatus(cmd.NewStatus); err != nil {
		return err
	}

	// Save user
	if err := h.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Publish domain events
	h.publishDomainEvents(ctx, user)

	return nil
}

// ChangeUserRole changes a user's role
func (h *UserCommandHandler) ChangeUserRole(ctx context.Context, cmd ChangeUserRoleCommand) error {
	// Find user
	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Change role (this will validate and raise domain events)
	if err := user.ChangeRole(cmd.NewRole); err != nil {
		return err
	}

	// Save user
	if err := h.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Publish domain events
	h.publishDomainEvents(ctx, user)

	return nil
}

// DeleteUser soft deletes a user
func (h *UserCommandHandler) DeleteUser(ctx context.Context, cmd DeleteUserCommand) error {
	// Find user
	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Soft delete user
	user.SoftDelete()

	// Save user
	if err := h.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Publish domain events
	h.publishDomainEvents(ctx, user)

	return nil
}

// RestoreUser restores a soft deleted user
func (h *UserCommandHandler) RestoreUser(ctx context.Context, cmd RestoreUserCommand) error {
	// Find user (including deleted ones)
	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if !user.IsDeleted() {
		return fmt.Errorf("user is not deleted")
	}

	// Restore user
	user.Restore()

	// Save user
	if err := h.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to restore user: %w", err)
	}

	// Publish domain events
	h.publishDomainEvents(ctx, user)

	return nil
}

// SetInviteCode sets invite code information for a user
func (h *UserCommandHandler) SetInviteCode(ctx context.Context, cmd SetInviteCodeCommand) error {
	// Find user
	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Set invite code
	user.SetInviteCode(cmd.InviteCodeID, cmd.InviteCode)

	// Save user
	if err := h.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Publish domain events
	h.publishDomainEvents(ctx, user)

	return nil
}

// publishDomainEvents publishes domain events and clears them from the aggregate
func (h *UserCommandHandler) publishDomainEvents(ctx context.Context, user *aggregate.User) error {
	events := user.DomainEvents()
	if len(events) == 0 {
		return nil
	}

	// Publish events using batch publish
	if err := h.eventPublisher.PublishBatch(ctx, events); err != nil {
		return fmt.Errorf("failed to publish domain events: %w", err)
	}

	// Clear events from aggregate
	user.ClearDomainEvents()

	return nil
}

// updateOAuthUserProfile updates OAuth user profile if data has changed
func (h *UserCommandHandler) updateOAuthUserProfile(user *aggregate.User, cmd AuthenticateOAuthUserCommand) {
	name := cmd.Name
	username := user.Username().String()
	if cmd.Username != nil {
		username = *cmd.Username
	}
	
	avatar := ""
	if cmd.Avatar != nil {
		avatar = *cmd.Avatar
	}

	// Only update if there are actual changes
	if user.Name().String() != name || user.Avatar().String() != avatar {
		user.UpdateProfile(username, name, avatar)
	}
}

// usernameExists checks if a username exists (helper method)
func (h *UserCommandHandler) usernameExists(ctx context.Context, username string) bool {
	exists, _ := h.userRepo.ExistsByUsername(ctx, username)
	return exists
}