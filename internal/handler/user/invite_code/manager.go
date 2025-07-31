package invitecode

import (
	invitecodemanagement "linke/internal/handler/user/invite_code/management"
	invitecodeoperation "linke/internal/handler/user/invite_code/operation"
	invitecodequery "linke/internal/handler/user/invite_code/query"
	invitecodestatistics "linke/internal/handler/user/invite_code/statistics"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// UserInviteCodeManager manages all user invite code-related operations with modular structure
type UserInviteCodeManager struct {
	// Sub-modules
	Management  *invitecodemanagement.InviteCodeManagementHandler
	Operation   *invitecodeoperation.InviteCodeValidationHandler
	Status      *invitecodeoperation.InviteCodeStatusHandler
	Query       *invitecodequery.InviteCodeListHandler
	Usage       *invitecodequery.InviteCodeUsageHandler
	Statistics  *invitecodestatistics.InviteCodeStatisticsHandler
}

// NewUserInviteCodeManager creates a new user invite code manager
func NewUserInviteCodeManager(inviteCodeService *service.InviteCodeService, inviteCodeUsageService *service.InviteCodeUsageService) *UserInviteCodeManager {
	return &UserInviteCodeManager{
		Management:  invitecodemanagement.NewInviteCodeManagementHandler(inviteCodeService, inviteCodeUsageService),
		Operation:   invitecodeoperation.NewInviteCodeValidationHandler(inviteCodeService, inviteCodeUsageService),
		Status:      invitecodeoperation.NewInviteCodeStatusHandler(inviteCodeService, inviteCodeUsageService),
		Query:       invitecodequery.NewInviteCodeListHandler(inviteCodeService, inviteCodeUsageService),
		Usage:       invitecodequery.NewInviteCodeUsageHandler(inviteCodeService, inviteCodeUsageService),
		Statistics:  invitecodestatistics.NewInviteCodeStatisticsHandler(inviteCodeService, inviteCodeUsageService),
	}
}

// ============= Compatibility Methods =============
// These methods provide backward compatibility with existing code

// CreateInviteCode provides backward compatibility for invite code creation
func (m *UserInviteCodeManager) CreateInviteCode(c *gin.Context) {
	m.Management.CreateInviteCode(c)
}

// GetInviteCode provides backward compatibility for invite code retrieval
func (m *UserInviteCodeManager) GetInviteCode(c *gin.Context) {
	m.Management.GetInviteCode(c)
}

// ValidateInviteCode provides backward compatibility for invite code validation
func (m *UserInviteCodeManager) ValidateInviteCode(c *gin.Context) {
	m.Operation.ValidateInviteCode(c)
}

// UpdateInviteCodeStatus provides backward compatibility for invite code status update
func (m *UserInviteCodeManager) UpdateInviteCodeStatus(c *gin.Context) {
	m.Status.UpdateInviteCodeStatus(c)
}

// DeleteInviteCode provides backward compatibility for invite code deletion
func (m *UserInviteCodeManager) DeleteInviteCode(c *gin.Context) {
	m.Management.DeleteInviteCode(c)
}

// GetInviteCodeStats provides backward compatibility for invite code statistics
func (m *UserInviteCodeManager) GetInviteCodeStats(c *gin.Context) {
	m.Statistics.GetInviteCodeStats(c)
}

// ListAllInviteCodes provides backward compatibility for listing all invite codes
func (m *UserInviteCodeManager) ListAllInviteCodes(c *gin.Context) {
	m.Query.ListAllInviteCodes(c)
}

// GetMyInviteCodes provides backward compatibility for getting user's invite codes
func (m *UserInviteCodeManager) GetMyInviteCodes(c *gin.Context) {
	m.Query.GetMyInviteCodes(c)
}

// GetInviteCodeUsages provides backward compatibility for getting invite code usages
func (m *UserInviteCodeManager) GetInviteCodeUsages(c *gin.Context) {
	m.Usage.GetInviteCodeUsages(c)
}