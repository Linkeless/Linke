package generic

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Repository defines the generic repository interface
type Repository[T any, ID comparable] interface {
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id ID) error
	FindByID(ctx context.Context, id ID) (*T, error)
	FindAll(ctx context.Context, opts ...QueryOption) ([]*T, error)
	Count(ctx context.Context, opts ...QueryOption) (int64, error)
	Exists(ctx context.Context, id ID) (bool, error)
	BatchCreate(ctx context.Context, entities []*T) error
	BatchDelete(ctx context.Context, ids []ID) error
}

// QueryOption is a function that modifies a query
type QueryOption func(*gorm.DB) *gorm.DB

// BaseRepository provides a generic GORM-based repository implementation
type BaseRepository[T any, ID comparable] struct {
	db        *gorm.DB
	tableName string
}

// NewBaseRepository creates a new generic repository
func NewBaseRepository[T any, ID comparable](db *gorm.DB) *BaseRepository[T, ID] {
	return &BaseRepository[T, ID]{
		db: db,
	}
}

// WithTable sets the table name for the repository
func (r *BaseRepository[T, ID]) WithTable(tableName string) *BaseRepository[T, ID] {
	r.tableName = tableName
	return r
}

// getDB returns the database instance with optional table name
func (r *BaseRepository[T, ID]) getDB() *gorm.DB {
	if r.tableName != "" {
		return r.db.Table(r.tableName)
	}
	return r.db
}

// Create creates a new entity
func (r *BaseRepository[T, ID]) Create(ctx context.Context, entity *T) error {
	if err := r.getDB().WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}
	return nil
}

// Update updates an existing entity
func (r *BaseRepository[T, ID]) Update(ctx context.Context, entity *T) error {
	if err := r.getDB().WithContext(ctx).Save(entity).Error; err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}
	return nil
}

// Delete deletes an entity by ID
func (r *BaseRepository[T, ID]) Delete(ctx context.Context, id ID) error {
	var entity T
	if err := r.getDB().WithContext(ctx).Delete(&entity, id).Error; err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}
	return nil
}

// FindByID finds an entity by ID
func (r *BaseRepository[T, ID]) FindByID(ctx context.Context, id ID) (*T, error) {
	var entity T
	if err := r.getDB().WithContext(ctx).First(&entity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("entity not found")
		}
		return nil, fmt.Errorf("failed to find entity: %w", err)
	}
	return &entity, nil
}

// FindAll finds all entities with optional query options
func (r *BaseRepository[T, ID]) FindAll(ctx context.Context, opts ...QueryOption) ([]*T, error) {
	db := r.getDB().WithContext(ctx)

	// Apply query options
	for _, opt := range opts {
		db = opt(db)
	}

	var entities []*T
	if err := db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find entities: %w", err)
	}
	return entities, nil
}

// Count counts entities with optional query options
func (r *BaseRepository[T, ID]) Count(ctx context.Context, opts ...QueryOption) (int64, error) {
	db := r.getDB().WithContext(ctx).Model(new(T))

	// Apply query options
	for _, opt := range opts {
		db = opt(db)
	}

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count entities: %w", err)
	}
	return count, nil
}

// Exists checks if an entity exists by ID
func (r *BaseRepository[T, ID]) Exists(ctx context.Context, id ID) (bool, error) {
	var count int64
	if err := r.getDB().WithContext(ctx).Model(new(T)).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return count > 0, nil
}

// BatchCreate creates multiple entities
func (r *BaseRepository[T, ID]) BatchCreate(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}
	if err := r.getDB().WithContext(ctx).CreateInBatches(entities, 100).Error; err != nil {
		return fmt.Errorf("failed to batch create entities: %w", err)
	}
	return nil
}

// BatchDelete deletes multiple entities by IDs
func (r *BaseRepository[T, ID]) BatchDelete(ctx context.Context, ids []ID) error {
	if len(ids) == 0 {
		return nil
	}
	var entity T
	if err := r.getDB().WithContext(ctx).Delete(&entity, ids).Error; err != nil {
		return fmt.Errorf("failed to batch delete entities: %w", err)
	}
	return nil
}

// Common query options

// WithLimit returns a query option that limits the number of results
func WithLimit(limit int) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

// WithOffset returns a query option that skips a number of results
func WithOffset(offset int) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	}
}

// WithOrder returns a query option that orders the results
func WithOrder(order string) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(order)
	}
}

// WithWhere returns a query option that adds a where condition
func WithWhere(query string, args ...interface{}) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	}
}

// WithPreload returns a query option that preloads associations
func WithPreload(column string, conditions ...interface{}) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(column, conditions...)
	}
}

// WithPagination returns a query option for pagination
func WithPagination(page, pageSize int) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
