package persistence

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"linke/internal/user/domain/aggregate"
	"linke/internal/user/domain/repository"
	"linke/internal/user/domain/valueobject"
)

// UserGormRepository implements UserRepository using GORM
type UserGormRepository struct {
	db     *gorm.DB
	mapper *UserMapper
}

// NewUserGormRepository creates a new UserGormRepository
func NewUserGormRepository(db *gorm.DB) repository.UserRepository {
	return &UserGormRepository{
		db:     db,
		mapper: NewUserMapper(),
	}
}

// NewUserGormReadRepository creates a new read-only UserRepository
func NewUserGormReadRepository(db *gorm.DB) repository.UserReadRepository {
	return &UserGormRepository{
		db:     db,
		mapper: NewUserMapper(),
	}
}

// Core CRUD operations

// Save saves a user to the database
func (r *UserGormRepository) Save(ctx context.Context, user *aggregate.User) error {
	po := r.mapper.ToPersistence(user)
	if po == nil {
		return errors.New("failed to convert user to persistence object")
	}

	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

// FindByID finds a user by ID
func (r *UserGormRepository) FindByID(ctx context.Context, id valueobject.UserID) (*aggregate.User, error) {
	var po UserPO
	
	// Try to find by numeric ID first (for legacy compatibility)
	if id.ToUint() > 0 {
		if err := r.db.WithContext(ctx).First(&po, id.ToUint()).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("user not found")
			}
			return nil, fmt.Errorf("failed to find user: %w", err)
		}
	} else {
		// For string-based IDs, we'd need to implement this differently
		// For now, return not found
		return nil, fmt.Errorf("user not found")
	}

	user, err := r.mapper.ToDomain(&po)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain model: %w", err)
	}

	return user, nil
}

// FindByEmail finds a user by email
func (r *UserGormRepository) FindByEmail(ctx context.Context, email valueobject.Email) (*aggregate.User, error) {
	var po UserPO
	
	if err := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email.String()).First(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	user, err := r.mapper.ToDomain(&po)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain model: %w", err)
	}

	return user, nil
}

// Update updates a user in the database
func (r *UserGormRepository) Update(ctx context.Context, user *aggregate.User) error {
	po := r.mapper.ToPersistence(user)
	if po == nil {
		return errors.New("failed to convert user to persistence object")
	}

	if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete hard deletes a user from the database
func (r *UserGormRepository) Delete(ctx context.Context, id valueobject.UserID) error {
	if err := r.db.WithContext(ctx).Unscoped().Delete(&UserPO{}, id.ToUint()).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// Query operations

// ExistsByEmail checks if a user exists by email
func (r *UserGormRepository) ExistsByEmail(ctx context.Context, email valueobject.Email) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&UserPO{}).Where("email = ? AND deleted_at IS NULL", email.String()).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check if user exists by email: %w", err)
	}

	return count > 0, nil
}

// ExistsByUsername checks if a user exists by username
func (r *UserGormRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&UserPO{}).Where("username = ? AND deleted_at IS NULL", username).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check if user exists by username: %w", err)
	}

	return count > 0, nil
}

// FindByUsername finds a user by username
func (r *UserGormRepository) FindByUsername(ctx context.Context, username string) (*aggregate.User, error) {
	var po UserPO
	
	if err := r.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}

	user, err := r.mapper.ToDomain(&po)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain model: %w", err)
	}

	return user, nil
}

// OAuth-specific operations

// FindByProviderID finds a user by provider ID
func (r *UserGormRepository) FindByProviderID(ctx context.Context, provider string, providerID valueobject.ProviderID) (*aggregate.User, error) {
	var po UserPO
	var query *gorm.DB

	switch provider {
	case "google":
		query = r.db.WithContext(ctx).Where("google_id = ? AND deleted_at IS NULL", providerID.String())
	case "github":
		query = r.db.WithContext(ctx).Where("github_id = ? AND deleted_at IS NULL", providerID.String())
	case "telegram":
		query = r.db.WithContext(ctx).Where("telegram_id = ? AND deleted_at IS NULL", providerID.String())
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	if err := query.First(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by provider ID: %w", err)
	}

	user, err := r.mapper.ToDomain(&po)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain model: %w", err)
	}

	return user, nil
}

// ExistsByProviderID checks if a user exists by provider ID
func (r *UserGormRepository) ExistsByProviderID(ctx context.Context, provider string, providerID valueobject.ProviderID) (bool, error) {
	var count int64
	var query *gorm.DB

	switch provider {
	case "google":
		query = r.db.WithContext(ctx).Model(&UserPO{}).Where("google_id = ? AND deleted_at IS NULL", providerID.String())
	case "github":
		query = r.db.WithContext(ctx).Model(&UserPO{}).Where("github_id = ? AND deleted_at IS NULL", providerID.String())
	case "telegram":
		query = r.db.WithContext(ctx).Model(&UserPO{}).Where("telegram_id = ? AND deleted_at IS NULL", providerID.String())
	default:
		return false, fmt.Errorf("unsupported provider: %s", provider)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check if user exists by provider ID: %w", err)
	}

	return count > 0, nil
}

// List operations with pagination

// FindAll finds all users with pagination
func (r *UserGormRepository) FindAll(ctx context.Context, offset, limit int) ([]*aggregate.User, error) {
	var pos []UserPO
	
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Offset(offset).Limit(limit).Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}

	users, err := r.mapper.ToDomainList(convertToPointerSlice(pos))
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain models: %w", err)
	}

	return users, nil
}

