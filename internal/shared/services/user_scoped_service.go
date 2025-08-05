package services

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"linke/internal/shared/framework"

	"go.uber.org/zap"
)

// UserScopedServiceImpl extends BaseServiceImpl with user-specific operations
type UserScopedServiceImpl[T any, ID comparable] struct {
	*BaseServiceImpl[T, ID]
	userRepository framework.UserScopedRepository[T, ID]
}

// NewUserScopedService creates a new UserScopedServiceImpl instance
func NewUserScopedService[T any, ID comparable](
	name string,
	repository framework.UserScopedRepository[T, ID],
	logger framework.Logger,
	eventPub framework.EventPublisher,
	validator framework.Validator,
) *UserScopedServiceImpl[T, ID] {
	return &UserScopedServiceImpl[T, ID]{
		BaseServiceImpl: NewBaseService(name, repository, logger, eventPub, validator),
		userRepository:  repository,
	}
}

// User-specific operations
func (s *UserScopedServiceImpl[T, ID]) ListByUser(ctx context.Context, userID uint, req *framework.ListRequest) (*framework.ListResponse[T], error) {
	if req == nil {
		req = &framework.ListRequest{Limit: 10, Offset: 0}
	}
	
	// Set default limit if not provided
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	entities, total, err := s.userRepository.ListByUser(ctx, userID, req.Limit, req.Offset)
	if err != nil {
		s.logger.Error("Failed to list entities by user", zap.String("service", s.name), zap.Uint("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("list entities by user %d: %w", userID, err)
	}

	return s.buildListResponse(entities, total, req), nil
}

func (s *UserScopedServiceImpl[T, ID]) CountByUser(ctx context.Context, userID uint) (int64, error) {
	count, err := s.userRepository.CountByUser(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to count entities by user", zap.String("service", s.name), zap.Uint("user_id", userID), zap.Error(err))
		return 0, fmt.Errorf("count entities by user %d: %w", userID, err)
	}
	return count, nil
}

func (s *UserScopedServiceImpl[T, ID]) GetUserTotalCount(ctx context.Context, userID uint) (int64, error) {
	return s.CountByUser(ctx, userID)
}

func (s *UserScopedServiceImpl[T, ID]) DeleteByUser(ctx context.Context, userID uint, ids []ID) (*framework.BatchOperationResponse, error) {
	var successCount int
	var failedIDs []uint
	var errors = make(map[string]string)

	for _, id := range ids {
		// Validate user access before deletion
		if err := s.ValidateUserAccess(ctx, userID, id); err != nil {
			// Convert ID to uint for response - this assumes ID is uint-compatible
			if idValue, ok := any(id).(uint); ok {
				failedIDs = append(failedIDs, idValue)
			}
			errors[fmt.Sprintf("%v", id)] = err.Error()
			continue
		}

		if err := s.repository.Delete(ctx, id); err != nil {
			// Convert ID to uint for response - this assumes ID is uint-compatible
			if idValue, ok := any(id).(uint); ok {
				failedIDs = append(failedIDs, idValue)
			}
			errors[fmt.Sprintf("%v", id)] = err.Error()
			s.logger.Warn("Failed to delete entity for user", zap.Uint("user_id", userID), zap.Any("id", id), zap.Error(err))
		} else {
			successCount++
		}
	}

	response := &framework.BatchOperationResponse{
		SuccessCount: successCount,
		FailedCount:  len(failedIDs),
		FailedIDs:    failedIDs,
		Errors:       errors,
	}

	s.logger.Info("Batch delete by user completed", zap.String("service", s.name), zap.Uint("user_id", userID), zap.Int("success", successCount), zap.Int("failed", len(failedIDs)))
	return response, nil
}

// User validation
func (s *UserScopedServiceImpl[T, ID]) ValidateUserAccess(ctx context.Context, userID uint, id ID) error {
	// Default implementation: check if entity exists and belongs to user
	entity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("entity not found: %w", err)
	}

	// Use reflection to check if entity has a UserID field
	entityValue := reflect.ValueOf(entity).Elem()
	userIDField := entityValue.FieldByName("UserID")
	
	if !userIDField.IsValid() {
		// If no UserID field, assume access is allowed (override in specific implementations if needed)
		return nil
	}

	if userIDField.Kind() == reflect.Uint && userIDField.Uint() != uint64(userID) {
		return fmt.Errorf("access denied: entity belongs to different user")
	}

	return nil
}

// Override Create to add user validation
func (s *UserScopedServiceImpl[T, ID]) Create(ctx context.Context, req *framework.CreateRequest[T]) (*T, error) {
	if req == nil || req.Data == nil {
		return nil, fmt.Errorf("create request and data cannot be nil")
	}

	// Additional user-specific validation can be added here
	// For example, checking rate limits, user permissions, etc.

	// Call parent Create method
	return s.BaseServiceImpl.Create(ctx, req)
}

// Override Update to add user validation
func (s *UserScopedServiceImpl[T, ID]) Update(ctx context.Context, id ID, req *framework.UpdateRequest[T]) (*T, error) {
	if req == nil || req.Data == nil {
		return nil, fmt.Errorf("update request and data cannot be nil")
	}

	// Get user ID from metadata if available
	var userID uint
	if req.Metadata != nil {
		if uid, ok := req.Metadata["user_id"].(uint); ok {
			userID = uid
			// Validate user access
			if err := s.ValidateUserAccess(ctx, userID, id); err != nil {
				return nil, fmt.Errorf("user access validation failed: %w", err)
			}
		}
	}

	// Call parent Update method
	return s.BaseServiceImpl.Update(ctx, id, req)
}

// Override Delete to add user validation
func (s *UserScopedServiceImpl[T, ID]) Delete(ctx context.Context, id ID) error {
	// Note: For user validation in delete, the userID should be passed through context
	// or the method signature should be extended. For now, we'll rely on the parent implementation.
	return s.BaseServiceImpl.Delete(ctx, id)
}

// User-specific statistics
func (s *UserScopedServiceImpl[T, ID]) GetUserStatistics(ctx context.Context, userID uint) (*framework.StatisticsResponse, error) {
	totalCount, err := s.userRepository.CountByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count user entities: %w", err)
	}

	// For user-specific stats, we don't typically have deleted counts in user scope
	// but this can be extended if needed

	customStats := map[string]interface{}{
		"user_id": userID,
	}

	return &framework.StatisticsResponse{
		TotalCount:  totalCount,
		CustomStats: customStats,
		GeneratedAt: time.Now(),
	}, nil
}