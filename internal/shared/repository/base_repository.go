package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"linke/internal/shared/framework"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BaseRepositoryImpl provides a concrete implementation of GenericRepository
type BaseRepositoryImpl[T any, ID comparable] struct {
	db     *gorm.DB
	logger framework.Logger
}

// NewBaseRepository creates a new BaseRepositoryImpl instance
func NewBaseRepository[T any, ID comparable](db *gorm.DB, logger framework.Logger) *BaseRepositoryImpl[T, ID] {
	return &BaseRepositoryImpl[T, ID]{
		db:     db,
		logger: logger,
	}
}

// Repository interface methods
func (r *BaseRepositoryImpl[T, ID]) GetDB() *gorm.DB {
	return r.db
}

func (r *BaseRepositoryImpl[T, ID]) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *BaseRepositoryImpl[T, ID]) CommitTransaction(tx *gorm.DB) error {
	return tx.Commit().Error
}

func (r *BaseRepositoryImpl[T, ID]) RollbackTransaction(tx *gorm.DB) error {
	return tx.Rollback().Error
}

// Basic CRUD operations
func (r *BaseRepositoryImpl[T, ID]) Create(ctx context.Context, entity *T) error {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		r.logger.Error("Failed to create entity", zap.Error(err))
		return fmt.Errorf("create entity: %w", err)
	}
	return nil
}

func (r *BaseRepositoryImpl[T, ID]) GetByID(ctx context.Context, id ID) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("entity with id %v not found", id)
		}
		r.logger.Error("Failed to get entity by ID", zap.Any("id", id), zap.Error(err))
		return nil, fmt.Errorf("get entity by id %v: %w", id, err)
	}
	return &entity, nil
}

func (r *BaseRepositoryImpl[T, ID]) Update(ctx context.Context, entity *T) error {
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		r.logger.Error("Failed to update entity", zap.Error(err))
		return fmt.Errorf("update entity: %w", err)
	}
	return nil
}

func (r *BaseRepositoryImpl[T, ID]) Delete(ctx context.Context, id ID) error {
	var entity T
	if err := r.db.WithContext(ctx).Delete(&entity, id).Error; err != nil {
		r.logger.Error("Failed to delete entity", zap.Any("id", id), zap.Error(err))
		return fmt.Errorf("delete entity with id %v: %w", id, err)
	}
	return nil
}

// Soft delete operations
func (r *BaseRepositoryImpl[T, ID]) SoftDelete(ctx context.Context, id ID) error {
	var entity T
	result := r.db.WithContext(ctx).Delete(&entity, id)
	if result.Error != nil {
		r.logger.Error("Failed to soft delete entity", zap.Any("id", id), zap.Error(result.Error))
		return fmt.Errorf("soft delete entity with id %v: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("entity with id %v not found for soft delete", id)
	}
	return nil
}

func (r *BaseRepositoryImpl[T, ID]) Restore(ctx context.Context, id ID) error {
	var entity T
	result := r.db.WithContext(ctx).Unscoped().Model(&entity).Where("id = ?", id).Update("deleted_at", nil)
	if result.Error != nil {
		r.logger.Error("Failed to restore entity", zap.Any("id", id), zap.Error(result.Error))
		return fmt.Errorf("restore entity with id %v: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("entity with id %v not found for restore", id)
	}
	return nil
}

func (r *BaseRepositoryImpl[T, ID]) HardDelete(ctx context.Context, id ID) error {
	var entity T
	result := r.db.WithContext(ctx).Unscoped().Delete(&entity, id)
	if result.Error != nil {
		r.logger.Error("Failed to hard delete entity", zap.Any("id", id), zap.Error(result.Error))
		return fmt.Errorf("hard delete entity with id %v: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("entity with id %v not found for hard delete", id)
	}
	return nil
}

// List operations with pagination
func (r *BaseRepositoryImpl[T, ID]) List(ctx context.Context, limit, offset int) ([]*T, int64, error) {
	var entities []*T
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(new(T)).Count(&total).Error; err != nil {
		r.logger.Error("Failed to count entities", zap.Error(err))
		return nil, 0, fmt.Errorf("count entities: %w", err)
	}

	// Get paginated records
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		r.logger.Error("Failed to list entities", zap.Error(err))
		return nil, 0, fmt.Errorf("list entities: %w", err)
	}

	return entities, total, nil
}

func (r *BaseRepositoryImpl[T, ID]) ListDeleted(ctx context.Context, limit, offset int) ([]*T, int64, error) {
	var entities []*T
	var total int64

	// Count total deleted records
	if err := r.db.WithContext(ctx).Unscoped().Model(new(T)).Where("deleted_at IS NOT NULL").Count(&total).Error; err != nil {
		r.logger.Error("Failed to count deleted entities", zap.Error(err))
		return nil, 0, fmt.Errorf("count deleted entities: %w", err)
	}

	// Get paginated deleted records
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		r.logger.Error("Failed to list deleted entities", zap.Error(err))
		return nil, 0, fmt.Errorf("list deleted entities: %w", err)
	}

	return entities, total, nil
}

func (r *BaseRepositoryImpl[T, ID]) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*T, int64, error) {
	var entities []*T
	var total int64

	// Count total records with status
	if err := r.db.WithContext(ctx).Model(new(T)).Where("status = ?", status).Count(&total).Error; err != nil {
		r.logger.Error("Failed to count entities by status", zap.String("status", status), zap.Error(err))
		return nil, 0, fmt.Errorf("count entities by status %s: %w", status, err)
	}

	// Get paginated records with status
	if err := r.db.WithContext(ctx).Where("status = ?", status).Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		r.logger.Error("Failed to list entities by status", zap.String("status", status), zap.Error(err))
		return nil, 0, fmt.Errorf("list entities by status %s: %w", status, err)
	}

	return entities, total, nil
}

