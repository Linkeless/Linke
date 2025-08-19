package dto

import (
	"linke/internal/domains/user/entities"
	"linke/internal/shared/pool"
)

// Object pools for performance optimization
var (
	userResponsePool = pool.NewPool(
		func() *UserResponse {
			return &UserResponse{}
		},
		func(resp *UserResponse) {
			*resp = UserResponse{}
		},
	)

	userAccountBindingResponsePool = pool.NewPool(
		func() *UserAccountBindingResponse {
			return &UserAccountBindingResponse{}
		},
		func(resp *UserAccountBindingResponse) {
			*resp = UserAccountBindingResponse{}
		},
	)

	userProfilePool = pool.NewPool(
		func() *UserProfile {
			return &UserProfile{}
		},
		func(profile *UserProfile) {
			*profile = UserProfile{}
		},
	)

	userStatsPool = pool.NewPool(
		func() *UserStats {
			return &UserStats{}
		},
		func(stats *UserStats) {
			*stats = UserStats{}
		},
	)
)

// Pool management functions for UserResponse

// GetUserResponse gets a UserResponse from the pool
func GetUserResponse() *UserResponse {
	return userResponsePool.Get()
}

// PutUserResponse returns a UserResponse to the pool
func PutUserResponse(resp *UserResponse) {
	if resp == nil {
		return
	}
	userResponsePool.Put(resp)
}

// Pool management functions for UserAccountBindingResponse

// GetUserAccountBindingResponse gets a UserAccountBindingResponse from the pool
func GetUserAccountBindingResponse() *UserAccountBindingResponse {
	return userAccountBindingResponsePool.Get()
}

// PutUserAccountBindingResponse returns a UserAccountBindingResponse to the pool
func PutUserAccountBindingResponse(resp *UserAccountBindingResponse) {
	if resp == nil {
		return
	}
	userAccountBindingResponsePool.Put(resp)
}

// Pool management functions for UserProfile

// GetUserProfile gets a UserProfile from the pool
func GetUserProfile() *UserProfile {
	return userProfilePool.Get()
}

// PutUserProfile returns a UserProfile to the pool
func PutUserProfile(profile *UserProfile) {
	if profile == nil {
		return
	}
	userProfilePool.Put(profile)
}

// Pool management functions for UserStats

// GetUserStats gets a UserStats from the pool
func GetUserStats() *UserStats {
	return userStatsPool.Get()
}

// PutUserStats returns a UserStats to the pool
func PutUserStats(stats *UserStats) {
	if stats == nil {
		return
	}
	userStatsPool.Put(stats)
}

// Conversion functions

// ToUserResponse converts User entity to UserResponse DTO
func ToUserResponse(user *entities.User) *UserResponse {
	if user == nil {
		return nil
	}

	resp := GetUserResponse()

	// Primary Key
	resp.ID = user.ID

	// Core Identity Fields
	resp.Email = user.Email
	resp.Username = user.Username
	resp.Name = user.Name
	resp.Avatar = user.Avatar

	// Authentication Fields
	resp.Provider = user.Provider
	resp.Status = user.Status
	resp.Role = user.Role

	// OAuth Provider IDs
	resp.GoogleID = user.GoogleID
	resp.GitHubID = user.GitHubID
	resp.TelegramID = user.TelegramID

	// Provider Metadata
	resp.ProviderData = user.ProviderData

	// Invite Code Fields
	resp.InviteCodeID = user.InviteCodeID
	resp.InviteCodeUsed = user.InviteCodeUsed

	// Timestamp Fields
	resp.CreatedAt = user.CreatedAt
	resp.UpdatedAt = user.UpdatedAt

	// Set DeletedAt only if valid
	if user.DeletedAt.Valid {
		resp.DeletedAt = &user.DeletedAt.Time
	}

	return resp
}

