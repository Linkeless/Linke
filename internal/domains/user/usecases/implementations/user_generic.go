package implementations

import (
	"context"
	"fmt"
	"strings"

	"linke/internal/domains/user/entities"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/services"
)

// UserServiceGeneric demonstrates the new generic service implementation
// This service extends the base generic service with user-specific operations
type UserServiceGeneric struct {
	*services.BaseServiceImpl[entities.User, uint]
	userRepo framework.GenericRepository[entities.User, uint]
}

// NewUserServiceGeneric creates a new generic user service instance
func NewUserServiceGeneric(
	userRepo framework.GenericRepository[entities.User, uint],
	logger framework.Logger,
	eventPub framework.EventPublisher,
	validator framework.Validator,
) interfaces.UserService {
	service := &UserServiceGeneric{
		BaseServiceImpl: services.NewBaseService("user-service", userRepo, logger, eventPub, validator),
		userRepo:        userRepo,
	}
	return service
}

// Domain-specific user operations that don't fit the generic pattern

func (s *UserServiceGeneric) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	// Custom query for finding user by email
	filters := map[string]interface{}{
		"email": email,
	}
	
	req := &framework.ListRequest{Limit: 1, Offset: 0}
	result, _, err := s.userRepo.ListWithFilters(ctx, filters, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("user with email %s not found", email)
	}

	return result[0], nil
}

func (s *UserServiceGeneric) GetActiveUserByID(ctx context.Context, id uint) (*entities.User, error) {
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.Status != "active" {
		return nil, fmt.Errorf("user %d is not active", id)
	}

	return user, nil
}

func (s *UserServiceGeneric) GetActiveUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user.Status != "active" {
		return nil, fmt.Errorf("user with email %s is not active", email)
	}

	return user, nil
}

func (s *UserServiceGeneric) ListUsersByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error) {
	filters := map[string]interface{}{
		"provider": provider,
	}
	
	req := &framework.ListRequest{Limit: limit, Offset: offset}
	response, err := s.ListWithFilters(ctx, filters, req)
	if err != nil {
		return nil, 0, err
	}

	return response.Data, response.Total, nil
}

func (s *UserServiceGeneric) UpdateUserRole(ctx context.Context, id uint, role string) (*entities.User, error) {
	// Get existing user
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update role
	user.Role = role
	
	// Use generic update
	updateReq := &framework.UpdateRequest[entities.User]{
		Data: user,
		Options: &framework.UpdateOptions{
			PublishEvents: true,
		},
	}

	return s.Update(ctx, id, updateReq)
}

func (s *UserServiceGeneric) GetUserStats(ctx context.Context) (*interfaces.UserStats, error) {
	// Use generic statistics as base
	stats, err := s.GetStatistics(ctx)
	if err != nil {
		return nil, err
	}

	// Get provider-specific counts
	providers := []string{"email", "google", "github", "telegram"}
	byProvider := make(map[string]int64)
	
	for _, provider := range providers {
		count, err := s.CountByStatus(ctx, provider)
		if err != nil {
			byProvider[provider] = 0
		} else {
			byProvider[provider] = count
		}
	}

	// Get recent signups (last 7 days)
	// This would need time-based repository operations - simplified for now
	recentSignups := int64(0)

	return &interfaces.UserStats{
		TotalUsers:    stats.TotalCount,
		ActiveUsers:   stats.ActiveCount,
		InactiveUsers: stats.InactiveCount,
		BannedUsers:   stats.StatusCounts["banned"],
		DeletedUsers:  stats.DeletedCount,
		ByProvider:    byProvider,
		RecentSignups: recentSignups,
	}, nil
}

// Core CRUD operations (maintaining original signatures for backward compatibility)

func (s *UserServiceGeneric) GetUserByID(ctx context.Context, id uint) (*entities.User, error) {
	return s.BaseServiceImpl.GetByID(ctx, id)
}

