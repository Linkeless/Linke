package query

import (
	"linke/internal/user/domain/valueobject"
)

// GetUserByIDQuery represents a query to get user by ID
type GetUserByIDQuery struct {
	UserID valueobject.UserID `json:"user_id" validate:"required"`
}

// GetUserByEmailQuery represents a query to get user by email
type GetUserByEmailQuery struct {
	Email string `json:"email" validate:"required,email"`
}

// GetUserByUsernameQuery represents a query to get user by username
type GetUserByUsernameQuery struct {
	Username string `json:"username" validate:"required,max=100"`
}

// GetUserByProviderIDQuery represents a query to get user by provider ID
type GetUserByProviderIDQuery struct {
	Provider   string `json:"provider" validate:"required,oneof=google github telegram"`
	ProviderID string `json:"provider_id" validate:"required,max=100"`
}

// ListUsersQuery represents a query to list users with pagination
type ListUsersQuery struct {
	Page   int    `json:"page" validate:"min=1"`
	Size   int    `json:"size" validate:"min=1,max=100"`
	Status string `json:"status,omitempty" validate:"omitempty,oneof=active inactive banned"`
	Role   string `json:"role,omitempty" validate:"omitempty,oneof=user admin"`
	Provider string `json:"provider,omitempty" validate:"omitempty,oneof=local google github telegram"`
}

// SearchUsersQuery represents a query to search users
type SearchUsersQuery struct {
	Query    string `json:"query" validate:"required,min=1,max=100"`
	Page     int    `json:"page" validate:"min=1"`
	Size     int    `json:"size" validate:"min=1,max=100"`
	Status   string `json:"status,omitempty" validate:"omitempty,oneof=active inactive banned"`
	Role     string `json:"role,omitempty" validate:"omitempty,oneof=user admin"`
	Provider string `json:"provider,omitempty" validate:"omitempty,oneof=local google github telegram"`
}

// GetUserStatsQuery represents a query to get user statistics
type GetUserStatsQuery struct {
	GroupBy string `json:"group_by,omitempty" validate:"omitempty,oneof=status role provider"`
}

// CheckEmailExistsQuery represents a query to check if email exists
type CheckEmailExistsQuery struct {
	Email string `json:"email" validate:"required,email"`
}

// CheckUsernameExistsQuery represents a query to check if username exists
type CheckUsernameExistsQuery struct {
	Username string `json:"username" validate:"required,max=100"`
}

// CheckProviderIDExistsQuery represents a query to check if provider ID exists
type CheckProviderIDExistsQuery struct {
	Provider   string `json:"provider" validate:"required,oneof=google github telegram"`
	ProviderID string `json:"provider_id" validate:"required,max=100"`
}

// GetDeletedUsersQuery represents a query to get deleted users
type GetDeletedUsersQuery struct {
	Page int `json:"page" validate:"min=1"`
	Size int `json:"size" validate:"min=1,max=100"`
}

// GetUsersByIDsQuery represents a query to get users by multiple IDs
type GetUsersByIDsQuery struct {
	UserIDs []valueobject.UserID `json:"user_ids" validate:"required,max=100"`
}

// Pagination helper
type PaginationQuery struct {
	Page int `json:"page" validate:"min=1"`
	Size int `json:"size" validate:"min=1,max=100"`
}

// GetOffset calculates the offset for database queries
func (p PaginationQuery) GetOffset() int {
	return (p.Page - 1) * p.Size
}

// GetLimit returns the limit for database queries
func (p PaginationQuery) GetLimit() int {
	return p.Size
}

// Validate pagination parameters and set defaults
func (p *PaginationQuery) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Size < 1 {
		p.Size = 10
	}
	if p.Size > 100 {
		p.Size = 100
	}
}