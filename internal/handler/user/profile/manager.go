package profile

import (
	profilemanagement "linke/internal/handler/user/profile/management"
	profileoperation "linke/internal/handler/user/profile/operation"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// UserProfileManager manages all user profile-related operations with modular structure
type UserProfileManager struct {
	// Sub-modules
	Management *profilemanagement.ProfileManagementHandler
	Operation  *profileoperation.ProfileOperationHandler
}

// NewUserProfileManager creates a new user profile manager
func NewUserProfileManager(userService *service.UserService) *UserProfileManager {
	return &UserProfileManager{
		Management: profilemanagement.NewProfileManagementHandler(userService),
		Operation:  profileoperation.NewProfileOperationHandler(userService),
	}
}

// ============= Compatibility Methods =============
// These methods provide backward compatibility with existing code

// GetProfile provides backward compatibility for profile retrieval
func (m *UserProfileManager) GetProfile(c *gin.Context) {
	m.Management.GetProfile(c)
}

// UpdateProfile provides backward compatibility for profile update
func (m *UserProfileManager) UpdateProfile(c *gin.Context) {
	m.Management.UpdateProfile(c)
}

// ChangePassword provides backward compatibility for password change
func (m *UserProfileManager) ChangePassword(c *gin.Context) {
	m.Operation.ChangePassword(c)
}