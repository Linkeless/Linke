package domain

import "context"

// TransactionManager defines transaction management interface
// This belongs to the domain layer as it defines business transaction boundaries
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}