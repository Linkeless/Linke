package user

import (
	"linke/internal/handler/admin/user/management"
	"linke/internal/handler/admin/user/operation"
	"linke/internal/handler/admin/user/query"
	"linke/internal/handler/admin/user/statistics"
	"linke/internal/handler/admin/user/status"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminUserManager manages all user-related admin handlers
type AdminUserManager struct {
	// Sub-handlers for different user management aspects
	Management *management.UserCRUDHandler
	Status     *status.UserStatusHandler
	Query      *query.UserListHandler
	Search     *query.UserSearchHandler
	Delete     *operation.UserDeleteHandler
	Batch      *operation.UserBatchHandler
	Stats      *statistics.UserStatsHandler
}

// NewAdminUserManager creates a new admin user manager with all sub-handlers
func NewAdminUserManager(userService *service.UserService, authService *service.AuthService) *AdminUserManager {
	return &AdminUserManager{
		Management: management.NewUserCRUDHandler(userService, authService),
		Status:     status.NewUserStatusHandler(userService, authService),
		Query:      query.NewUserListHandler(userService, authService),
		Search:     query.NewUserSearchHandler(userService, authService),
		Delete:     operation.NewUserDeleteHandler(userService, authService),
		Batch:      operation.NewUserBatchHandler(userService, authService),
		Stats:      statistics.NewUserStatsHandler(userService, authService),
	}
}

// Legacy compatibility layer - maintains the same interface as the original AdminUserHandler
// This allows existing code to continue working without changes while using the modular structure internally

// CreateUser delegates to the management module
func (m *AdminUserManager) CreateUser(c *gin.Context) {
	m.Management.CreateUser(c)
}

// GetUser delegates to the management module
func (m *AdminUserManager) GetUser(c *gin.Context) {
	m.Management.GetUser(c)
}

// UpdateUser delegates to the management module
func (m *AdminUserManager) UpdateUser(c *gin.Context) {
	m.Management.UpdateUser(c)
}

// PatchUser delegates to the management module
func (m *AdminUserManager) PatchUser(c *gin.Context) {
	m.Management.PatchUser(c)
}

// UpdateUserRole delegates to the status module
func (m *AdminUserManager) UpdateUserRole(c *gin.Context) {
	m.Status.UpdateUserRole(c)
}

// UpdateUserStatus delegates to the status module
func (m *AdminUserManager) UpdateUserStatus(c *gin.Context) {
	m.Status.UpdateUserStatus(c)
}

// ResetUserPassword delegates to the status module
func (m *AdminUserManager) ResetUserPassword(c *gin.Context) {
	m.Status.ResetUserPassword(c)
}

// ListUsers delegates to the query module
func (m *AdminUserManager) ListUsers(c *gin.Context) {
	m.Query.ListUsers(c)
}

// ListDeletedUsers delegates to the query module
func (m *AdminUserManager) ListDeletedUsers(c *gin.Context) {
	m.Query.ListDeletedUsers(c)
}

// ListUsersByProvider delegates to the query module
func (m *AdminUserManager) ListUsersByProvider(c *gin.Context) {
	m.Query.ListUsersByProvider(c)
}

// SearchUsers delegates to the search module
func (m *AdminUserManager) SearchUsers(c *gin.Context) {
	m.Search.SearchUsers(c)
}

// SoftDeleteUser delegates to the delete module
func (m *AdminUserManager) SoftDeleteUser(c *gin.Context) {
	m.Delete.SoftDeleteUser(c)
}

// RestoreUser delegates to the delete module
func (m *AdminUserManager) RestoreUser(c *gin.Context) {
	m.Delete.RestoreUser(c)
}

// HardDeleteUser delegates to the delete module
func (m *AdminUserManager) HardDeleteUser(c *gin.Context) {
	m.Delete.HardDeleteUser(c)
}

// BatchDeleteUsers delegates to the batch module
func (m *AdminUserManager) BatchDeleteUsers(c *gin.Context) {
	m.Batch.BatchDeleteUsers(c)
}

// BatchRestoreUsers delegates to the batch module
func (m *AdminUserManager) BatchRestoreUsers(c *gin.Context) {
	m.Batch.BatchRestoreUsers(c)
}

// GetUserStats delegates to the stats module
func (m *AdminUserManager) GetUserStats(c *gin.Context) {
	m.Stats.GetUserStats(c)
}