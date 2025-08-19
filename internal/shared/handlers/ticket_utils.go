package handlers

import (
	"context"

	userEntities "linke/internal/domains/user/entities"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	sharedDTO "linke/internal/shared/dto"
	"linke/internal/shared/logger"
)

// TicketHandlerUtils provides common utilities for ticket handlers
type TicketHandlerUtils struct {
	userService userInterfaces.UserService
}

// NewTicketHandlerUtils creates a new ticket handler utilities instance
func NewTicketHandlerUtils(userService userInterfaces.UserService) *TicketHandlerUtils {
	return &TicketHandlerUtils{
		userService: userService,
	}
}

// ConvertUserToBasicDTO converts a user entity to UserBasicDTO
func ConvertUserToBasicDTO(user *userEntities.User) *sharedDTO.UserBasicDTO {
	if user == nil {
		return nil
	}

	return &sharedDTO.UserBasicDTO{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Name:     user.Name,
		Avatar:   user.Avatar,
		Provider: user.Provider,
		Status:   user.Status,
		Role:     user.Role,
	}
}

// BatchPopulateTicketUserData populates user data for a ticket response using batch loading
func (utils *TicketHandlerUtils) BatchPopulateTicketUserData(ctx context.Context, ticketResponse interface{}, userIDCollector *UserIDCollector) error {
	// This is a generic approach - in practice you'd have specific methods for each response type
	if userIDCollector.Count() == 0 {
		return nil
	}

	// Batch load all users
	userLoader := NewBatchUserLoader(utils.userService)
	if err := userLoader.LoadUsers(ctx, userIDCollector.ToSlice()); err != nil {
		logger.Error("Failed to batch load users for ticket response",
			logger.Int("user_count", userIDCollector.Count()),
			logger.ErrorField(err))
		return err
	}

	// The actual population would be handled by type assertion
	// This is just a framework - specific implementations would be in the handlers
	return nil
}

// ValidateAdminRole checks if a user has admin role
func ValidateAdminRole(user *userEntities.User) bool {
	return user != nil && user.Role == "admin"
}

// CreateUserLoader creates and configures a batch user loader for ticket operations
func (utils *TicketHandlerUtils) CreateUserLoader() *BatchUserLoader {
	loader := NewBatchUserLoader(utils.userService)
	loader.SetMaxCacheSize(500) // Optimize for ticket operations
	return loader
}

// PopulateUserData populates user data for a single user ID
func (utils *TicketHandlerUtils) PopulateUserData(ctx context.Context, userID uint) *sharedDTO.UserBasicDTO {
	user, err := utils.userService.GetUserByID(ctx, userID)
	if err != nil {
		logger.Warn("Failed to load user data",
			logger.Uint("user_id", userID),
			logger.ErrorField(err))
		return nil
	}
	return ConvertUserToBasicDTO(user)
}

// CreateOptimizedUserLoader creates a user loader with preload capabilities for common operations
func (utils *TicketHandlerUtils) CreateOptimizedUserLoader(ctx context.Context, preloadUserIDs []uint) *BatchUserLoader {
	loader := utils.CreateUserLoader()

	if len(preloadUserIDs) > 0 {
		loader.AddPreloadIDs(preloadUserIDs)
		// Trigger preload
		if err := loader.LoadUsers(ctx, []uint{}); err != nil {
			logger.Warn("Failed to preload users", logger.ErrorField(err))
		}
	}

	return loader
}
