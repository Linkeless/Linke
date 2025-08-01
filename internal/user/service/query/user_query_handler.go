package query

import (
	"context"
	"fmt"

	"linke/internal/user/domain/aggregate"
	"linke/internal/user/domain/repository"
	"linke/internal/user/domain/valueobject"
)

// UserQueryHandler handles user-related queries
type UserQueryHandler struct {
	userRepo repository.UserReadRepository
}

// NewUserQueryHandler creates a new user query handler
func NewUserQueryHandler(userRepo repository.UserReadRepository) *UserQueryHandler {
	return &UserQueryHandler{
		userRepo: userRepo,
	}
}

// GetUserByID gets a user by ID
func (h *UserQueryHandler) GetUserByID(ctx context.Context, q GetUserByIDQuery) (*aggregate.User, error) {
	user, err := h.userRepo.FindByID(ctx, q.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// GetUserByEmail gets a user by email
func (h *UserQueryHandler) GetUserByEmail(ctx context.Context, q GetUserByEmailQuery) (*aggregate.User, error) {
	email, err := valueobject.NewEmail(q.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	user, err := h.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// GetUserByUsername gets a user by username
func (h *UserQueryHandler) GetUserByUsername(ctx context.Context, q GetUserByUsernameQuery) (*aggregate.User, error) {
	user, err := h.userRepo.FindByUsername(ctx, q.Username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// GetUserByProviderID gets a user by provider ID
func (h *UserQueryHandler) GetUserByProviderID(ctx context.Context, q GetUserByProviderIDQuery) (*aggregate.User, error) {
	providerID, err := valueobject.NewProviderID(q.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("invalid provider ID: %w", err)
	}

	user, err := h.userRepo.FindByProviderID(ctx, q.Provider, providerID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// UserListResult represents the result of a user list query
type UserListResult struct {
	Users      []*aggregate.User `json:"users"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Size       int           `json:"size"`
	TotalPages int           `json:"total_pages"`
	HasNext    bool          `json:"has_next"`
	HasPrev    bool          `json:"has_prev"`
}

// ListUsers lists users with pagination and filters
func (h *UserQueryHandler) ListUsers(ctx context.Context, q ListUsersQuery) (*UserListResult, error) {
	// Validate and set defaults
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 {
		q.Size = 10
	}
	if q.Size > 100 {
		q.Size = 100
	}

	offset := (q.Page - 1) * q.Size
	limit := q.Size

	var users []*aggregate.User
	var total int64
	var err error

	// Apply filters based on query parameters
	switch {
	case q.Status != "" && q.Role != "" && q.Provider != "":
		// Multiple filters - need to implement combined query
		users, err = h.listUsersWithMultipleFilters(ctx, q, offset, limit)
		if err != nil {
			return nil, err
		}
		total, err = h.countUsersWithMultipleFilters(ctx, q)
		
	case q.Status != "":
		status, err := valueobject.NewUserStatus(q.Status)
		if err != nil {
			return nil, fmt.Errorf("invalid status: %w", err)
		}
		users, err = h.userRepo.FindByStatus(ctx, status, offset, limit)
		if err != nil {
			return nil, err
		}
		total, err = h.userRepo.CountByStatus(ctx, status)
		
	case q.Role != "":
		role, err := valueobject.NewUserRole(q.Role)
		if err != nil {
			return nil, fmt.Errorf("invalid role: %w", err)
		}
		users, err = h.userRepo.FindByRole(ctx, role, offset, limit)
		if err != nil {
			return nil, err
		}
		total, err = h.userRepo.CountByRole(ctx, role)
		
	case q.Provider != "":
		provider, err := valueobject.NewProvider(q.Provider)
		if err != nil {
			return nil, fmt.Errorf("invalid provider: %w", err)
		}
		users, err = h.userRepo.FindByProvider(ctx, provider, offset, limit)
		if err != nil {
			return nil, err
		}
		total, err = h.userRepo.CountByProvider(ctx, provider)
		
	default:
		users, err = h.userRepo.FindAll(ctx, offset, limit)
		if err != nil {
			return nil, err
		}
		total, err = h.userRepo.Count(ctx)
	}

	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(q.Size) - 1) / int64(q.Size))
	hasNext := q.Page < totalPages
	hasPrev := q.Page > 1

	return &UserListResult{
		Users:      users,
		Total:      total,
		Page:       q.Page,
		Size:       q.Size,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// SearchUsers searches users by email or username
func (h *UserQueryHandler) SearchUsers(ctx context.Context, q SearchUsersQuery) (*UserListResult, error) {
	// Validate and set defaults
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 {
		q.Size = 10
	}
	if q.Size > 100 {
		q.Size = 100
	}

	offset := (q.Page - 1) * q.Size
	limit := q.Size

	users, err := h.userRepo.SearchByEmailOrUsername(ctx, q.Query, offset, limit)
	if err != nil {
		return nil, err
	}

	// For search, we might not need exact count for performance reasons
	// You could implement a separate count query if needed
	total := int64(len(users))
	if len(users) == q.Size {
		// There might be more results
		total = int64(q.Page * q.Size)
	}

	totalPages := int((total + int64(q.Size) - 1) / int64(q.Size))
	hasNext := len(users) == q.Size
	hasPrev := q.Page > 1

	return &UserListResult{
		Users:      users,
		Total:      total,
		Page:       q.Page,
		Size:       q.Size,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// UserStats represents user statistics
type UserStats struct {
	Total      int64            `json:"total"`
	ByStatus   map[string]int64 `json:"by_status,omitempty"`
	ByRole     map[string]int64 `json:"by_role,omitempty"`
	ByProvider map[string]int64 `json:"by_provider,omitempty"`
}

// GetUserStats gets user statistics
func (h *UserQueryHandler) GetUserStats(ctx context.Context, q GetUserStatsQuery) (*UserStats, error) {
	stats := &UserStats{}

	// Get total count
	total, err := h.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.Total = total

	// Get stats by group
	switch q.GroupBy {
	case "status":
		stats.ByStatus = make(map[string]int64)
		for _, status := range []string{"active", "inactive", "banned"} {
			statusVO, _ := valueobject.NewUserStatus(status)
			count, err := h.userRepo.CountByStatus(ctx, statusVO)
			if err != nil {
				return nil, err
			}
			stats.ByStatus[status] = count
		}

	case "role":
		stats.ByRole = make(map[string]int64)
		for _, role := range []string{"user", "admin"} {
			roleVO, _ := valueobject.NewUserRole(role)
			count, err := h.userRepo.CountByRole(ctx, roleVO)
			if err != nil {
				return nil, err
			}
			stats.ByRole[role] = count
		}

	case "provider":
		stats.ByProvider = make(map[string]int64)
		for _, provider := range []string{"local", "google", "github", "telegram"} {
			providerVO, _ := valueobject.NewProvider(provider)
			count, err := h.userRepo.CountByProvider(ctx, providerVO)
			if err != nil {
				return nil, err
			}
			stats.ByProvider[provider] = count
		}

	default:
		// Get all stats
		stats.ByStatus = make(map[string]int64)
		stats.ByRole = make(map[string]int64)
		stats.ByProvider = make(map[string]int64)

		for _, status := range []string{"active", "inactive", "banned"} {
			statusVO, _ := valueobject.NewUserStatus(status)
			count, err := h.userRepo.CountByStatus(ctx, statusVO)
			if err != nil {
				return nil, err
			}
			stats.ByStatus[status] = count
		}

		for _, role := range []string{"user", "admin"} {
			roleVO, _ := valueobject.NewUserRole(role)
			count, err := h.userRepo.CountByRole(ctx, roleVO)
			if err != nil {
				return nil, err
			}
			stats.ByRole[role] = count
		}

		for _, provider := range []string{"local", "google", "github", "telegram"} {
			providerVO, _ := valueobject.NewProvider(provider)
			count, err := h.userRepo.CountByProvider(ctx, providerVO)
			if err != nil {
				return nil, err
			}
			stats.ByProvider[provider] = count
		}
	}

	return stats, nil
}

// CheckEmailExists checks if an email exists
func (h *UserQueryHandler) CheckEmailExists(ctx context.Context, q CheckEmailExistsQuery) (bool, error) {
	email, err := valueobject.NewEmail(q.Email)
	if err != nil {
		return false, fmt.Errorf("invalid email: %w", err)
	}

	return h.userRepo.ExistsByEmail(ctx, email)
}

// CheckUsernameExists checks if a username exists
func (h *UserQueryHandler) CheckUsernameExists(ctx context.Context, q CheckUsernameExistsQuery) (bool, error) {
	return h.userRepo.ExistsByUsername(ctx, q.Username)
}

// CheckProviderIDExists checks if a provider ID exists
func (h *UserQueryHandler) CheckProviderIDExists(ctx context.Context, q CheckProviderIDExistsQuery) (bool, error) {
	providerID, err := valueobject.NewProviderID(q.ProviderID)
	if err != nil {
		return false, fmt.Errorf("invalid provider ID: %w", err)
	}

	return h.userRepo.ExistsByProviderID(ctx, q.Provider, providerID)
}

// GetDeletedUsers gets deleted users with pagination
// Note: This method is not available in the read-only repository
// It requires write repository access due to soft delete operations
func (h *UserQueryHandler) GetDeletedUsers(ctx context.Context, q GetDeletedUsersQuery) (*UserListResult, error) {
	// This functionality would need to be implemented with a write repository
	// or by extending the UserReadRepository interface
	return nil, fmt.Errorf("GetDeletedUsers is not available with read-only repository")
}

// GetUsersByIDs gets multiple users by their IDs
func (h *UserQueryHandler) GetUsersByIDs(ctx context.Context, q GetUsersByIDsQuery) ([]*aggregate.User, error) {
	return h.userRepo.FindByIDs(ctx, q.UserIDs)
}

// Helper methods for complex queries

// listUsersWithMultipleFilters handles queries with multiple filters
func (h *UserQueryHandler) listUsersWithMultipleFilters(ctx context.Context, q ListUsersQuery, offset, limit int) ([]*aggregate.User, error) {
	// This is a simplified implementation
	// In a real application, you might want to implement this in the repository layer
	// with proper SQL joins or use a more sophisticated query builder
	
	// Start with all users and apply filters
	users, err := h.userRepo.FindAll(ctx, 0, 1000) // Get a larger set first
	if err != nil {
		return nil, err
	}

	// Apply filters
	var filtered []*aggregate.User
	for _, user := range users {
		if h.matchesFilters(user, q) {
			filtered = append(filtered, user)
		}
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(filtered) {
		return []*aggregate.User{}, nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], nil
}

// countUsersWithMultipleFilters counts users matching multiple filters
func (h *UserQueryHandler) countUsersWithMultipleFilters(ctx context.Context, q ListUsersQuery) (int64, error) {
	// Similar to above, this is simplified
	users, err := h.userRepo.FindAll(ctx, 0, 10000) // Large limit for counting
	if err != nil {
		return 0, err
	}

	count := int64(0)
	for _, user := range users {
		if h.matchesFilters(user, q) {
			count++
		}
	}

	return count, nil
}

// matchesFilters checks if a user matches the query filters
func (h *UserQueryHandler) matchesFilters(user *aggregate.User, q ListUsersQuery) bool {
	if q.Status != "" && user.Status().String() != q.Status {
		return false
	}
	if q.Role != "" && user.Role().String() != q.Role {
		return false
	}
	if q.Provider != "" && user.Provider().String() != q.Provider {
		return false
	}
	return true
}