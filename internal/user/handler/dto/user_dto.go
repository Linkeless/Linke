package dto

import (
	"time"

	"linke/internal/user/domain/aggregate"
)

// UserResponse represents the user information returned to clients
type UserResponse struct {
	ID             uint      `json:"id" example:"1"`
	Email          string    `json:"email" example:"user@example.com"`
	Username       string    `json:"username" example:"johndoe"`
	Name           string    `json:"name" example:"John Doe"`
	Avatar         string    `json:"avatar" example:"https://example.com/avatar.jpg"`
	Provider       string    `json:"provider" example:"local"`
	Status         string    `json:"status" example:"active"`
	Role           string    `json:"role" example:"user"`
	ProviderData   *string   `json:"provider_data,omitempty" swaggertype:"string" example:"{}"`
	InviteCodeID   *uint     `json:"invite_code_id,omitempty" example:"123"`
	InviteCodeUsed *string   `json:"invite_code_used,omitempty" example:"INVITE123"`
	CreatedAt      time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt      time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" example:"2023-01-01T00:00:00Z"`

	// OAuth provider-specific IDs
	GoogleID   *string `json:"google_id,omitempty" example:"1234567890"`
	GitHubID   *string `json:"github_id,omitempty" example:"octocat"`
	TelegramID *string `json:"telegram_id,omitempty" example:"987654321"`
}

// FromUser converts User domain model to UserResponse DTO
func FromUser(user *aggregate.User) UserResponse {
	if user == nil {
		return UserResponse{}
	}

	resp := UserResponse{
		ID:             user.ID().ToUint(),
		Email:          user.Email().String(),
		Username:       user.Username().String(),
		Name:           user.Name().String(),
		Avatar:         user.Avatar().String(),
		Provider:       user.Provider().String(),
		Status:         user.Status().String(),
		Role:           user.Role().String(),
		ProviderData:   user.ProviderData(),
		InviteCodeID:   getInviteCodeID(user),
		InviteCodeUsed: getInviteCodeUsed(user),
		CreatedAt:      user.CreatedAt(),
		UpdatedAt:      user.UpdatedAt(),
	}

	// Set OAuth provider IDs
	oauthAccounts := user.OAuthAccounts()
	if googleAccount, exists := oauthAccounts["google"]; exists {
		googleIDStr := googleAccount.ProviderID().String()
		resp.GoogleID = &googleIDStr
	}
	if githubAccount, exists := oauthAccounts["github"]; exists {
		githubIDStr := githubAccount.ProviderID().String()
		resp.GitHubID = &githubIDStr
	}
	if telegramAccount, exists := oauthAccounts["telegram"]; exists {
		telegramIDStr := telegramAccount.ProviderID().String()
		resp.TelegramID = &telegramIDStr
	}

	// Set DeletedAt only if valid
	if user.IsDeleted() {
		deletedAt := user.DeletedAt()
		resp.DeletedAt = deletedAt
	}

	return resp
}

// FromUsers converts slice of User domain models to slice of UserResponse DTOs
func FromUsers(users []*aggregate.User) []UserResponse {
	if len(users) == 0 {
		return []UserResponse{}
	}

	responses := make([]UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, FromUser(user))
	}

	return responses
}

// Helper functions for invite code access
func getInviteCodeID(user *aggregate.User) *uint {
	inviteCode := user.InviteCode()
	if inviteCode.IsEmpty() {
		return nil
	}
	id := inviteCode.ID()
	return &id
}

func getInviteCodeUsed(user *aggregate.User) *string {
	inviteCode := user.InviteCode()
	if inviteCode.IsEmpty() {
		return nil
	}
	code := inviteCode.String()
	return &code
}

// CreateUserRequest represents the request structure for creating a new user
type CreateUserRequest struct {
	Email      string  `json:"email" binding:"required,email,max=255" example:"user@example.com"`
	Password   string  `json:"password" binding:"required,min=6,max=128" example:"password123"`
	Username   *string `json:"username" binding:"omitempty,max=100" example:"johndoe"`
	Name       *string `json:"name" binding:"omitempty,max=255" example:"John Doe"`
	InviteCode *string `json:"invite_code" binding:"omitempty,max=32" example:"INVITE123"`
}

// LoginRequest represents the request structure for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// ChangePasswordRequest represents the request structure for changing password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpassword123"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=128" example:"newpassword123"`
}

// UpdateProfileRequest represents the request structure for updating user profile
type UpdateProfileRequest struct {
	Name     *string `json:"name" binding:"omitempty,max=255" example:"John Doe"`
	Username *string `json:"username" binding:"omitempty,max=100" example:"johndoe"`
	Avatar   *string `json:"avatar" binding:"omitempty,max=500" example:"https://example.com/avatar.jpg"`
}

// ChangeUserStatusRequest represents the request structure for changing user status (Admin only)
type ChangeUserStatusRequest struct {
	Status string  `json:"status" binding:"required,oneof=active inactive banned" example:"active"`
	Reason *string `json:"reason" binding:"omitempty,max=500" example:"User requested deactivation"`
}

