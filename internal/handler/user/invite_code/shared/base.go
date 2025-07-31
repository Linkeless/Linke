package invitecode

import (
	"linke/internal/service"
)

// BaseInviteCodeHandler provides common dependencies for all invite code handlers
type BaseInviteCodeHandler struct {
	InviteCodeService      *service.InviteCodeService
	InviteCodeUsageService *service.InviteCodeUsageService
}

// NewBaseInviteCodeHandler creates a new base invite code handler
func NewBaseInviteCodeHandler(inviteCodeService *service.InviteCodeService, inviteCodeUsageService *service.InviteCodeUsageService) *BaseInviteCodeHandler {
	return &BaseInviteCodeHandler{
		InviteCodeService:      inviteCodeService,
		InviteCodeUsageService: inviteCodeUsageService,
	}
}