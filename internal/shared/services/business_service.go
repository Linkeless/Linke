package services

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"linke/internal/shared/framework"
	"linke/internal/shared/logger"
)

// BusinessServiceImpl extends BaseServiceImpl with business logic operations
type BusinessServiceImpl[T any, ID comparable] struct {
	*BaseServiceImpl[T, ID]
	businessRules []BusinessRule[T]
	workflows     map[string]WorkflowHandler[T]
}

// BusinessRule defines a business rule validation function
type BusinessRule[T any] func(ctx context.Context, entity *T) error

// WorkflowHandler defines a workflow processing function
type WorkflowHandler[T any] func(ctx context.Context, entity *T, params map[string]any) error

// NewBusinessService creates a new BusinessServiceImpl instance
func NewBusinessService[T any, ID comparable](
	name string,
	repository framework.GenericRepository[T, ID],
	logger framework.Logger,
	eventPub framework.EventPublisher,
	validator framework.Validator,
) *BusinessServiceImpl[T, ID] {
	return &BusinessServiceImpl[T, ID]{
		BaseServiceImpl: NewBaseService(name, repository, logger, eventPub, validator),
		businessRules:   make([]BusinessRule[T], 0),
		workflows:       make(map[string]WorkflowHandler[T]),
	}
}

// AddBusinessRule adds a business rule to the service
func (s *BusinessServiceImpl[T, ID]) AddBusinessRule(rule BusinessRule[T]) {
	s.businessRules = append(s.businessRules, rule)
}

// AddWorkflow adds a workflow handler to the service
func (s *BusinessServiceImpl[T, ID]) AddWorkflow(action string, handler WorkflowHandler[T]) {
	s.workflows[action] = handler
}

// Business validation
func (s *BusinessServiceImpl[T, ID]) ValidateBusinessRules(ctx context.Context, entity *T) error {
	for i, rule := range s.businessRules {
		if err := rule(ctx, entity); err != nil {
			s.logger.Error("Business rule validation failed",
				logger.String("service", s.name),
				logger.Int("rule_index", i),
				logger.ErrorField(err))
			return fmt.Errorf("business rule %d failed: %w", i, err)
		}
	}
	return nil
}

func (s *BusinessServiceImpl[T, ID]) ValidateBusinessRulesForUpdate(ctx context.Context, id ID, req *framework.UpdateRequest[T]) error {
	// Get existing entity for business rule validation
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get existing entity for business validation: %w", err)
	}

	// Create a merged entity for validation (existing + updates)
	mergedEntity := s.mergeEntityForValidation(existing, req.Data)

	return s.ValidateBusinessRules(ctx, mergedEntity)
}

// Event publishing (enhanced implementations)
func (s *BusinessServiceImpl[T, ID]) PublishCreatedEvent(ctx context.Context, entity *T) error {
	if s.eventPub == nil {
		return nil
	}

	// Create a domain event based on entity type
	eventType := fmt.Sprintf("%s.created", s.getEntityTypeName())
	event := &DomainEventImpl{
		Type:      eventType,
		Data:      entity,
		Timestamp: time.Now(),
		Source:    s.name,
	}

	if err := s.eventPub.PublishAsync(ctx, event); err != nil {
		s.logger.Error("Failed to publish created event",
			logger.String("service", s.name),
			logger.String("event_type", eventType),
			logger.ErrorField(err))
		return fmt.Errorf("publish created event: %w", err)
	}

	s.logger.Info("Published created event",
		logger.String("service", s.name),
		logger.String("event_type", eventType))
	return nil
}

func (s *BusinessServiceImpl[T, ID]) PublishUpdatedEvent(ctx context.Context, old, new *T) error {
	if s.eventPub == nil {
		return nil
	}

	eventType := fmt.Sprintf("%s.updated", s.getEntityTypeName())
	event := &DomainEventImpl{
		Type:      eventType,
		Data:      map[string]*T{"old": old, "new": new},
		Timestamp: time.Now(),
		Source:    s.name,
	}

	if err := s.eventPub.PublishAsync(ctx, event); err != nil {
		s.logger.Error("Failed to publish updated event",
			logger.String("service", s.name),
			logger.String("event_type", eventType),
			logger.ErrorField(err))
		return fmt.Errorf("publish updated event: %w", err)
	}

	s.logger.Info("Published updated event",
		logger.String("service", s.name),
		logger.String("event_type", eventType))
	return nil
}

func (s *BusinessServiceImpl[T, ID]) PublishDeletedEvent(ctx context.Context, entity *T) error {
	if s.eventPub == nil {
		return nil
	}

	eventType := fmt.Sprintf("%s.deleted", s.getEntityTypeName())
	event := &DomainEventImpl{
		Type:      eventType,
		Data:      entity,
		Timestamp: time.Now(),
		Source:    s.name,
	}

	if err := s.eventPub.PublishAsync(ctx, event); err != nil {
		s.logger.Error("Failed to publish deleted event",
			logger.String("service", s.name),
			logger.String("event_type", eventType),
			logger.ErrorField(err))
		return fmt.Errorf("publish deleted event: %w", err)
	}

	s.logger.Info("Published deleted event",
		logger.String("service", s.name),
		logger.String("event_type", eventType))
	return nil
}