// Search operations
func (r *BaseRepositoryImpl[T, ID]) Search(ctx context.Context, query string, limit, offset int) ([]*T, int64, error) {
	var entities []*T
	var total int64

	// Basic search implementation - can be overridden by specific repositories
	searchQuery := "%" + query + "%"

	// Try to search in common fields
	db := r.db.WithContext(ctx).Model(new(T))

	// Use reflection to check if entity has searchable fields
	entityType := reflect.TypeOf(new(T)).Elem()
	var whereConditions []string
	var whereArgs []any

	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		// Search in string fields that might contain searchable content
		if field.Type.Kind() == reflect.String {
			fieldName := field.Tag.Get("gorm")
			if fieldName == "" {
				fieldName = toSnakeCase(field.Name)
			}
			// Common searchable fields
			if contains([]string{"name", "title", "description", "email", "username"}, fieldName) {
				whereConditions = append(whereConditions, fieldName+" LIKE ?")
				whereArgs = append(whereArgs, searchQuery)
			}
		}
	}

	if len(whereConditions) > 0 {
		whereClause := fmt.Sprintf("(%s)", whereConditions[0])
		for _, condition := range whereConditions[1:] {
			whereClause += " OR " + condition
		}
		db = db.Where(whereClause, whereArgs...)
	}

	// Count total matching records
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count search results", zap.String("query", query), zap.Error(err))
		return nil, 0, fmt.Errorf("count search results for query %s: %w", query, err)
	}

	// Get paginated search results
	if err := db.Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		r.logger.Error("Failed to search entities", zap.String("query", query), zap.Error(err))
		return nil, 0, fmt.Errorf("search entities for query %s: %w", query, err)
	}

	return entities, total, nil
}

// Status management
func (r *BaseRepositoryImpl[T, ID]) UpdateStatus(ctx context.Context, id ID, status string) error {
	var entity T
	result := r.db.WithContext(ctx).Model(&entity).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		r.logger.Error("Failed to update entity status", zap.Any("id", id), zap.String("status", status), zap.Error(result.Error))
		return fmt.Errorf("update entity status for id %v to %s: %w", id, status, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("entity with id %v not found for status update", id)
	}
	return nil
}

// Statistics operations
func (r *BaseRepositoryImpl[T, ID]) CountTotal(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(new(T)).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count total entities", zap.Error(err))
		return 0, fmt.Errorf("count total entities: %w", err)
	}
	return count, nil
}

func (r *BaseRepositoryImpl[T, ID]) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(new(T)).Where("status = ?", status).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count entities by status", zap.String("status", status), zap.Error(err))
		return 0, fmt.Errorf("count entities by status %s: %w", status, err)
	}
	return count, nil
}

func (r *BaseRepositoryImpl[T, ID]) CountDeleted(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Unscoped().Model(new(T)).Where("deleted_at IS NOT NULL").Count(&count).Error; err != nil {
		r.logger.Error("Failed to count deleted entities", zap.Error(err))
		return 0, fmt.Errorf("count deleted entities: %w", err)
	}
	return count, nil
}

// Batch operations
func (r *BaseRepositoryImpl[T, ID]) BatchDelete(ctx context.Context, ids []ID) (int, []ID, error) {
	if len(ids) == 0 {
		return 0, nil, nil
	}

	var successCount int
	var failedIDs []ID

	for _, id := range ids {
		if err := r.SoftDelete(ctx, id); err != nil {
			r.logger.Warn("Failed to delete entity in batch", zap.Any("id", id), zap.Error(err))
			failedIDs = append(failedIDs, id)
		} else {
			successCount++
		}
	}

	return successCount, failedIDs, nil
}

