package implementations

import (
	"context"

	"linke/internal/domains/user/entities"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/events"
	"linke/internal/shared/framework"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// EventAwareUserService wraps UserService with event publishing capabilities
type EventAwareUserService struct {
	userService interfaces.UserService
	eventBus    events.EventBus
	logger      framework.Logger
}

// NewEventAwareUserService creates a new event-aware user service
func NewEventAwareUserService(
	userService interfaces.UserService,
	eventBus events.EventBus,
	logger framework.Logger,
) *EventAwareUserService {
	return &EventAwareUserService{
		userService: userService,
		eventBus:    eventBus,
		logger:      logger,
	}
}

// CreateUser creates a new user and publishes user creation events
func (s *EventAwareUserService) CreateUser(ctx context.Context, user *entities.User) error {
	// Create user first
	if err := s.userService.CreateUser(ctx, user); err != nil {
		return err
	}

	// Publish user created event
	userEvent := events.NewUserEvent(
		events.EventTypeUserCreated,
		user.ID,
		map[string]any{
			"user_id":    user.ID,
			"email":      user.Email,
			"name":       user.Name,
			"username":   user.Username,
			"provider":   user.Provider,
			"status":     user.Status,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		},
	)

	// Add correlation ID if available in context
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			events.SetCorrelationID(userEvent, id)
		}
	}

	if err := s.eventBus.Publish(ctx, userEvent); err != nil {
		s.logger.Error("Failed to publish user created event",
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err),
		)
		// Don't fail the operation if event publishing fails
	}

	s.logger.Info("User created and event published",
		logger.Uint("user_id", user.ID),
		logger.String("email", user.Email),
	)

	return nil
}

// UpdateUser updates a user and publishes user updated events
func (s *EventAwareUserService) UpdateUser(ctx context.Context, user *entities.User) error {
	// Get the original user to compare changes
	originalUser, err := s.userService.GetUserByID(ctx, user.ID)
	if err != nil {
		return err
	}

	// Update user
	if err := s.userService.UpdateUser(ctx, user); err != nil {
		return err
	}

	// Create event data with before/after comparison
	eventData := map[string]any{
		"user_id":    user.ID,
		"updated_at": user.UpdatedAt,
		"changes":    make(map[string]any),
	}

	// Track what changed
	changes := eventData["changes"].(map[string]any)
	if originalUser.Name != user.Name {
		changes["name"] = map[string]any{
			"from": originalUser.Name,
			"to":   user.Name,
		}
	}
	if originalUser.Email != user.Email {
		changes["email"] = map[string]any{
			"from": originalUser.Email,
			"to":   user.Email,
		}
	}
	if originalUser.Username != user.Username {
		changes["username"] = map[string]any{
			"from": originalUser.Username,
			"to":   user.Username,
		}
	}
	// Add more field comparisons as needed

	// Only publish event if there were actual changes
	if len(changes) > 0 {
		userEvent := events.NewUserEvent(
			events.EventTypeUserUpdated,
			user.ID,
			eventData,
		)

		if err := s.eventBus.Publish(ctx, userEvent); err != nil {
			s.logger.Error("Failed to publish user updated event",
				logger.Uint("user_id", user.ID),
				logger.ErrorField(err),
			)
		} else {
			s.logger.Info("User updated and event published",
				logger.Uint("user_id", user.ID),
				logger.Any("changes", changes),
			)
		}
	}

	return nil
}

// UpdateUserStatus updates a user's status and publishes status change events
func (s *EventAwareUserService) UpdateUserStatus(ctx context.Context, id uint, status string) (*entities.User, error) {
	// Get the original user to track the old status
	originalUser, err := s.userService.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update user status
	updatedUser, err := s.userService.UpdateUserStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}

	// Only publish event if status actually changed
	if originalUser.Status != status {
		userEvent := events.NewUserEvent(
			events.EventTypeUserStatusChanged,
			id,
			map[string]any{
				"user_id":    id,
				"old_status": originalUser.Status,
				"new_status": status,
				"changed_at": updatedUser.UpdatedAt,
			},
		)

		if err := s.eventBus.Publish(ctx, userEvent); err != nil {
			s.logger.Error("Failed to publish user status changed event",
				logger.Uint("user_id", id),
				logger.String("old_status", originalUser.Status),
				logger.String("new_status", status),
				logger.ErrorField(err),
			)
		} else {
			s.logger.Info("User status changed and event published",
				logger.Uint("user_id", id),
				logger.String("old_status", originalUser.Status),
				logger.String("new_status", status),
			)
		}
	}

	return updatedUser, nil
}