// ToUserAccountBindingResponse converts UserAccountBinding entity to UserAccountBindingResponse DTO
func ToUserAccountBindingResponse(binding *entities.UserAccountBinding) *UserAccountBindingResponse {
	if binding == nil {
		return nil
	}

	resp := GetUserAccountBindingResponse()

	// Primary Key
	resp.ID = binding.ID

	// Provider Information
	resp.Provider = binding.Provider
	resp.ProviderUserID = binding.ProviderUserID

	// Provider User Data
	resp.ProviderEmail = binding.ProviderEmail
	resp.ProviderUsername = binding.ProviderUsername
	resp.ProviderName = binding.ProviderName
	resp.ProviderAvatar = binding.ProviderAvatar

	// Binding Status
	resp.IsPrimary = binding.IsPrimary
	resp.BoundAt = binding.BoundAt
	resp.LastUsedAt = binding.LastUsedAt

	// Timestamp Fields
	resp.CreatedAt = binding.CreatedAt
	resp.UpdatedAt = binding.UpdatedAt

	// Set DeletedAt only if valid
	if binding.DeletedAt.Valid {
		resp.DeletedAt = &binding.DeletedAt.Time
	}

	return resp
}

// ToUserProfile converts User entity to UserProfile DTO with computed fields
func ToUserProfile(user *entities.User, profile *UserProfile) *UserProfile {
	if user == nil {
		return nil
	}

	if profile == nil {
		profile = GetUserProfile()
	}

	// Convert user to response
	userResponse := ToUserResponse(user)
	profile.User = *userResponse
	// Return the user response to pool
	PutUserResponse(userResponse)

	// Note: Computed fields like IsEmailVerified, AccountAge, etc.
	// should be set by the calling service based on business logic

	return profile
}

// Batch conversion functions

// ToUserResponseSlice converts slice of User entities to slice of UserResponse DTOs
func ToUserResponseSlice(users []*entities.User) []*UserResponse {
	if users == nil {
		return nil
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = ToUserResponse(user)
	}
	return responses
}

// ToUserAccountBindingResponseSlice converts slice of UserAccountBinding entities to slice of UserAccountBindingResponse DTOs
func ToUserAccountBindingResponseSlice(bindings []*entities.UserAccountBinding) []*UserAccountBindingResponse {
	if bindings == nil {
		return nil
	}

	responses := make([]*UserAccountBindingResponse, len(bindings))
	for i, binding := range bindings {
		responses[i] = ToUserAccountBindingResponse(binding)
	}
	return responses
}

// Batch cleanup functions

// PutUserResponseSlice returns all UserResponse DTOs in a slice to the pool
func PutUserResponseSlice(responses []*UserResponse) {
	for _, resp := range responses {
		if resp != nil {
			PutUserResponse(resp)
		}
	}
}

// PutUserAccountBindingResponseSlice returns all UserAccountBindingResponse DTOs in a slice to the pool
func PutUserAccountBindingResponseSlice(responses []*UserAccountBindingResponse) {
	for _, resp := range responses {
		if resp != nil {
			PutUserAccountBindingResponse(resp)
		}
	}
}

// Pool statistics functions

// GetUserPoolStats returns pool statistics for monitoring using generic pool stats
func GetUserPoolStats() map[string]interface{} {
	return map[string]interface{}{
		"user_response":         userResponsePool.Stats(),
		"user_binding_response": userAccountBindingResponsePool.Stats(),
		"user_profile":          userProfilePool.Stats(),
		"user_stats":            userStatsPool.Stats(),
	}
}

// ResetUserPoolStats resets pool statistics for benchmarking
// Note: The new pool implementation doesn't support resetting stats,
// so this is a no-op function for backward compatibility
func ResetUserPoolStats() {
	// No-op: New pool implementation doesn't support stat reset
	// This function exists for backward compatibility with tests
}

// Utility functions for request validation

// ValidateProvider checks if a provider is valid for binding operations
func ValidateProvider(provider string) bool {
	validProviders := []string{"google", "github", "telegram"}
	for _, validProvider := range validProviders {
		if provider == validProvider {
			return true
		}
	}
	return false
}

// ValidateUserStatus checks if a user status is valid
func ValidateUserStatus(status string) bool {
	validStatuses := []string{"active", "inactive", "banned"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// ValidateUserRole checks if a user role is valid
func ValidateUserRole(role string) bool {
	validRoles := []string{"user", "admin"}
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}