func (r *BaseRepositoryImpl[T, ID]) BatchRestore(ctx context.Context, ids []ID) (int, []ID, error) {
	if len(ids) == 0 {
		return 0, nil, nil
	}

	var successCount int
	var failedIDs []ID

	for _, id := range ids {
		if err := r.Restore(ctx, id); err != nil {
			r.logger.Warn("Failed to restore entity in batch", zap.Any("id", id), zap.Error(err))
			failedIDs = append(failedIDs, id)
		} else {
			successCount++
		}
	}

	return successCount, failedIDs, nil
}

func (r *BaseRepositoryImpl[T, ID]) BatchUpdateStatus(ctx context.Context, ids []ID, status string) (int, []ID, error) {
	if len(ids) == 0 {
		return 0, nil, nil
	}

	var successCount int
	var failedIDs []ID

	for _, id := range ids {
		if err := r.UpdateStatus(ctx, id, status); err != nil {
			r.logger.Warn("Failed to update entity status in batch", zap.Any("id", id), zap.String("status", status), zap.Error(err))
			failedIDs = append(failedIDs, id)
		} else {
			successCount++
		}
	}

	return successCount, failedIDs, nil
}

// Existence checks
func (r *BaseRepositoryImpl[T, ID]) ExistsByID(ctx context.Context, id ID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Count(&count).Error; err != nil {
		r.logger.Error("Failed to check entity exists", zap.Any("id", id), zap.Error(err))
		return false, fmt.Errorf("check entity exists for id %v: %w", id, err)
	}
	return count > 0, nil
}

// Advanced filtering
func (r *BaseRepositoryImpl[T, ID]) ListWithFilters(ctx context.Context, filters map[string]any, limit, offset int) ([]*T, int64, error) {
	var entities []*T
	var total int64

	db := r.db.WithContext(ctx).Model(new(T))

	// Apply filters
	for field, value := range filters {
		if value != nil {
			db = db.Where(field+" = ?", value)
		}
	}

	// Count total filtered records
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count filtered entities", zap.Any("filters", filters), zap.Error(err))
		return nil, 0, fmt.Errorf("count filtered entities: %w", err)
	}

	// Get paginated filtered records
	if err := db.Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		r.logger.Error("Failed to list filtered entities", zap.Any("filters", filters), zap.Error(err))
		return nil, 0, fmt.Errorf("list filtered entities: %w", err)
	}

	return entities, total, nil
}

// UserScopedRepositoryImpl extends BaseRepositoryImpl with user-specific operations
type UserScopedRepositoryImpl[T any, ID comparable] struct {
	*BaseRepositoryImpl[T, ID]
}

// NewUserScopedRepository creates a new UserScopedRepositoryImpl instance
func NewUserScopedRepository[T any, ID comparable](db *gorm.DB, logger framework.Logger) *UserScopedRepositoryImpl[T, ID] {
	return &UserScopedRepositoryImpl[T, ID]{
		BaseRepositoryImpl: NewBaseRepository[T, ID](db, logger),
	}
}

// User-specific operations
func (r *UserScopedRepositoryImpl[T, ID]) ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*T, int64, error) {
	var entities []*T
	var total int64

	// Count total records for user
	if err := r.db.WithContext(ctx).Model(new(T)).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		r.logger.Error("Failed to count entities by user", zap.Uint("user_id", userID), zap.Error(err))
		return nil, 0, fmt.Errorf("count entities by user %d: %w", userID, err)
	}

	// Get paginated records for user
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		r.logger.Error("Failed to list entities by user", zap.Uint("user_id", userID), zap.Error(err))
		return nil, 0, fmt.Errorf("list entities by user %d: %w", userID, err)
	}

	return entities, total, nil
}

func (r *UserScopedRepositoryImpl[T, ID]) CountByUser(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(new(T)).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count entities by user", zap.Uint("user_id", userID), zap.Error(err))
		return 0, fmt.Errorf("count entities by user %d: %w", userID, err)
	}
	return count, nil
}

func (r *UserScopedRepositoryImpl[T, ID]) GetUserTotalCount(ctx context.Context, userID uint) (int64, error) {
	return r.CountByUser(ctx, userID)
}

// TimeBasedRepositoryImpl extends BaseRepositoryImpl with time-based operations
type TimeBasedRepositoryImpl[T any, ID comparable] struct {
	*BaseRepositoryImpl[T, ID]
}

// NewTimeBasedRepository creates a new TimeBasedRepositoryImpl instance
func NewTimeBasedRepository[T any, ID comparable](db *gorm.DB, logger framework.Logger) *TimeBasedRepositoryImpl[T, ID] {
	return &TimeBasedRepositoryImpl[T, ID]{
		BaseRepositoryImpl: NewBaseRepository[T, ID](db, logger),
	}
}