// FindByStatus finds users by status with pagination
func (r *UserGormRepository) FindByStatus(ctx context.Context, status valueobject.UserStatus, offset, limit int) ([]*aggregate.User, error) {
	var pos []UserPO
	
	if err := r.db.WithContext(ctx).Where("status = ? AND deleted_at IS NULL", status.String()).Offset(offset).Limit(limit).Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find users by status: %w", err)
	}

	users, err := r.mapper.ToDomainList(convertToPointerSlice(pos))
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain models: %w", err)
	}

	return users, nil
}

// FindByRole finds users by role with pagination
func (r *UserGormRepository) FindByRole(ctx context.Context, role valueobject.UserRole, offset, limit int) ([]*aggregate.User, error) {
	var pos []UserPO
	
	if err := r.db.WithContext(ctx).Where("role = ? AND deleted_at IS NULL", role.String()).Offset(offset).Limit(limit).Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find users by role: %w", err)
	}

	users, err := r.mapper.ToDomainList(convertToPointerSlice(pos))
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain models: %w", err)
	}

	return users, nil
}

// FindByProvider finds users by provider with pagination
func (r *UserGormRepository) FindByProvider(ctx context.Context, provider valueobject.Provider, offset, limit int) ([]*aggregate.User, error) {
	var pos []UserPO
	
	if err := r.db.WithContext(ctx).Where("provider = ? AND deleted_at IS NULL", provider.String()).Offset(offset).Limit(limit).Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find users by provider: %w", err)
	}

	users, err := r.mapper.ToDomainList(convertToPointerSlice(pos))
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain models: %w", err)
	}

	return users, nil
}

// Count operations

// Count counts all users
func (r *UserGormRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&UserPO{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// CountByStatus counts users by status
func (r *UserGormRepository) CountByStatus(ctx context.Context, status valueobject.UserStatus) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&UserPO{}).Where("status = ? AND deleted_at IS NULL", status.String()).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users by status: %w", err)
	}

	return count, nil
}

// CountByRole counts users by role
func (r *UserGormRepository) CountByRole(ctx context.Context, role valueobject.UserRole) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&UserPO{}).Where("role = ? AND deleted_at IS NULL", role.String()).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users by role: %w", err)
	}

	return count, nil
}

// CountByProvider counts users by provider
func (r *UserGormRepository) CountByProvider(ctx context.Context, provider valueobject.Provider) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&UserPO{}).Where("provider = ? AND deleted_at IS NULL", provider.String()).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users by provider: %w", err)
	}

	return count, nil
}

// Search operations

// SearchByEmailOrUsername searches users by email or username
func (r *UserGormRepository) SearchByEmailOrUsername(ctx context.Context, query string, offset, limit int) ([]*aggregate.User, error) {
	var pos []UserPO
	searchQuery := "%" + query + "%"
	
	if err := r.db.WithContext(ctx).
		Where("(email LIKE ? OR username LIKE ?) AND deleted_at IS NULL", searchQuery, searchQuery).
		Offset(offset).Limit(limit).Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	users, err := r.mapper.ToDomainList(convertToPointerSlice(pos))
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain models: %w", err)
	}

	return users, nil
}

// Soft delete operations

// SoftDelete performs soft delete on a user
func (r *UserGormRepository) SoftDelete(ctx context.Context, id valueobject.UserID) error {
	if err := r.db.WithContext(ctx).Delete(&UserPO{}, id.ToUint()).Error; err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	return nil
}

// Restore restores a soft deleted user
func (r *UserGormRepository) Restore(ctx context.Context, id valueobject.UserID) error {
	if err := r.db.WithContext(ctx).Unscoped().Model(&UserPO{}).Where("id = ?", id.ToUint()).Update("deleted_at", nil).Error; err != nil {
		return fmt.Errorf("failed to restore user: %w", err)
	}

	return nil
}

// FindDeleted finds deleted users with pagination
func (r *UserGormRepository) FindDeleted(ctx context.Context, offset, limit int) ([]*aggregate.User, error) {
	var pos []UserPO
	
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Offset(offset).Limit(limit).Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find deleted users: %w", err)
	}

	users, err := r.mapper.ToDomainList(convertToPointerSlice(pos))
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain models: %w", err)
	}

	return users, nil
}

// Batch operations

// SaveBatch saves multiple users in a single transaction
func (r *UserGormRepository) SaveBatch(ctx context.Context, users []*aggregate.User) error {
	if len(users) == 0 {
		return nil
	}

	pos := r.mapper.ToPersistenceList(users)
	if len(pos) == 0 {
		return errors.New("failed to convert users to persistence objects")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, po := range pos {
			if err := tx.Create(po).Error; err != nil {
				return fmt.Errorf("failed to save user batch: %w", err)
			}
		}
		return nil
	})
}

// FindByIDs finds users by multiple IDs
func (r *UserGormRepository) FindByIDs(ctx context.Context, ids []valueobject.UserID) ([]*aggregate.User, error) {
	if len(ids) == 0 {
		return []*aggregate.User{}, nil
	}

	// Convert UserIDs to uint slice
	uintIDs := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id.ToUint() > 0 {
			uintIDs = append(uintIDs, id.ToUint())
		}
	}

	if len(uintIDs) == 0 {
		return []*aggregate.User{}, nil
	}

	var pos []UserPO
	if err := r.db.WithContext(ctx).Where("id IN ? AND deleted_at IS NULL", uintIDs).Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find users by IDs: %w", err)
	}

	users, err := r.mapper.ToDomainList(convertToPointerSlice(pos))
	if err != nil {
		return nil, fmt.Errorf("failed to convert to domain models: %w", err)
	}

	return users, nil
}

// Helper functions

// convertToPointerSlice converts slice of UserPO to slice of *UserPO
func convertToPointerSlice(pos []UserPO) []*UserPO {
	result := make([]*UserPO, len(pos))
	for i := range pos {
		result[i] = &pos[i]
	}
	return result
}