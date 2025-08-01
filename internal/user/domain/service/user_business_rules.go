package service

import (
	"linke/internal/user/domain/valueobject"
)

// UserBusinessRules encapsulates business rules for user operations
type UserBusinessRules struct{}

// NewUserBusinessRules creates a new instance of user business rules
func NewUserBusinessRules() *UserBusinessRules {
	return &UserBusinessRules{}
}

// CanChangeUserStatus validates if a user status can be changed
func (r *UserBusinessRules) CanChangeUserStatus(
	currentUserID, targetUserID valueobject.UserID,
	currentUserRole valueobject.UserRole,
	targetUserStatus, newStatus valueobject.UserStatus,
) error {
	// Only admins can change user status
	if !currentUserRole.IsAdmin() {
		return valueobject.NewBusinessRuleViolationError(
			"STATUS_CHANGE_PERMISSION",
			"only administrators can change user status",
		)
	}
	
	// Users cannot suspend themselves
	if currentUserID.Equals(targetUserID) && newStatus.IsSuspended() {
		return valueobject.ErrCannotSuspendSelf
	}
	
	return nil
}

// CanChangeUserRole validates if a user role can be changed
func (r *UserBusinessRules) CanChangeUserRole(
	currentUserID, targetUserID valueobject.UserID,
	currentUserRole valueobject.UserRole,
	targetUserRole, newRole valueobject.UserRole,
) error {
	// Only admins can change user roles
	if !currentUserRole.IsAdmin() {
		return valueobject.NewBusinessRuleViolationError(
			"ROLE_CHANGE_PERMISSION",
			"only administrators can change user roles",
		)
	}
	
	// Users cannot change their own role
	if currentUserID.Equals(targetUserID) {
		return valueobject.NewBusinessRuleViolationError(
			"SELF_ROLE_CHANGE",
			"users cannot change their own role",
		)
	}
	
	return nil
}

// CanDeleteUser validates if a user can be deleted
func (r *UserBusinessRules) CanDeleteUser(
	currentUserID, targetUserID valueobject.UserID,
	currentUserRole, targetUserRole valueobject.UserRole,
	isLastAdmin bool,
) error {
	// Only admins can delete users
	if !currentUserRole.IsAdmin() {
		return valueobject.NewBusinessRuleViolationError(
			"DELETE_PERMISSION",
			"only administrators can delete users",
		)
	}
	
	// Users cannot delete themselves
	if currentUserID.Equals(targetUserID) {
		return valueobject.NewBusinessRuleViolationError(
			"SELF_DELETE",
			"users cannot delete themselves",
		)
	}
	
	// Cannot delete the last admin
	if targetUserRole.IsAdmin() && isLastAdmin {
		return valueobject.ErrCannotDeleteLastAdmin
	}
	
	return nil
}

// ValidateProfileUpdate validates profile update constraints
func (r *UserBusinessRules) ValidateProfileUpdate(
	currentUserID, targetUserID valueobject.UserID,
	currentUserRole valueobject.UserRole,
) error {
	// Users can update their own profile
	if currentUserID.Equals(targetUserID) {
		return nil
	}
	
	// Admins can update any profile
	if currentUserRole.IsAdmin() {
		return nil
	}
	
	return valueobject.NewBusinessRuleViolationError(
		"PROFILE_UPDATE_PERMISSION",
		"users can only update their own profile",
	)
}

// ValidatePasswordChange validates password change constraints
func (r *UserBusinessRules) ValidatePasswordChange(
	currentUserID, targetUserID valueobject.UserID,
	currentUserRole valueobject.UserRole,
) error {
	// Users can only change their own password
	if !currentUserID.Equals(targetUserID) {
		return valueobject.NewBusinessRuleViolationError(
			"PASSWORD_CHANGE_PERMISSION",
			"users can only change their own password",
		)
	}
	
	return nil
}