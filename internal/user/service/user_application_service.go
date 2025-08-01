package service

import (
	"context"

	"linke/internal/user/domain/aggregate"
	"linke/internal/user/service/command"
	"linke/internal/user/service/query"
)

// UserApplicationService handles user-related use cases by coordinating 
// between command handlers and query handlers
type UserApplicationService struct {
	commandHandler *command.UserCommandHandler
	queryHandler   *query.UserQueryHandler
}

// NewUserApplicationService creates a new UserApplicationService
func NewUserApplicationService(
	commandHandler *command.UserCommandHandler,
	queryHandler *query.UserQueryHandler,
) *UserApplicationService {
	return &UserApplicationService{
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
	}
}

// Command Operations - delegating to command handler

// CreateUser creates a new user with email and password
func (s *UserApplicationService) CreateUser(ctx context.Context, cmd command.CreateUserCommand) (*aggregate.User, error) {
	return s.commandHandler.CreateUser(ctx, cmd)
}

// CreateOAuthUser creates a new user from OAuth provider
func (s *UserApplicationService) CreateOAuthUser(ctx context.Context, cmd command.CreateOAuthUserCommand) (*aggregate.User, error) {
	return s.commandHandler.CreateOAuthUser(ctx, cmd)
}

// AuthenticateUser authenticates a user with email and password
func (s *UserApplicationService) AuthenticateUser(ctx context.Context, cmd command.AuthenticateUserCommand) (*aggregate.User, error) {
	return s.commandHandler.AuthenticateUser(ctx, cmd)
}

// AuthenticateOAuthUser authenticates or creates a user via OAuth
func (s *UserApplicationService) AuthenticateOAuthUser(ctx context.Context, cmd command.AuthenticateOAuthUserCommand) (*aggregate.User, error) {
	return s.commandHandler.AuthenticateOAuthUser(ctx, cmd)
}

// ChangePassword changes a user's password
func (s *UserApplicationService) ChangePassword(ctx context.Context, cmd command.ChangePasswordCommand) error {
	return s.commandHandler.ChangePassword(ctx, cmd)
}

// UpdateUserProfile updates a user's profile
func (s *UserApplicationService) UpdateUserProfile(ctx context.Context, cmd command.UpdateUserProfileCommand) error {
	return s.commandHandler.UpdateUserProfile(ctx, cmd)
}

// ChangeUserStatus changes a user's status
func (s *UserApplicationService) ChangeUserStatus(ctx context.Context, cmd command.ChangeUserStatusCommand) error {
	return s.commandHandler.ChangeUserStatus(ctx, cmd)
}

// ChangeUserRole changes a user's role
func (s *UserApplicationService) ChangeUserRole(ctx context.Context, cmd command.ChangeUserRoleCommand) error {
	return s.commandHandler.ChangeUserRole(ctx, cmd)
}

// DeleteUser soft deletes a user
func (s *UserApplicationService) DeleteUser(ctx context.Context, cmd command.DeleteUserCommand) error {
	return s.commandHandler.DeleteUser(ctx, cmd)
}

// RestoreUser restores a soft deleted user
func (s *UserApplicationService) RestoreUser(ctx context.Context, cmd command.RestoreUserCommand) error {
	return s.commandHandler.RestoreUser(ctx, cmd)
}

// SetInviteCode sets invite code information for a user
func (s *UserApplicationService) SetInviteCode(ctx context.Context, cmd command.SetInviteCodeCommand) error {
	return s.commandHandler.SetInviteCode(ctx, cmd)
}

// Query Operations - delegating to query handler

// GetUserByID gets a user by ID
func (s *UserApplicationService) GetUserByID(ctx context.Context, q query.GetUserByIDQuery) (*aggregate.User, error) {
	return s.queryHandler.GetUserByID(ctx, q)
}

// GetUserByEmail gets a user by email
func (s *UserApplicationService) GetUserByEmail(ctx context.Context, q query.GetUserByEmailQuery) (*aggregate.User, error) {
	return s.queryHandler.GetUserByEmail(ctx, q)
}

// GetUserByUsername gets a user by username
func (s *UserApplicationService) GetUserByUsername(ctx context.Context, q query.GetUserByUsernameQuery) (*aggregate.User, error) {
	return s.queryHandler.GetUserByUsername(ctx, q)
}

// GetUserByProviderID gets a user by provider ID
func (s *UserApplicationService) GetUserByProviderID(ctx context.Context, q query.GetUserByProviderIDQuery) (*aggregate.User, error) {
	return s.queryHandler.GetUserByProviderID(ctx, q)
}

// ListUsers lists users with pagination and filters
func (s *UserApplicationService) ListUsers(ctx context.Context, q query.ListUsersQuery) (*query.UserListResult, error) {
	return s.queryHandler.ListUsers(ctx, q)
}

// SearchUsers searches users by email or username
func (s *UserApplicationService) SearchUsers(ctx context.Context, q query.SearchUsersQuery) (*query.UserListResult, error) {
	return s.queryHandler.SearchUsers(ctx, q)
}

// GetUserStats gets user statistics
func (s *UserApplicationService) GetUserStats(ctx context.Context, q query.GetUserStatsQuery) (*query.UserStats, error) {
	return s.queryHandler.GetUserStats(ctx, q)
}

// CheckEmailExists checks if an email exists
func (s *UserApplicationService) CheckEmailExists(ctx context.Context, q query.CheckEmailExistsQuery) (bool, error) {
	return s.queryHandler.CheckEmailExists(ctx, q)
}

// CheckUsernameExists checks if a username exists
func (s *UserApplicationService) CheckUsernameExists(ctx context.Context, q query.CheckUsernameExistsQuery) (bool, error) {
	return s.queryHandler.CheckUsernameExists(ctx, q)
}

// CheckProviderIDExists checks if a provider ID exists
func (s *UserApplicationService) CheckProviderIDExists(ctx context.Context, q query.CheckProviderIDExistsQuery) (bool, error) {
	return s.queryHandler.CheckProviderIDExists(ctx, q)
}

// GetUsersByIDs gets multiple users by their IDs
func (s *UserApplicationService) GetUsersByIDs(ctx context.Context, q query.GetUsersByIDsQuery) ([]*aggregate.User, error) {
	return s.queryHandler.GetUsersByIDs(ctx, q)
}