func (s *UserServiceGeneric) UpdateUserStatus(ctx context.Context, id uint, status string) (*entities.User, error) {
	return s.BaseServiceImpl.UpdateStatus(ctx, id, status)
}

// Legacy method support for backward compatibility
// These methods delegate to the generic service methods

func (s *UserServiceGeneric) CreateUser(ctx context.Context, user *entities.User) error {
	createReq := &framework.CreateRequest[entities.User]{
		Data: user,
		Options: &framework.CreateOptions{
			PublishEvents: true,
		},
	}

	_, err := s.Create(ctx, createReq)
	return err
}

func (s *UserServiceGeneric) UpdateUser(ctx context.Context, user *entities.User) error {
	updateReq := &framework.UpdateRequest[entities.User]{
		Data: user,
		Options: &framework.UpdateOptions{
			PublishEvents: true,
		},
	}

	_, err := s.Update(ctx, user.ID, updateReq)
	return err
}

func (s *UserServiceGeneric) SoftDeleteUser(ctx context.Context, id uint) error {
	return s.SoftDelete(ctx, id)
}

func (s *UserServiceGeneric) RestoreUser(ctx context.Context, id uint) error {
	return s.Restore(ctx, id)
}

func (s *UserServiceGeneric) HardDeleteUser(ctx context.Context, id uint) error {
	return s.HardDelete(ctx, id)
}

func (s *UserServiceGeneric) ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	req := &framework.ListRequest{Limit: limit, Offset: offset}
	response, err := s.List(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	return response.Data, response.Total, nil
}

func (s *UserServiceGeneric) ListDeletedUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	req := &framework.ListRequest{Limit: limit, Offset: offset}
	response, err := s.ListDeleted(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	return response.Data, response.Total, nil
}

func (s *UserServiceGeneric) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, int64, error) {
	req := &framework.ListRequest{Limit: limit, Offset: offset}
	response, err := s.Search(ctx, query, req)
	if err != nil {
		return nil, 0, err
	}
	return response.Data, response.Total, nil
}

func (s *UserServiceGeneric) BatchDeleteUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	response, err := s.BatchDelete(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Convert to legacy response format
	return &interfaces.BatchOperationResult{
		DeletedCount: response.SuccessCount,
		FailedIDs:    response.FailedIDs,
	}, nil
}

func (s *UserServiceGeneric) BatchRestoreUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	response, err := s.BatchRestore(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Convert to legacy response format
	return &interfaces.BatchOperationResult{
		RestoredCount: response.SuccessCount,
		FailedIDs:     response.FailedIDs,
	}, nil
}

// Override validation methods for user-specific business rules
func (s *UserServiceGeneric) ValidateCreate(ctx context.Context, req *framework.CreateRequest[entities.User]) error {
	if req == nil || req.Data == nil {
		return fmt.Errorf("create request and user data cannot be nil")
	}

	user := req.Data

	// Validate email
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}

	// Check email format
	if !strings.Contains(user.Email, "@") {
		return fmt.Errorf("invalid email format")
	}

	// Check for duplicate email
	existing, err := s.GetUserByEmail(ctx, user.Email)
	if err == nil && existing != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	// Call parent validation
	return s.BaseServiceImpl.ValidateCreate(ctx, req)
}

func (s *UserServiceGeneric) ValidateUpdate(ctx context.Context, id uint, req *framework.UpdateRequest[entities.User]) error {
	if req == nil || req.Data == nil {
		return fmt.Errorf("update request and user data cannot be nil")
	}

	user := req.Data

	// Validate email if being updated
	if user.Email != "" {
		if !strings.Contains(user.Email, "@") {
			return fmt.Errorf("invalid email format")
		}

		// Check for duplicate email (excluding current user)
		existing, err := s.GetUserByEmail(ctx, user.Email)
		if err == nil && existing != nil && existing.ID != id {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
	}

	// Call parent validation
	return s.BaseServiceImpl.ValidateUpdate(ctx, id, req)
}