// ChangeUserRoleRequest represents the request structure for changing user role (Admin only)
type ChangeUserRoleRequest struct {
	Role   string  `json:"role" binding:"required,oneof=user admin" example:"admin"`
	Reason *string `json:"reason" binding:"omitempty,max=500" example:"Promoted to admin"`
}

// DeleteUserRequest represents the request structure for deleting a user (Admin only)
type DeleteUserRequest struct {
	Reason *string `json:"reason" binding:"omitempty,max=500" example:"Account cleanup"`
}

// RestoreUserRequest represents the request structure for restoring a deleted user (Admin only)
type RestoreUserRequest struct {
	Reason *string `json:"reason" binding:"omitempty,max=500" example:"Restore accidentally deleted account"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User  UserResponse  `json:"user"`
	Token TokenResponse `json:"token"`
}

// TokenResponse represents the token information
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	RefreshToken *string   `json:"refresh_token,omitempty"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// UserListResponse represents the paginated user list response
type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Size       int            `json:"size"`
	TotalPages int            `json:"total_pages"`
	HasNext    bool           `json:"has_next"`
	HasPrev    bool           `json:"has_prev"`
}

// UserStatsResponse represents user statistics
type UserStatsResponse struct {
	Total      int64            `json:"total"`
	ByStatus   map[string]int64 `json:"by_status,omitempty"`
	ByRole     map[string]int64 `json:"by_role,omitempty"`
	ByProvider map[string]int64 `json:"by_provider,omitempty"`
}

// ListUsersQuery represents query parameters for listing users
type ListUsersQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1" example:"1"`
	Size     int    `form:"size" binding:"omitempty,min=1,max=100" example:"10"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive banned" example:"active"`
	Role     string `form:"role" binding:"omitempty,oneof=user admin" example:"user"`
	Provider string `form:"provider" binding:"omitempty,oneof=local google github telegram" example:"local"`
}

// SearchUsersQuery represents query parameters for searching users
type SearchUsersQuery struct {
	Query    string `form:"q" binding:"required,min=1,max=100" example:"john"`
	Page     int    `form:"page" binding:"omitempty,min=1" example:"1"`
	Size     int    `form:"size" binding:"omitempty,min=1,max=100" example:"10"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive banned" example:"active"`
	Role     string `form:"role" binding:"omitempty,oneof=user admin" example:"user"`
	Provider string `form:"provider" binding:"omitempty,oneof=local google github telegram" example:"local"`
}

// UserStatsQuery represents query parameters for user statistics
type UserStatsQuery struct {
	GroupBy string `form:"group_by" binding:"omitempty,oneof=status role provider" example:"status"`
}

// OAuth related DTOs

// OAuthAuthURLRequest represents request for OAuth authorization URL
type OAuthAuthURLRequest struct {
	Provider    string  `json:"provider" binding:"required,oneof=google github telegram" example:"google"`
	State       *string `json:"state" binding:"omitempty,max=255" example:"random_state_123"`
	RedirectURI *string `json:"redirect_uri" binding:"omitempty,url,max=500" example:"https://example.com/callback"`
}

// OAuthAuthURLResponse represents OAuth authorization URL response
type OAuthAuthURLResponse struct {
	AuthURL string `json:"auth_url" example:"https://accounts.google.com/oauth/authorize?..."`
	State   string `json:"state" example:"random_state_123"`
}

// OAuthTokenExchangeRequest represents OAuth token exchange request
type OAuthTokenExchangeRequest struct {
	Provider string  `json:"provider" binding:"required,oneof=google github" example:"google"`
	Code     string  `json:"code" binding:"required" example:"authorization_code_123"`
	State    *string `json:"state" binding:"omitempty,max=255" example:"random_state_123"`
}

// TelegramAuthRequest represents Telegram authentication data
type TelegramAuthRequest struct {
	ID        string  `json:"id" binding:"required" example:"123456789"`
	FirstName string  `json:"first_name" binding:"required" example:"John"`
	LastName  *string `json:"last_name" binding:"omitempty" example:"Doe"`
	Username  *string `json:"username" binding:"omitempty" example:"johndoe"`
	PhotoURL  *string `json:"photo_url" binding:"omitempty" example:"https://t.me/i/userpic/320/johndoe.jpg"`
	AuthDate  string  `json:"auth_date" binding:"required" example:"1234567890"`
	Hash      string  `json:"hash" binding:"required" example:"hash_value_123"`
}

// Validation helper methods

// Validate validates and sets defaults for ListUsersQuery
func (q *ListUsersQuery) Validate() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 {
		q.Size = 10
	}
	if q.Size > 100 {
		q.Size = 100
	}
}

// Validate validates and sets defaults for SearchUsersQuery
func (q *SearchUsersQuery) Validate() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 {
		q.Size = 10
	}
	if q.Size > 100 {
		q.Size = 100
	}
}