// SoftDeleteUser soft deletes a user and publishes user deleted events
func (s *EventAwareUserService) SoftDeleteUser(ctx context.Context, id uint) error {
	// Get user details before deletion for the event
	user, err := s.userService.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	// Soft delete user
	if err := s.userService.SoftDeleteUser(ctx, id); err != nil {
		return err
	}

	// Get active subscriptions for cascade operations
	// TODO: This would be replaced with actual subscription service call
	activeSubscriptions := []any{
		map[string]any{
			"id":     1, // Placeholder subscription ID
			"status": "active",
		},
	}

	// Publish user deleted event
	userEvent := events.NewUserEvent(
		events.EventTypeUserDeleted,
		id,
		map[string]any{
			"user_id":              id,
			"email":                user.Email,
			"name":                 user.Name,
			"deleted_at":           user.DeletedAt,
			"active_subscriptions": activeSubscriptions,
		},
	)

	if err := s.eventBus.Publish(ctx, userEvent); err != nil {
		s.logger.Error("Failed to publish user deleted event",
			logger.Uint("user_id", id),
			logger.ErrorField(err),
		)
		// Don't fail the operation if event publishing fails
	} else {
		s.logger.Info("User deleted and event published",
			logger.Uint("user_id", id),
			logger.String("email", user.Email),
		)
	}

	return nil
}

// RestoreUser restores a soft deleted user and publishes events
func (s *EventAwareUserService) RestoreUser(ctx context.Context, id uint) error {
	// Restore user
	if err := s.userService.RestoreUser(ctx, id); err != nil {
		return err
	}

	// Get restored user details
	user, err := s.userService.GetUserByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user details after restore",
			logger.Uint("user_id", id),
			logger.ErrorField(err),
		)
		return nil // Don't fail the restore operation
	}

	// Publish user created event (since restoration is like re-creation)
	userEvent := events.NewUserEvent(
		events.EventTypeUserCreated,
		id,
		map[string]any{
			"user_id":     id,
			"email":       user.Email,
			"name":        user.Name,
			"username":    user.Username,
			"provider":    user.Provider,
			"status":      user.Status,
			"role":        user.Role,
			"restored_at": user.UpdatedAt,
			"is_restored": true,
		},
	)

	if err := s.eventBus.Publish(ctx, userEvent); err != nil {
		s.logger.Error("Failed to publish user restored event",
			logger.Uint("user_id", id),
			logger.ErrorField(err),
		)
	} else {
		s.logger.Info("User restored and event published",
			logger.Uint("user_id", id),
			logger.String("email", user.Email),
		)
	}

	return nil
}

// Delegate all other methods to the wrapped service
func (s *EventAwareUserService) GetUserByID(ctx context.Context, id uint) (*entities.User, error) {
	return s.userService.GetUserByID(ctx, id)
}

func (s *EventAwareUserService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	return s.userService.GetUserByEmail(ctx, email)
}

func (s *EventAwareUserService) GetUserByTelegramID(ctx context.Context, telegramID string) (*entities.User, error) {
	return s.userService.GetUserByTelegramID(ctx, telegramID)
}

func (s *EventAwareUserService) GetActiveUserByID(ctx context.Context, id uint) (*entities.User, error) {
	return s.userService.GetActiveUserByID(ctx, id)
}

func (s *EventAwareUserService) GetActiveUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	return s.userService.GetActiveUserByEmail(ctx, email)
}

func (s *EventAwareUserService) GetUsersByIDs(ctx context.Context, ids []uint) ([]*entities.User, error) {
	return s.userService.GetUsersByIDs(ctx, ids)
}

func (s *EventAwareUserService) HardDeleteUser(ctx context.Context, id uint) error {
	// Get user details before hard deletion
	user, err := s.userService.GetUserByID(ctx, id)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// Hard delete user
	if err := s.userService.HardDeleteUser(ctx, id); err != nil {
		return err
	}

	// Publish hard delete event if user was found
	if user != nil {
		userEvent := events.NewUserEvent(
			events.EventTypeUserDeleted,
			id,
			map[string]any{
				"user_id":     id,
				"email":       user.Email,
				"name":        user.Name,
				"hard_delete": true,
				"deleted_at":  user.DeletedAt,
			},
		)

		if err := s.eventBus.Publish(ctx, userEvent); err != nil {
			s.logger.Error("Failed to publish user hard deleted event",
				logger.Uint("user_id", id),
				logger.ErrorField(err),
			)
		} else {
			s.logger.Info("User hard deleted and event published",
				logger.Uint("user_id", id),
				logger.String("email", user.Email),
			)
		}
	}

	return nil
}

func (s *EventAwareUserService) ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	return s.userService.ListUsers(ctx, limit, offset)
}

