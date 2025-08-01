package command

import (
	"linke/internal/user/domain/valueobject"
)

// CreateUserCommand represents a command to create a new user
type CreateUserCommand struct {
	Email      string  `json:"email" validate:"required,email,max=255"`
	Password   string  `json:"password" validate:"required,min=6,max=128"`
	Username   *string `json:"username,omitempty" validate:"omitempty,max=100"`
	Name       *string `json:"name,omitempty" validate:"omitempty,max=255"`
	InviteCode *string `json:"invite_code,omitempty" validate:"omitempty,max=32"`
}

// CreateOAuthUserCommand represents a command to create a new OAuth user
type CreateOAuthUserCommand struct {
	Email        string  `json:"email" validate:"required,email,max=255"`
	Name         string  `json:"name" validate:"required,max=255"`
	Username     *string `json:"username,omitempty" validate:"omitempty,max=100"`
	Avatar       *string `json:"avatar,omitempty" validate:"omitempty,max=500"`
	Provider     string  `json:"provider" validate:"required,oneof=google github telegram"`
	ProviderID   string  `json:"provider_id" validate:"required,max=100"`
	ProviderData *string `json:"provider_data,omitempty"`
}

// AuthenticateUserCommand represents a command to authenticate a user
type AuthenticateUserCommand struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// AuthenticateOAuthUserCommand represents a command to authenticate via OAuth
type AuthenticateOAuthUserCommand struct {
	Provider     string  `json:"provider" validate:"required,oneof=google github telegram"`
	ProviderID   string  `json:"provider_id" validate:"required,max=100"`
	Email        string  `json:"email" validate:"required,email,max=255"`
	Name         string  `json:"name" validate:"required,max=255"`
	Username     *string `json:"username,omitempty" validate:"omitempty,max=100"`
	Avatar       *string `json:"avatar,omitempty" validate:"omitempty,max=500"`
	ProviderData *string `json:"provider_data,omitempty"`
	IPAddress    string  `json:"ip_address,omitempty"`
	UserAgent    string  `json:"user_agent,omitempty"`
}

// ChangePasswordCommand represents a command to change user password
type ChangePasswordCommand struct {
	UserID      valueobject.UserID `json:"user_id" validate:"required"`
	OldPassword string             `json:"old_password" validate:"required"`
	NewPassword string             `json:"new_password" validate:"required,min=6,max=128"`
}

// UpdateUserProfileCommand represents a command to update user profile
type UpdateUserProfileCommand struct {
	UserID   valueobject.UserID `json:"user_id" validate:"required"`
	Name     *string            `json:"name,omitempty" validate:"omitempty,max=255"`
	Username *string            `json:"username,omitempty" validate:"omitempty,max=100"`
	Avatar   *string            `json:"avatar,omitempty" validate:"omitempty,max=500"`
}

// ChangeUserStatusCommand represents a command to change user status
type ChangeUserStatusCommand struct {
	UserID    valueobject.UserID `json:"user_id" validate:"required"`
	NewStatus string             `json:"new_status" validate:"required,oneof=active inactive banned"`
	Reason    *string            `json:"reason,omitempty" validate:"omitempty,max=500"`
}

// ChangeUserRoleCommand represents a command to change user role
type ChangeUserRoleCommand struct {
	UserID      valueobject.UserID `json:"user_id" validate:"required"`
	NewRole     string             `json:"new_role" validate:"required,oneof=user admin"`
	ChangedBy   valueobject.UserID `json:"changed_by" validate:"required"`
	Reason      *string            `json:"reason,omitempty" validate:"omitempty,max=500"`
}

// DeleteUserCommand represents a command to delete a user
type DeleteUserCommand struct {
	UserID    valueobject.UserID `json:"user_id" validate:"required"`
	DeletedBy valueobject.UserID `json:"deleted_by" validate:"required"`
	Reason    *string            `json:"reason,omitempty" validate:"omitempty,max=500"`
}

// RestoreUserCommand represents a command to restore a deleted user
type RestoreUserCommand struct {
	UserID     valueobject.UserID `json:"user_id" validate:"required"`
	RestoredBy valueobject.UserID `json:"restored_by" validate:"required"`
	Reason     *string            `json:"reason,omitempty" validate:"omitempty,max=500"`
}

// SetInviteCodeCommand represents a command to set invite code for a user
type SetInviteCodeCommand struct {
	UserID       valueobject.UserID `json:"user_id" validate:"required"`
	InviteCodeID uint               `json:"invite_code_id" validate:"required"`
	InviteCode   string             `json:"invite_code" validate:"required,max=32"`
}