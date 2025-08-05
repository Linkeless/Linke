package services

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"linke/internal/shared/framework"

	"go.uber.org/zap"
)

// BaseServiceImpl provides a concrete implementation of GenericService
type BaseServiceImpl[T any, ID comparable] struct {
	name       string
	repository framework.GenericRepository[T, ID]
	logger     framework.Logger
	eventPub   framework.EventPublisher
	validator  framework.Validator
}

// NewBaseService creates a new BaseServiceImpl instance
func NewBaseService[T any, ID comparable](
	name string,
	repository framework.GenericRepository[T, ID],
	logger framework.Logger,
	eventPub framework.EventPublisher,
	validator framework.Validator,
) *BaseServiceImpl[T, ID] {
	return &BaseServiceImpl[T, ID]{
		name:       name,
		repository: repository,
		logger:     logger,
		eventPub:   eventPub,
		validator:  validator,
	}
}

// Service interface methods
func (s *BaseServiceImpl[T, ID]) GetName() string {
	return s.name
}

func (s *BaseServiceImpl[T, ID]) Initialize(ctx context.Context) error {
	s.logger.Info("Initializing service", zap.String("service", s.name))
	return nil
}

func (s *BaseServiceImpl[T, ID]) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down service", zap.String("service", s.name))
	return nil
}

