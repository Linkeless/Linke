package stubs

import (
	"context"
	"fmt"

	"linke/internal/domains/referral/dto"
	"linke/internal/domains/referral/entities"
	"linke/internal/domains/referral/usecases/interfaces"
)

// stubInviteCodeService provides a minimal stub implementation of InviteCodeService
// This is a temporary implementation to allow the application to start without the full referral system
type stubInviteCodeService struct{}

func NewStubInviteCodeService() interfaces.InviteCodeService {
	return &stubInviteCodeService{}
}

func (s *stubInviteCodeService) CreateInviteCode(ctx context.Context, createdByID uint, req *dto.CreateInviteCodeRequest) (*entities.InviteCode, error) {
	return nil, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetInviteCode(ctx context.Context, codeID uint) (*entities.InviteCode, error) {
	return nil, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetInviteCodeByCode(ctx context.Context, code string) (*entities.InviteCode, error) {
	return nil, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) UpdateInviteCode(ctx context.Context, inviteCodeID uint, req *dto.UpdateInviteCodeRequest) (*entities.InviteCode, error) {
	return nil, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) DeleteInviteCode(ctx context.Context, inviteCodeID uint) error {
	return fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetInviteCodes(ctx context.Context, req *dto.GetInviteCodesRequest) ([]*entities.InviteCode, int64, error) {
	return nil, 0, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetUserInviteCodes(ctx context.Context, userID uint, limit, offset int) ([]*entities.InviteCode, int64, error) {
	return nil, 0, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) ValidateInviteCode(ctx context.Context, code string) (*entities.InviteCode, error) {
	// Return nil to indicate the code is not valid (but don't fail startup)
	return nil, fmt.Errorf("invite code validation not available")
}

func (s *stubInviteCodeService) UseInviteCode(ctx context.Context, code string, userID uint, ipAddress, userAgent string) (*entities.InviteCodeUsage, error) {
	return nil, fmt.Errorf("invite code usage not implemented")
}

func (s *stubInviteCodeService) ActivateInviteCode(ctx context.Context, inviteCodeID uint) error {
	return fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) DeactivateInviteCode(ctx context.Context, inviteCodeID uint) error {
	return fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) ExpireInviteCode(ctx context.Context, inviteCodeID uint) error {
	return fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetInviteCodeUsage(ctx context.Context, inviteCodeID uint, limit, offset int) ([]*entities.InviteCodeUsage, int64, error) {
	return nil, 0, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetUserInviteCodeUsage(ctx context.Context, userID uint, limit, offset int) ([]*entities.InviteCodeUsage, int64, error) {
	return nil, 0, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetInviteCodeStatistics(ctx context.Context, inviteCodeID uint) (map[string]any, error) {
	return map[string]any{
		"total_codes":  0,
		"used_codes":   0,
		"active_codes": 0,
	}, nil
}

func (s *stubInviteCodeService) GetUserInviteCodeStatistics(ctx context.Context, userID uint) (map[string]any, error) {
	return map[string]any{
		"total_codes":  0,
		"used_codes":   0,
		"active_codes": 0,
	}, nil
}

func (s *stubInviteCodeService) GenerateInviteCode() (string, error) {
	return "", fmt.Errorf("invite code generation not implemented")
}
