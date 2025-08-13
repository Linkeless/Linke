package interfaces

import (
	"context"
	"linke/internal/domains/referral/dto"
	"linke/internal/domains/referral/entities"
)

// InviteCodeService defines the interface for invite code operations
type InviteCodeService interface {
	// Invite code CRUD operations
	CreateInviteCode(ctx context.Context, createdByID uint, req *dto.CreateInviteCodeRequest) (*entities.InviteCode, error)
	GetInviteCode(ctx context.Context, inviteCodeID uint) (*entities.InviteCode, error)
	GetInviteCodeByCode(ctx context.Context, code string) (*entities.InviteCode, error)
	UpdateInviteCode(ctx context.Context, inviteCodeID uint, req *dto.UpdateInviteCodeRequest) (*entities.InviteCode, error)
	DeleteInviteCode(ctx context.Context, inviteCodeID uint) error

	// Invite code listing and filtering
	GetInviteCodes(ctx context.Context, req *dto.GetInviteCodesRequest) ([]*entities.InviteCode, int64, error)
	GetUserInviteCodes(ctx context.Context, userID uint, limit, offset int) ([]*entities.InviteCode, int64, error)

	// Invite code validation and usage
	ValidateInviteCode(ctx context.Context, code string) (*entities.InviteCode, error)
	UseInviteCode(ctx context.Context, code string, userID uint, ipAddress, userAgent string) (*entities.InviteCodeUsage, error)

	// Invite code management
	ActivateInviteCode(ctx context.Context, inviteCodeID uint) error
	DeactivateInviteCode(ctx context.Context, inviteCodeID uint) error
	ExpireInviteCode(ctx context.Context, inviteCodeID uint) error

	// Usage tracking
	GetInviteCodeUsage(ctx context.Context, inviteCodeID uint, limit, offset int) ([]*entities.InviteCodeUsage, int64, error)
	GetUserInviteCodeUsage(ctx context.Context, userID uint, limit, offset int) ([]*entities.InviteCodeUsage, int64, error)

	// Invite code statistics
	GetInviteCodeStatistics(ctx context.Context, inviteCodeID uint) (map[string]any, error)
	GetUserInviteCodeStatistics(ctx context.Context, userID uint) (map[string]any, error)

	// Utility
	GenerateInviteCode() (string, error)
}