// Time-based queries
func (r *TimeBasedRepositoryImpl[T, ID]) ListByDateRange(ctx context.Context, field string, start, end time.Time, limit, offset int) ([]*T, int64, error) {
	var entities []*T
	var total int64

	db := r.db.WithContext(ctx).Model(new(T)).Where(field+" BETWEEN ? AND ?", start, end)

	// Count total records in date range
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count entities by date range", zap.String("field", field), zap.Time("start", start), zap.Time("end", end), zap.Error(err))
		return nil, 0, fmt.Errorf("count entities by date range: %w", err)
	}

	// Get paginated records in date range
	if err := db.Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		r.logger.Error("Failed to list entities by date range", zap.String("field", field), zap.Time("start", start), zap.Time("end", end), zap.Error(err))
		return nil, 0, fmt.Errorf("list entities by date range: %w", err)
	}

	return entities, total, nil
}

func (r *TimeBasedRepositoryImpl[T, ID]) ListCreatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*T, int64, error) {
	return r.ListByDateRange(ctx, "created_at", after, time.Now(), limit, offset)
}

func (r *TimeBasedRepositoryImpl[T, ID]) ListUpdatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*T, int64, error) {
	return r.ListByDateRange(ctx, "updated_at", after, time.Now(), limit, offset)
}

// Utility functions
func toSnakeCase(str string) string {
	var result []byte
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		if r >= 'A' && r <= 'Z' {
			result = append(result, byte(r-'A'+'a'))
		} else {
			result = append(result, byte(r))
		}
	}
	return string(result)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// UserScopedTimeBasedRepositoryImpl extends BaseRepositoryImpl with both user-specific and time-based operations
type UserScopedTimeBasedRepositoryImpl[T any, ID comparable] struct {
	*BaseRepositoryImpl[T, ID]
}

// NewUserScopedTimeBasedRepository creates a new UserScopedTimeBasedRepositoryImpl instance
func NewUserScopedTimeBasedRepository[T any, ID comparable](db *gorm.DB, logger framework.Logger) *UserScopedTimeBasedRepositoryImpl[T, ID] {
	return &UserScopedTimeBasedRepositoryImpl[T, ID]{
		BaseRepositoryImpl: NewBaseRepository[T, ID](db, logger),
	}
}

// User-specific operations
func (r *UserScopedTimeBasedRepositoryImpl[T, ID]) ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*T, int64, error) {
	var entities []*T
	var total int64

	// Count total records for user
	if err := r.GetDB().WithContext(ctx).Model(new(T)).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		r.logger.Error("Failed to count entities by user", zap.Uint("user_id", userID), zap.Error(err))
		return nil, 0, fmt.Errorf("count entities by user %d: %w", userID, err)
	}

	// Get paginated records for user
	if err := r.GetDB().WithContext(ctx).Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		r.logger.Error("Failed to list entities by user", zap.Uint("user_id", userID), zap.Error(err))
		return nil, 0, fmt.Errorf("list entities by user %d: %w", userID, err)
	}

	return entities, total, nil
}

func (r *UserScopedTimeBasedRepositoryImpl[T, ID]) CountByUser(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(new(T)).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count entities by user", zap.Uint("user_id", userID), zap.Error(err))
		return 0, fmt.Errorf("count entities by user %d: %w", userID, err)
	}
	return count, nil
}

func (r *UserScopedTimeBasedRepositoryImpl[T, ID]) GetUserTotalCount(ctx context.Context, userID uint) (int64, error) {
	return r.CountByUser(ctx, userID)
}

// Time-based operations
func (r *UserScopedTimeBasedRepositoryImpl[T, ID]) ListByDateRange(ctx context.Context, field string, start, end time.Time, limit, offset int) ([]*T, int64, error) {
	var entities []*T
	var total int64

	db := r.GetDB().WithContext(ctx).Model(new(T)).Where(field+" BETWEEN ? AND ?", start, end)

	// Count total records in date range
	if err := db.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count entities by date range", zap.String("field", field), zap.Time("start", start), zap.Time("end", end), zap.Error(err))
		return nil, 0, fmt.Errorf("count entities by date range: %w", err)
	}

	// Get paginated records in date range
	if err := db.Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
		r.logger.Error("Failed to list entities by date range", zap.String("field", field), zap.Time("start", start), zap.Time("end", end), zap.Error(err))
		return nil, 0, fmt.Errorf("list entities by date range: %w", err)
	}

	return entities, total, nil
}

func (r *UserScopedTimeBasedRepositoryImpl[T, ID]) ListCreatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*T, int64, error) {
	return r.ListByDateRange(ctx, "created_at", after, time.Now(), limit, offset)
}

func (r *UserScopedTimeBasedRepositoryImpl[T, ID]) ListUpdatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*T, int64, error) {
	return r.ListByDateRange(ctx, "updated_at", after, time.Now(), limit, offset)
}