// Audit operations (placeholder implementation)
func (s *BusinessServiceImpl[T, ID]) GetAuditLog(ctx context.Context, id ID, req *framework.ListRequest) (*framework.ListResponse[framework.AuditLogEntry], error) {
	// This would typically interact with an audit log repository
	// For now, return empty response
	return &framework.ListResponse[framework.AuditLogEntry]{
		Data:       []*framework.AuditLogEntry{},
		Total:      0,
		Limit:      req.Limit,
		Offset:     req.Offset,
		HasMore:    false,
		TotalPages: 0,
	}, nil
}

// Workflow operations
func (s *BusinessServiceImpl[T, ID]) ProcessWorkflow(ctx context.Context, id ID, action string, params map[string]any) error {
	handler, exists := s.workflows[action]
	if !exists {
		return fmt.Errorf("workflow action '%s' not found", action)
	}

	// Get entity
	entity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get entity for workflow: %w", err)
	}

	// Execute workflow
	if err := handler(ctx, entity, params); err != nil {
		s.logger.Error("Workflow execution failed",
			logger.String("service", s.name),
			logger.Any("id", id),
			logger.String("action", action),
			logger.ErrorField(err))
		return fmt.Errorf("workflow '%s' failed: %w", action, err)
	}

	s.logger.Info("Workflow executed successfully",
		logger.String("service", s.name),
		logger.Any("id", id),
		logger.String("action", action))
	return nil
}

// Override Create to add business validation and event publishing
func (s *BusinessServiceImpl[T, ID]) Create(ctx context.Context, req *framework.CreateRequest[T]) (*T, error) {
	if req == nil || req.Data == nil {
		return nil, fmt.Errorf("create request and data cannot be nil")
	}

	// Business rules validation
	if err := s.ValidateBusinessRules(ctx, req.Data); err != nil {
		return nil, fmt.Errorf("business validation failed: %w", err)
	}

	// Call parent Create (which handles basic validation and persistence)
	entity, err := s.BaseServiceImpl.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	// Process workflows if enabled
	if req.Options != nil && req.Options.ProcessWorkflows {
		if err := s.ProcessWorkflow(ctx, s.getEntityID(entity), "create", req.Metadata); err != nil {
			s.logger.Warn("Create workflow failed", logger.ErrorField(err))
		}
	}

	return entity, nil
}

// Override Update to add business validation
func (s *BusinessServiceImpl[T, ID]) Update(ctx context.Context, id ID, req *framework.UpdateRequest[T]) (*T, error) {
	if req == nil || req.Data == nil {
		return nil, fmt.Errorf("update request and data cannot be nil")
	}

	// Business rules validation for update
	if err := s.ValidateBusinessRulesForUpdate(ctx, id, req); err != nil {
		return nil, fmt.Errorf("business validation failed: %w", err)
	}

	// Call parent Update
	entity, err := s.BaseServiceImpl.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// Process workflows if enabled
	if req.Options != nil && req.Options.ProcessWorkflows {
		if err := s.ProcessWorkflow(ctx, id, "update", req.Metadata); err != nil {
			s.logger.Warn("Update workflow failed", logger.ErrorField(err))
		}
	}

	return entity, nil
}

// Helper methods
func (s *BusinessServiceImpl[T, ID]) getEntityTypeName() string {
	var entity T
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	return entityType.Name()
}

func (s *BusinessServiceImpl[T, ID]) getEntityID(entity *T) ID {
	// Use reflection to get ID field
	entityValue := reflect.ValueOf(entity).Elem()
	idField := entityValue.FieldByName("ID")

	if !idField.IsValid() {
		// Fallback - this should be handled better in real implementation
		var zero ID
		return zero
	}

	// Convert to ID type - this is a simplified approach
	// In real implementation, you'd need proper type conversion based on ID type
	switch idField.Kind() {
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		// For uint-based IDs
		if idValue, ok := any(uint(idField.Uint())).(ID); ok {
			return idValue
		}
	case reflect.Int, reflect.Int32, reflect.Int64:
		// For int-based IDs
		if idValue, ok := any(int(idField.Int())).(ID); ok {
			return idValue
		}
	case reflect.String:
		// For string-based IDs
		if idValue, ok := any(idField.String()).(ID); ok {
			return idValue
		}
	}

	var zero ID
	return zero
}

func (s *BusinessServiceImpl[T, ID]) mergeEntityForValidation(existing, updates *T) *T {
	// This is a simplified merge - in real implementation, you'd use a proper merge strategy
	// For now, return the updates as they should contain the fields being updated
	return updates
}

// DomainEventImpl implements the DomainEvent interface
type DomainEventImpl struct {
	Type      string
	Data      any
	Timestamp time.Time
	Source    string
	ID        string
	Version   string
}

func (e *DomainEventImpl) EventType() string {
	return e.Type
}

func (e *DomainEventImpl) EventData() any {
	return e.Data
}

func (e *DomainEventImpl) EventTime() time.Time {
	return e.Timestamp
}

func (e *DomainEventImpl) EventID() string {
	if e.ID == "" {
		// Generate a simple ID - in real implementation, use UUID or similar
		e.ID = fmt.Sprintf("%s-%d", e.Type, e.Timestamp.UnixNano())
	}
	return e.ID
}

func (e *DomainEventImpl) EventVersion() string {
	if e.Version == "" {
		e.Version = "1.0"
	}
	return e.Version
}

func (e *DomainEventImpl) EventSource() string {
	return e.Source
}
