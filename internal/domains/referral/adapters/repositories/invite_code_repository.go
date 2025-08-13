package repositories

import (
	"context"
	"linke/internal/domains/referral/entities"
	"linke/internal/domains/referral/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/repository"

	"gorm.io/gorm"
)

// InviteCodeRepository implements the InviteCodeRepository interface
type InviteCodeRepository struct {
	*repository.UserScopedRepositoryImpl[entities.InviteCode, uint]
}

// NewInviteCodeRepository creates a new invite code repository
func NewInviteCodeRepository(db *gorm.DB, frameworkLogger framework.Logger) interfaces.InviteCodeRepository {
	return &InviteCodeRepository{
		UserScopedRepositoryImpl: repository.NewUserScopedRepository[entities.InviteCode, uint](db, frameworkLogger),
	}
}

// GetByID gets an invite code by ID (override base implementation for custom logic)
func (r *InviteCodeRepository) GetByID(ctx context.Context, inviteCodeID uint) (*entities.InviteCode, error) {
	return r.UserScopedRepositoryImpl.GetByID(ctx, inviteCodeID)
}

// GetByCode gets an invite code by code string
func (r *InviteCodeRepository) GetByCode(ctx context.Context, code string) (*entities.InviteCode, error) {
	var inviteCode entities.InviteCode
	if err := r.GetDB().WithContext(ctx).Where("code = ?", code).First(&inviteCode).Error; err != nil {
		return nil, err
	}
	return &inviteCode, nil
}

// Update updates an invite code
func (r *InviteCodeRepository) Update(ctx context.Context, inviteCode *entities.InviteCode) error {
	return r.UserScopedRepositoryImpl.Update(ctx, inviteCode)
}

// Delete deletes an invite code
func (r *InviteCodeRepository) Delete(ctx context.Context, inviteCodeID uint) error {
	return r.UserScopedRepositoryImpl.Delete(ctx, inviteCodeID)
}

// List lists invite codes (implements GenericRepository interface)
func (r *InviteCodeRepository) List(ctx context.Context, limit, offset int) ([]*entities.InviteCode, int64, error) {
	return r.UserScopedRepositoryImpl.List(ctx, limit, offset)
}

// ListWithFilters lists invite codes with filters
func (r *InviteCodeRepository) ListWithFilters(ctx context.Context, filters map[string]any, limit, offset int) ([]*entities.InviteCode, int64, error) {
	var inviteCodes []*entities.InviteCode
	var total int64

	query := r.GetDB().WithContext(ctx).Model(&entities.InviteCode{})

	// Apply filters
	if userID, ok := filters["user_id"]; ok && userID != uint(0) {
		query = query.Where("user_id = ?", userID)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if campaignID, ok := filters["campaign_id"]; ok && campaignID != nil {
		query = query.Where("campaign_id = ?", campaignID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get results with pagination
	if err := query.Limit(limit).Offset(offset).Find(&inviteCodes).Error; err != nil {
		return nil, 0, err
	}

	return inviteCodes, total, nil
}

// GetUserInviteCodes gets invite codes for a specific user
func (r *InviteCodeRepository) GetUserInviteCodes(ctx context.Context, userID uint, limit, offset int) ([]*entities.InviteCode, int64, error) {
	return r.UserScopedRepositoryImpl.ListByUser(ctx, userID, limit, offset)
}

// BatchDelete deletes multiple invite codes by IDs
func (r *InviteCodeRepository) BatchDelete(ctx context.Context, inviteCodeIDs []uint) (int, []uint, error) {
	return r.UserScopedRepositoryImpl.BatchDelete(ctx, inviteCodeIDs)
}

// BatchRestore restores multiple soft-deleted invite codes by IDs
func (r *InviteCodeRepository) BatchRestore(ctx context.Context, inviteCodeIDs []uint) (int, []uint, error) {
	return r.UserScopedRepositoryImpl.BatchRestore(ctx, inviteCodeIDs)
}

// BatchUpdateStatus updates status for multiple invite codes by IDs
func (r *InviteCodeRepository) BatchUpdateStatus(ctx context.Context, inviteCodeIDs []uint, status string) (int, []uint, error) {
	return r.UserScopedRepositoryImpl.BatchUpdateStatus(ctx, inviteCodeIDs, status)
}

// ExistsByCode checks if an invite code exists by code string
func (r *InviteCodeRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.InviteCode{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateUsageCount updates the usage count of an invite code
func (r *InviteCodeRepository) UpdateUsageCount(ctx context.Context, id uint, usedCount int) error {
	return r.GetDB().WithContext(ctx).Model(&entities.InviteCode{}).Where("id = ?", id).Update("used_count", usedCount).Error
}

// IncrementUsageCount increments the usage count of an invite code
func (r *InviteCodeRepository) IncrementUsageCount(ctx context.Context, id uint) error {
	return r.GetDB().WithContext(ctx).Model(&entities.InviteCode{}).Where("id = ?", id).Update("used_count", gorm.Expr("used_count + 1")).Error
}

// ListByCreator lists invite codes by creator
func (r *InviteCodeRepository) ListByCreator(ctx context.Context, creatorID uint, limit, offset int) ([]*entities.InviteCode, int64, error) {
	return r.UserScopedRepositoryImpl.ListByUser(ctx, creatorID, limit, offset)
}

// ListActive lists active invite codes
func (r *InviteCodeRepository) ListActive(ctx context.Context, limit, offset int) ([]*entities.InviteCode, int64, error) {
	return r.UserScopedRepositoryImpl.ListByStatus(ctx, "active", limit, offset)
}

// ListAvailable lists available invite codes
func (r *InviteCodeRepository) ListAvailable(ctx context.Context, limit, offset int) ([]*entities.InviteCode, int64, error) {
	var inviteCodes []*entities.InviteCode
	var total int64

	query := r.GetDB().WithContext(ctx).Where("status = ? AND (max_uses = 0 OR used_count < max_uses)", "active")

	if err := query.Model(&entities.InviteCode{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&inviteCodes).Error; err != nil {
		return nil, 0, err
	}

	return inviteCodes, total, nil
}

// ListExhausted lists exhausted invite codes
func (r *InviteCodeRepository) ListExhausted(ctx context.Context, limit, offset int) ([]*entities.InviteCode, int64, error) {
	var inviteCodes []*entities.InviteCode
	var total int64

	query := r.GetDB().WithContext(ctx).Where("max_uses > 0 AND used_count >= max_uses")

	if err := query.Model(&entities.InviteCode{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&inviteCodes).Error; err != nil {
		return nil, 0, err
	}

	return inviteCodes, total, nil
}

// CountByCreator counts invite codes by creator
func (r *InviteCodeRepository) CountByCreator(ctx context.Context, creatorID uint) (int64, error) {
	return r.UserScopedRepositoryImpl.CountByUser(ctx, creatorID)
}