func (s *EventAwareUserService) ListDeletedUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	return s.userService.ListDeletedUsers(ctx, limit, offset)
}

func (s *EventAwareUserService) ListUsersByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error) {
	return s.userService.ListUsersByProvider(ctx, provider, limit, offset)
}

func (s *EventAwareUserService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, int64, error) {
	return s.userService.SearchUsers(ctx, query, limit, offset)
}

func (s *EventAwareUserService) UpdateUserRole(ctx context.Context, id uint, role string) (*entities.User, error) {
	// Get the original user to track the old role
	originalUser, err := s.userService.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update user role
	updatedUser, err := s.userService.UpdateUserRole(ctx, id, role)
	if err != nil {
		return nil, err
	}

	// Only publish event if role actually changed
	if originalUser.Role != role {
		userEvent := events.NewUserEvent(
			events.EventTypeUserUpdated,
			id,
			map[string]any{
				"user_id":     id,
				"old_role":    originalUser.Role,
				"new_role":    role,
				"changed_at":  updatedUser.UpdatedAt,
				"change_type": "role_update",
			},
		)

		if err := s.eventBus.Publish(ctx, userEvent); err != nil {
			s.logger.Error("Failed to publish user role changed event",
				logger.Uint("user_id", id),
				logger.String("old_role", originalUser.Role),
				logger.String("new_role", role),
				logger.ErrorField(err),
			)
		} else {
			s.logger.Info("User role changed and event published",
				logger.Uint("user_id", id),
				logger.String("old_role", originalUser.Role),
				logger.String("new_role", role),
			)
		}
	}

	return updatedUser, nil
}

func (s *EventAwareUserService) GetUserStats(ctx context.Context) (*interfaces.UserStats, error) {
	return s.userService.GetUserStats(ctx)
}

func (s *EventAwareUserService) BatchDeleteUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	// Get user details before batch deletion
	users := make(map[uint]*entities.User)
	for _, id := range ids {
		if user, err := s.userService.GetUserByID(ctx, id); err == nil {
			users[id] = user
		}
	}

	// Perform batch deletion
	result, err := s.userService.BatchDeleteUsers(ctx, ids)
	if err != nil {
		return result, err
	}

	// Publish events for successfully deleted users
	for _, id := range ids {
		// Skip if this ID was in the failed list
		failed := false
		for _, failedID := range result.FailedIDs {
			if failedID == id {
				failed = true
				break
			}
		}

		if !failed && users[id] != nil {
			user := users[id]
			userEvent := events.NewUserEvent(
				events.EventTypeUserDeleted,
				id,
				map[string]any{
					"user_id":      id,
					"email":        user.Email,
					"name":         user.Name,
					"batch_delete": true,
					"deleted_at":   user.DeletedAt,
				},
			)

			if err := s.eventBus.Publish(ctx, userEvent); err != nil {
				s.logger.Error("Failed to publish batch user deleted event",
					logger.Uint("user_id", id),
					logger.ErrorField(err),
				)
			}
		}
	}

	s.logger.Info("Batch user deletion completed with events",
		logger.Int("deleted_count", result.DeletedCount),
		logger.Int("failed_count", len(result.FailedIDs)),
	)

	return result, nil
}

func (s *EventAwareUserService) BatchRestoreUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	// Perform batch restoration
	result, err := s.userService.BatchRestoreUsers(ctx, ids)
	if err != nil {
		return result, err
	}

	// Publish events for successfully restored users
	for _, id := range ids {
		// Skip if this ID was in the failed list
		failed := false
		for _, failedID := range result.FailedIDs {
			if failedID == id {
				failed = true
				break
			}
		}

		if !failed {
			// Get restored user details
			if user, err := s.userService.GetUserByID(ctx, id); err == nil {
				userEvent := events.NewUserEvent(
					events.EventTypeUserCreated,
					id,
					map[string]any{
						"user_id":       id,
						"email":         user.Email,
						"name":          user.Name,
						"username":      user.Username,
						"provider":      user.Provider,
						"status":        user.Status,
						"role":          user.Role,
						"batch_restore": true,
						"restored_at":   user.UpdatedAt,
					},
				)

				if err := s.eventBus.Publish(ctx, userEvent); err != nil {
					s.logger.Error("Failed to publish batch user restored event",
						logger.Uint("user_id", id),
						logger.ErrorField(err),
					)
				}
			}
		}
	}

	s.logger.Info("Batch user restoration completed with events",
		logger.Int("restored_count", result.RestoredCount),
		logger.Int("failed_count", len(result.FailedIDs)),
	)

	return result, nil
}
