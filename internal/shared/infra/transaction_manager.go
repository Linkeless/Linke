package infra

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"linke/internal/shared/domain"
)

// GormTransactionManager implements TransactionManager using GORM
// This belongs to the infrastructure layer as it contains technical implementation details
type GormTransactionManager struct {
	db *gorm.DB
}

// NewGormTransactionManager creates a new GormTransactionManager
func NewGormTransactionManager(db *gorm.DB) domain.TransactionManager {
	return &GormTransactionManager{db: db}
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// If the function completes successfully, the transaction is committed
func (tm *GormTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create a new context with the transaction
		txCtx := context.WithValue(ctx, "tx", tx)
		
		if err := fn(txCtx); err != nil {
			// Error will automatically trigger rollback
			return fmt.Errorf("transaction failed: %w", err)
		}
		
		// Success will automatically trigger commit
		return nil
	})
}

// GetTransaction extracts the transaction from context if available
// This is a utility function for repositories to access the current transaction
func GetTransaction(ctx context.Context) (*gorm.DB, bool) {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx, true
	}
	return nil, false
}

// InMemoryTransactionManager is a no-op implementation for testing
type InMemoryTransactionManager struct{}

// NewInMemoryTransactionManager creates a new InMemoryTransactionManager
func NewInMemoryTransactionManager() domain.TransactionManager {
	return &InMemoryTransactionManager{}
}

// WithTransaction simply executes the function without transaction support
func (tm *InMemoryTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}