// Basic CRUD operations
func (s *BaseServiceImpl[T, ID]) Create(ctx context.Context, req *framework.CreateRequest[T]) (*T, error) {
	if req == nil || req.Data == nil {
		return nil, fmt.Errorf("create request and data cannot be nil")
	}

	// Validate if not skipped
	if req.Options == nil || !req.Options.SkipValidation {
		if err := s.ValidateCreate(ctx, req); err != nil {
			s.logger.Error("Create validation failed", zap.Error(err))
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	// Create entity
	if err := s.repository.Create(ctx, req.Data); err != nil {
		s.logger.Error("Failed to create entity", zap.String("service", s.name), zap.Error(err))
		return nil, fmt.Errorf("create entity: %w", err)
	}

	// Publish events if enabled
	if req.Options != nil && req.Options.PublishEvents && s.eventPub != nil {
		if err := s.PublishCreatedEvent(ctx, req.Data); err != nil {
			s.logger.Warn("Failed to publish created event", zap.Error(err))
		}
	}

	s.logger.Info("Entity created successfully", zap.String("service", s.name))
	return req.Data, nil
}

func (s *BaseServiceImpl[T, ID]) GetByID(ctx context.Context, id ID) (*T, error) {
	entity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get entity by ID", zap.String("service", s.name), zap.Any("id", id), zap.Error(err))
		return nil, fmt.Errorf("get entity by id %v: %w", id, err)
	}
	return entity, nil
}

func (s *BaseServiceImpl[T, ID]) Update(ctx context.Context, id ID, req *framework.UpdateRequest[T]) (*T, error) {
	if req == nil || req.Data == nil {
		return nil, fmt.Errorf("update request and data cannot be nil")
	}

	// Get existing entity
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get existing entity: %w", err)
	}

	// Validate if not skipped
	if req.Options == nil || !req.Options.SkipValidation {
		if err := s.ValidateUpdate(ctx, id, req); err != nil {
			s.logger.Error("Update validation failed", zap.Any("id", id), zap.Error(err))
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	// Update entity
	if err := s.repository.Update(ctx, req.Data); err != nil {
		s.logger.Error("Failed to update entity", zap.String("service", s.name), zap.Any("id", id), zap.Error(err))
		return nil, fmt.Errorf("update entity: %w", err)
	}

	// Publish events if enabled
	if req.Options != nil && req.Options.PublishEvents && s.eventPub != nil {
		if err := s.PublishUpdatedEvent(ctx, existing, req.Data); err != nil {
			s.logger.Warn("Failed to publish updated event", zap.Error(err))
		}
	}

	s.logger.Info("Entity updated successfully", zap.String("service", s.name), zap.Any("id", id))
	return req.Data, nil
}

func (s *BaseServiceImpl[T, ID]) Delete(ctx context.Context, id ID) error {
	// Validate delete operation
	if err := s.ValidateDelete(ctx, id); err != nil {
		s.logger.Error("Delete validation failed", zap.Any("id", id), zap.Error(err))
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get entity before deleting for events
	var entity *T
	if s.eventPub != nil {
		var err error
		entity, err = s.repository.GetByID(ctx, id)
		if err != nil {
			s.logger.Warn("Failed to get entity before delete for events", zap.Any("id", id), zap.Error(err))
		}
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete entity", zap.String("service", s.name), zap.Any("id", id), zap.Error(err))
		return fmt.Errorf("delete entity with id %v: %w", id, err)
	}

	// Publish events
	if entity != nil && s.eventPub != nil {
		if err := s.PublishDeletedEvent(ctx, entity); err != nil {
			s.logger.Warn("Failed to publish deleted event", zap.Error(err))
		}
	}

	s.logger.Info("Entity deleted successfully", zap.String("service", s.name), zap.Any("id", id))
	return nil
}

// Soft delete operations
func (s *BaseServiceImpl[T, ID]) SoftDelete(ctx context.Context, id ID) error {
	// Validate delete operation
	if err := s.ValidateDelete(ctx, id); err != nil {
		s.logger.Error("Soft delete validation failed", zap.Any("id", id), zap.Error(err))
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.repository.SoftDelete(ctx, id); err != nil {
		s.logger.Error("Failed to soft delete entity", zap.String("service", s.name), zap.Any("id", id), zap.Error(err))
		return fmt.Errorf("soft delete entity with id %v: %w", id, err)
	}

	s.logger.Info("Entity soft deleted successfully", zap.String("service", s.name), zap.Any("id", id))
	return nil
}

func (s *BaseServiceImpl[T, ID]) Restore(ctx context.Context, id ID) error {
	if err := s.repository.Restore(ctx, id); err != nil {
		s.logger.Error("Failed to restore entity", zap.String("service", s.name), zap.Any("id", id), zap.Error(err))
		return fmt.Errorf("restore entity with id %v: %w", id, err)
	}

	s.logger.Info("Entity restored successfully", zap.String("service", s.name), zap.Any("id", id))
	return nil
}

func (s *BaseServiceImpl[T, ID]) HardDelete(ctx context.Context, id ID) error {
	// Validate delete operation
	if err := s.ValidateDelete(ctx, id); err != nil {
		s.logger.Error("Hard delete validation failed", zap.Any("id", id), zap.Error(err))
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.repository.HardDelete(ctx, id); err != nil {
		s.logger.Error("Failed to hard delete entity", zap.String("service", s.name), zap.Any("id", id), zap.Error(err))
		return fmt.Errorf("hard delete entity with id %v: %w", id, err)
	}

	s.logger.Info("Entity hard deleted successfully", zap.String("service", s.name), zap.Any("id", id))
	return nil
}

// List operations with pagination
func (s *BaseServiceImpl[T, ID]) List(ctx context.Context, req *framework.ListRequest) (*framework.ListResponse[T], error) {
	if req == nil {
		req = &framework.ListRequest{Limit: 10, Offset: 0}
	}
	
	// Set default limit if not provided
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	entities, total, err := s.repository.List(ctx, req.Limit, req.Offset)
	if err != nil {
		s.logger.Error("Failed to list entities", zap.String("service", s.name), zap.Error(err))
		return nil, fmt.Errorf("list entities: %w", err)
	}

	return s.buildListResponse(entities, total, req), nil
}

func (s *BaseServiceImpl[T, ID]) ListDeleted(ctx context.Context, req *framework.ListRequest) (*framework.ListResponse[T], error) {
	if req == nil {
		req = &framework.ListRequest{Limit: 10, Offset: 0}
	}
	
	entities, total, err := s.repository.ListDeleted(ctx, req.Limit, req.Offset)
	if err != nil {
		s.logger.Error("Failed to list deleted entities", zap.String("service", s.name), zap.Error(err))
		return nil, fmt.Errorf("list deleted entities: %w", err)
	}

	return s.buildListResponse(entities, total, req), nil
}

func (s *BaseServiceImpl[T, ID]) ListByStatus(ctx context.Context, status string, req *framework.ListRequest) (*framework.ListResponse[T], error) {
	if req == nil {
		req = &framework.ListRequest{Limit: 10, Offset: 0}
	}
	
	entities, total, err := s.repository.ListByStatus(ctx, status, req.Limit, req.Offset)
	if err != nil {
		s.logger.Error("Failed to list entities by status", zap.String("service", s.name), zap.String("status", status), zap.Error(err))
		return nil, fmt.Errorf("list entities by status %s: %w", status, err)
	}

	return s.buildListResponse(entities, total, req), nil
}

// Search operations
func (s *BaseServiceImpl[T, ID]) Search(ctx context.Context, query string, req *framework.ListRequest) (*framework.ListResponse[T], error) {
	if req == nil {
		req = &framework.ListRequest{Limit: 10, Offset: 0}
	}
	
	entities, total, err := s.repository.Search(ctx, query, req.Limit, req.Offset)
	if err != nil {
		s.logger.Error("Failed to search entities", zap.String("service", s.name), zap.String("query", query), zap.Error(err))
		return nil, fmt.Errorf("search entities: %w", err)
	}

	return s.buildListResponse(entities, total, req), nil
}

// Status management
func (s *BaseServiceImpl[T, ID]) UpdateStatus(ctx context.Context, id ID, status string) (*T, error) {
	if err := s.repository.UpdateStatus(ctx, id, status); err != nil {
		s.logger.Error("Failed to update entity status", zap.String("service", s.name), zap.Any("id", id), zap.String("status", status), zap.Error(err))
		return nil, fmt.Errorf("update entity status: %w", err)
	}

	// Get updated entity
	entity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get updated entity", zap.Any("id", id), zap.Error(err))
		return nil, fmt.Errorf("get updated entity: %w", err)
	}

	s.logger.Info("Entity status updated successfully", zap.String("service", s.name), zap.Any("id", id), zap.String("status", status))
	return entity, nil
}

// Statistics operations
func (s *BaseServiceImpl[T, ID]) GetStatistics(ctx context.Context) (*framework.StatisticsResponse, error) {
	totalCount, err := s.repository.CountTotal(ctx)
	if err != nil {
		return nil, fmt.Errorf("count total: %w", err)
	}

	deletedCount, err := s.repository.CountDeleted(ctx)
	if err != nil {
		return nil, fmt.Errorf("count deleted: %w", err)
	}

	// Try to get status counts for common statuses
	statusCounts := make(map[string]int64)
	commonStatuses := []string{"active", "inactive", "pending", "completed", "cancelled"}
	
	var activeCount, inactiveCount int64
	for _, status := range commonStatuses {
		count, err := s.repository.CountByStatus(ctx, status)
		if err != nil {
			s.logger.Debug("Failed to count by status", zap.String("status", status), zap.Error(err))
			continue
		}
		statusCounts[status] = count
		
		if status == "active" {
			activeCount = count
		} else if status == "inactive" {
			inactiveCount = count
		}
	}

	return &framework.StatisticsResponse{
		TotalCount:    totalCount,
		ActiveCount:   activeCount,
		InactiveCount: inactiveCount,
		DeletedCount:  deletedCount,
		StatusCounts:  statusCounts,
		GeneratedAt:   time.Now(),
	}, nil
}

func (s *BaseServiceImpl[T, ID]) CountTotal(ctx context.Context) (int64, error) {
	return s.repository.CountTotal(ctx)
}

func (s *BaseServiceImpl[T, ID]) CountByStatus(ctx context.Context, status string) (int64, error) {
	return s.repository.CountByStatus(ctx, status)
}

func (s *BaseServiceImpl[T, ID]) CountDeleted(ctx context.Context) (int64, error) {
	return s.repository.CountDeleted(ctx)
}

// Batch operations
func (s *BaseServiceImpl[T, ID]) BatchDelete(ctx context.Context, ids []ID) (*framework.BatchOperationResponse, error) {
	successCount, failedIDs, err := s.repository.BatchDelete(ctx, ids)
	if err != nil {
		s.logger.Error("Batch delete failed", zap.String("service", s.name), zap.Error(err))
		return nil, fmt.Errorf("batch delete: %w", err)
	}

	response := &framework.BatchOperationResponse{
		SuccessCount: successCount,
		FailedCount:  len(failedIDs),
	}
	
	// Convert failed IDs to []uint for response
	if len(failedIDs) > 0 {
		response.FailedIDs = make([]uint, len(failedIDs))
		for i, id := range failedIDs {
			// This assumes ID can be converted to uint - may need adjustment for other ID types
			response.FailedIDs[i] = uint(reflect.ValueOf(id).Uint())
		}
	}

	s.logger.Info("Batch delete completed", zap.String("service", s.name), zap.Int("success", successCount), zap.Int("failed", len(failedIDs)))
	return response, nil
}

func (s *BaseServiceImpl[T, ID]) BatchRestore(ctx context.Context, ids []ID) (*framework.BatchOperationResponse, error) {
	successCount, failedIDs, err := s.repository.BatchRestore(ctx, ids)
	if err != nil {
		s.logger.Error("Batch restore failed", zap.String("service", s.name), zap.Error(err))
		return nil, fmt.Errorf("batch restore: %w", err)
	}

	response := &framework.BatchOperationResponse{
		SuccessCount: successCount,
		FailedCount:  len(failedIDs),
	}
	
	// Convert failed IDs to []uint for response
	if len(failedIDs) > 0 {
		response.FailedIDs = make([]uint, len(failedIDs))
		for i, id := range failedIDs {
			response.FailedIDs[i] = uint(reflect.ValueOf(id).Uint())
		}
	}

	s.logger.Info("Batch restore completed", zap.String("service", s.name), zap.Int("success", successCount), zap.Int("failed", len(failedIDs)))
	return response, nil
}

func (s *BaseServiceImpl[T, ID]) BatchUpdateStatus(ctx context.Context, ids []ID, status string) (*framework.BatchOperationResponse, error) {
	successCount, failedIDs, err := s.repository.BatchUpdateStatus(ctx, ids, status)
	if err != nil {
		s.logger.Error("Batch update status failed", zap.String("service", s.name), zap.String("status", status), zap.Error(err))
		return nil, fmt.Errorf("batch update status: %w", err)
	}

	response := &framework.BatchOperationResponse{
		SuccessCount: successCount,
		FailedCount:  len(failedIDs),
	}
	
	// Convert failed IDs to []uint for response
	if len(failedIDs) > 0 {
		response.FailedIDs = make([]uint, len(failedIDs))
		for i, id := range failedIDs {
			response.FailedIDs[i] = uint(reflect.ValueOf(id).Uint())
		}
	}

	s.logger.Info("Batch update status completed", zap.String("service", s.name), zap.String("status", status), zap.Int("success", successCount), zap.Int("failed", len(failedIDs)))
	return response, nil
}

// Existence checks
func (s *BaseServiceImpl[T, ID]) ExistsByID(ctx context.Context, id ID) (bool, error) {
	return s.repository.ExistsByID(ctx, id)
}

// Advanced filtering
func (s *BaseServiceImpl[T, ID]) ListWithFilters(ctx context.Context, filters map[string]any, req *framework.ListRequest) (*framework.ListResponse[T], error) {
	if req == nil {
		req = &framework.ListRequest{Limit: 10, Offset: 0}
	}
	
	entities, total, err := s.repository.ListWithFilters(ctx, filters, req.Limit, req.Offset)
	if err != nil {
		s.logger.Error("Failed to list entities with filters", zap.String("service", s.name), zap.Any("filters", filters), zap.Error(err))
		return nil, fmt.Errorf("list entities with filters: %w", err)
	}

	return s.buildListResponse(entities, total, req), nil
}

// Validation hooks (default implementations - can be overridden)
func (s *BaseServiceImpl[T, ID]) ValidateCreate(ctx context.Context, req *framework.CreateRequest[T]) error {
	if s.validator == nil {
		return nil
	}
	return s.validator.ValidateStruct(req.Data)
}

func (s *BaseServiceImpl[T, ID]) ValidateUpdate(ctx context.Context, id ID, req *framework.UpdateRequest[T]) error {
	if s.validator == nil {
		return nil
	}
	return s.validator.ValidateStruct(req.Data)
}

func (s *BaseServiceImpl[T, ID]) ValidateDelete(ctx context.Context, id ID) error {
	// Check if entity exists
	exists, err := s.repository.ExistsByID(ctx, id)
	if err != nil {
		return fmt.Errorf("check entity exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("entity with id %v not found", id)
	}
	return nil
}

// Event publishing methods (default implementations - can be overridden)
func (s *BaseServiceImpl[T, ID]) PublishCreatedEvent(ctx context.Context, entity *T) error {
	if s.eventPub == nil {
		return nil
	}
	// Default implementation would need to be customized based on event structure
	// This is a placeholder
	s.logger.Debug("Published created event", zap.String("service", s.name))
	return nil
}

func (s *BaseServiceImpl[T, ID]) PublishUpdatedEvent(ctx context.Context, old *T, new *T) error {
	if s.eventPub == nil {
		return nil
	}
	// Default implementation would need to be customized based on event structure
	// This is a placeholder
	s.logger.Debug("Published updated event", zap.String("service", s.name))
	return nil
}

func (s *BaseServiceImpl[T, ID]) PublishDeletedEvent(ctx context.Context, entity *T) error {
	if s.eventPub == nil {
		return nil
	}
	// Default implementation would need to be customized based on event structure
	// This is a placeholder
	s.logger.Debug("Published deleted event", zap.String("service", s.name))
	return nil
}

// Helper methods
func (s *BaseServiceImpl[T, ID]) buildListResponse(entities []*T, total int64, req *framework.ListRequest) *framework.ListResponse[T] {
	hasMore := int64(req.Offset+req.Limit) < total
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	if totalPages < 0 {
		totalPages = 0
	}

	return &framework.ListResponse[T]{
		Data:       entities,
		Total:      total,
		Limit:      req.Limit,
		Offset:     req.Offset,
		HasMore:    hasMore,
		TotalPages: totalPages,
	}
}