package profile

import (
	"linke/internal/service"
)

// BaseProfileHandler provides common dependencies for all profile handlers
type BaseProfileHandler struct {
	UserService *service.UserService
}

// NewBaseProfileHandler creates a new base profile handler
func NewBaseProfileHandler(userService *service.UserService) *BaseProfileHandler {
	return &BaseProfileHandler{
		UserService: userService,
	}
}