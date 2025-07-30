package shared

import (
	"linke/internal/service"
)

// BaseHandler provides common dependencies for user handlers
type BaseHandler struct {
	UserService *service.UserService
	AuthService *service.AuthService
	Validator   *UserValidator
}

// NewBaseHandler creates a new base handler with common dependencies
func NewBaseHandler(userService *service.UserService, authService *service.AuthService) *BaseHandler {
	return &BaseHandler{
		UserService: userService,
		AuthService: authService,
		Validator:   NewUserValidator(),
	}
}