package versioning

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"linke/internal/shared/logger"
)

// EventVersionManager handles event schema versioning and migration
type EventVersionManager struct {
	schemas   map[string]map[string]*EventSchema
	migrators map[string]map[string]EventMigrator
	logger    logger.Logger
}

// EventSchema defines the structure and validation rules for an event version
type EventSchema struct {
	EventType    string                 `json:"event_type"`
	Version      string                 `json:"version"`
	Fields       map[string]FieldSchema `json:"fields"`
	Required     []string               `json:"required"`
	Deprecated   bool                   `json:"deprecated"`
	DeprecatedAt *time.Time             `json:"deprecated_at,omitempty"`
	NextVersion  string                 `json:"next_version,omitempty"`
}

// FieldSchema defines validation rules for event fields
type FieldSchema struct {
	Type        string       `json:"type"` // string, number, boolean, object, array
	Required    bool         `json:"required"`
	Deprecated  bool         `json:"deprecated"`
	Default     any          `json:"default,omitempty"`
	Description string       `json:"description,omitempty"`
	Constraints *Constraints `json:"constraints,omitempty"`
}

// Constraints define validation constraints for fields
type Constraints struct {
	MinLength *int     `json:"min_length,omitempty"`
	MaxLength *int     `json:"max_length,omitempty"`
	Pattern   *string  `json:"pattern,omitempty"`
	Enum      []string `json:"enum,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
}

// EventMigrator handles migration between event versions
type EventMigrator interface {
	Migrate(ctx context.Context, oldData map[string]any) (map[string]any, error)
	SourceVersion() string
	TargetVersion() string
}

// NewEventVersionManager creates a new event version manager
func NewEventVersionManager() *EventVersionManager {
	return &EventVersionManager{
		schemas:   make(map[string]map[string]*EventSchema),
		migrators: make(map[string]map[string]EventMigrator),
		logger:    logger.GetGlobalLogger(),
	}
}

// RegisterSchema registers an event schema for a specific version
func (vm *EventVersionManager) RegisterSchema(schema *EventSchema) error {
	if _, exists := vm.schemas[schema.EventType]; !exists {
		vm.schemas[schema.EventType] = make(map[string]*EventSchema)
	}

	vm.schemas[schema.EventType][schema.Version] = schema

	vm.logger.Info("Event schema registered",
		logger.String("event_type", schema.EventType),
		logger.String("version", schema.Version),
		logger.Any("deprecated", schema.Deprecated),
	)

	return nil
}

// RegisterMigrator registers a migrator between two versions
func (vm *EventVersionManager) RegisterMigrator(eventType string, migrator EventMigrator) error {
	if _, exists := vm.migrators[eventType]; !exists {
		vm.migrators[eventType] = make(map[string]EventMigrator)
	}

	key := fmt.Sprintf("%s-%s", migrator.SourceVersion(), migrator.TargetVersion())
	vm.migrators[eventType][key] = migrator

	vm.logger.Info("Event migrator registered",
		logger.String("event_type", eventType),
		logger.String("source_version", migrator.SourceVersion()),
		logger.String("target_version", migrator.TargetVersion()),
	)

	return nil
}

// ValidateEvent validates an event against its schema
func (vm *EventVersionManager) ValidateEvent(eventType, version string, data map[string]any) error {
	schema, err := vm.GetSchema(eventType, version)
	if err != nil {
		return fmt.Errorf("failed to get schema: %w", err)
	}

	return vm.validateAgainstSchema(data, schema)
}

// GetSchema retrieves a schema for a specific event type and version
func (vm *EventVersionManager) GetSchema(eventType, version string) (*EventSchema, error) {
	eventSchemas, exists := vm.schemas[eventType]
	if !exists {
		return nil, fmt.Errorf("no schemas found for event type: %s", eventType)
	}

	schema, exists := eventSchemas[version]
	if !exists {
		return nil, fmt.Errorf("schema not found for event type %s version %s", eventType, version)
	}

	return schema, nil
}

// GetLatestVersion returns the latest version for an event type
func (vm *EventVersionManager) GetLatestVersion(eventType string) (string, error) {
	eventSchemas, exists := vm.schemas[eventType]
	if !exists {
		return "", fmt.Errorf("no schemas found for event type: %s", eventType)
	}

	var latestVersion string
	var latestVersionNum float64

	for version := range eventSchemas {
		versionNum, err := parseVersion(version)
		if err != nil {
			continue // Skip invalid versions
		}

		if versionNum > latestVersionNum {
			latestVersionNum = versionNum
			latestVersion = version
		}
	}

	if latestVersion == "" {
		return "", fmt.Errorf("no valid versions found for event type: %s", eventType)
	}

	return latestVersion, nil
}

// MigrateEvent migrates an event from one version to another
func (vm *EventVersionManager) MigrateEvent(ctx context.Context, eventType, sourceVersion, targetVersion string, data map[string]any) (map[string]any, error) {
	// If versions are the same, no migration needed
	if sourceVersion == targetVersion {
		return data, nil
	}

	// Find migration path
	migrationPath, err := vm.findMigrationPath(eventType, sourceVersion, targetVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to find migration path: %w", err)
	}

	// Apply migrations in sequence
	currentData := data
	for _, migrator := range migrationPath {
		currentData, err = migrator.Migrate(ctx, currentData)
		if err != nil {
			return nil, fmt.Errorf("migration failed from %s to %s: %w",
				migrator.SourceVersion(), migrator.TargetVersion(), err)
		}

		vm.logger.Debug("Event migrated",
			logger.String("event_type", eventType),
			logger.String("from_version", migrator.SourceVersion()),
			logger.String("to_version", migrator.TargetVersion()),
		)
	}

	// Validate migrated data against target schema
	if err := vm.ValidateEvent(eventType, targetVersion, currentData); err != nil {
		return nil, fmt.Errorf("migrated event validation failed: %w", err)
	}

	return currentData, nil
}

// IsVersionDeprecated checks if a version is deprecated
func (vm *EventVersionManager) IsVersionDeprecated(eventType, version string) bool {
	schema, err := vm.GetSchema(eventType, version)
	if err != nil {
		return false
	}

	return schema.Deprecated
}

// GetCompatibleVersions returns all versions compatible with the given version
func (vm *EventVersionManager) GetCompatibleVersions(eventType, version string) ([]string, error) {
	eventSchemas, exists := vm.schemas[eventType]
	if !exists {
		return nil, fmt.Errorf("no schemas found for event type: %s", eventType)
	}

	var compatibleVersions []string

	// For now, we consider all versions compatible if we can migrate between them
	// In a more sophisticated implementation, you might have explicit compatibility rules
	for v := range eventSchemas {
		if v == version {
			compatibleVersions = append(compatibleVersions, v)
			continue
		}

		// Check if we can migrate to this version
		_, err := vm.findMigrationPath(eventType, version, v)
		if err == nil {
			compatibleVersions = append(compatibleVersions, v)
		}
	}

	return compatibleVersions, nil
}

// findMigrationPath finds a sequence of migrators to go from source to target version
func (vm *EventVersionManager) findMigrationPath(eventType, sourceVersion, targetVersion string) ([]EventMigrator, error) {
	migrators, exists := vm.migrators[eventType]
	if !exists {
		return nil, fmt.Errorf("no migrators found for event type: %s", eventType)
	}

	// Simple direct migration first
	key := fmt.Sprintf("%s-%s", sourceVersion, targetVersion)
	if migrator, exists := migrators[key]; exists {
		return []EventMigrator{migrator}, nil
	}

	// For now, we only support direct migrations
	// In a more sophisticated implementation, you could implement graph traversal
	// to find multi-step migration paths
	return nil, fmt.Errorf("no migration path found from %s to %s for event type %s",
		sourceVersion, targetVersion, eventType)
}

// validateAgainstSchema validates data against a schema
func (vm *EventVersionManager) validateAgainstSchema(data map[string]any, schema *EventSchema) error {
	// Check required fields
	for _, requiredField := range schema.Required {
		if _, exists := data[requiredField]; !exists {
			return fmt.Errorf("required field missing: %s", requiredField)
		}
	}

	// Validate each field
	for fieldName, fieldValue := range data {
		fieldSchema, exists := schema.Fields[fieldName]
		if !exists {
			// Unknown field - log warning but don't fail validation
			vm.logger.Warn("Unknown field in event data",
				logger.String("event_type", schema.EventType),
				logger.String("version", schema.Version),
				logger.String("field", fieldName),
			)
			continue
		}

		if err := vm.validateField(fieldName, fieldValue, fieldSchema); err != nil {
			return fmt.Errorf("field validation failed for %s: %w", fieldName, err)
		}
	}

	return nil
}

// validateField validates a single field against its schema
func (vm *EventVersionManager) validateField(fieldName string, value any, schema FieldSchema) error {
	// Type validation
	if err := vm.validateFieldType(value, schema.Type); err != nil {
		return fmt.Errorf("type validation failed: %w", err)
	}

	// Constraint validation
	if schema.Constraints != nil {
		if err := vm.validateConstraints(value, schema.Constraints); err != nil {
			return fmt.Errorf("constraint validation failed: %w", err)
		}
	}

	return nil
}

// validateFieldType validates the type of a field value
func (vm *EventVersionManager) validateFieldType(value any, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number":
		switch value.(type) {
		case int, int64, float64:
			// Valid number types
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	case "array":
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	default:
		return fmt.Errorf("unknown field type: %s", expectedType)
	}

	return nil
}

// validateConstraints validates field constraints
func (vm *EventVersionManager) validateConstraints(value any, constraints *Constraints) error {
	// String constraints
	if str, ok := value.(string); ok {
		if constraints.MinLength != nil && len(str) < *constraints.MinLength {
			return fmt.Errorf("string too short, minimum length: %d", *constraints.MinLength)
		}
		if constraints.MaxLength != nil && len(str) > *constraints.MaxLength {
			return fmt.Errorf("string too long, maximum length: %d", *constraints.MaxLength)
		}
		if constraints.Enum != nil {
			found := false
			for _, enumValue := range constraints.Enum {
				if str == enumValue {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("value not in enum: %v", constraints.Enum)
			}
		}
	}

	// Number constraints
	if num, ok := value.(float64); ok {
		if constraints.Min != nil && num < *constraints.Min {
			return fmt.Errorf("number too small, minimum: %f", *constraints.Min)
		}
		if constraints.Max != nil && num > *constraints.Max {
			return fmt.Errorf("number too large, maximum: %f", *constraints.Max)
		}
	}

	return nil
}

// parseVersion parses a version string to a comparable number
func parseVersion(version string) (float64, error) {
	// Simple version parsing - assumes format like "1.0", "2.1", etc.
	parts := strings.Split(version, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid version format: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	return float64(major) + float64(minor)/10.0, nil
}

// EventVersioningMiddleware provides middleware for automatic event versioning
type EventVersioningMiddleware struct {
	versionManager *EventVersionManager
	targetVersion  string // Version to migrate all events to
}

// NewEventVersioningMiddleware creates a new event versioning middleware
func NewEventVersioningMiddleware(versionManager *EventVersionManager, targetVersion string) *EventVersioningMiddleware {
	return &EventVersioningMiddleware{
		versionManager: versionManager,
		targetVersion:  targetVersion,
	}
}

// Process processes events with automatic versioning
func (m *EventVersioningMiddleware) Process(ctx context.Context, event any, next func(context.Context, any) error) error {
	// This would integrate with the event system to automatically handle versioning
	// For now, it's a placeholder for the concept
	return next(ctx, event)
}

// EventSchemaRegistry provides a centralized registry for event schemas
type EventSchemaRegistry struct {
	versionManager *EventVersionManager
}

// NewEventSchemaRegistry creates a new event schema registry
func NewEventSchemaRegistry() *EventSchemaRegistry {
	return &EventSchemaRegistry{
		versionManager: NewEventVersionManager(),
	}
}

// RegisterDefaultSchemas registers default schemas for the Linke platform
func (r *EventSchemaRegistry) RegisterDefaultSchemas() error {
	schemas := []*EventSchema{
		// User events
		{
			EventType: "user.created",
			Version:   "1.0",
			Fields: map[string]FieldSchema{
				"user_id":    {Type: "number", Required: true},
				"email":      {Type: "string", Required: true},
				"username":   {Type: "string", Required: true},
				"created_at": {Type: "string", Required: true},
				"status":     {Type: "string", Required: true, Constraints: &Constraints{Enum: []string{"active", "inactive", "suspended"}}},
			},
			Required: []string{"user_id", "email", "username", "created_at", "status"},
		},
		// Payment events
		{
			EventType: "payment.completed",
			Version:   "1.0",
			Fields: map[string]FieldSchema{
				"payment_id": {Type: "string", Required: true},
				"user_id":    {Type: "number", Required: true},
				"order_id":   {Type: "number", Required: true},
				"amount":     {Type: "number", Required: true, Constraints: &Constraints{Min: func() *float64 { v := 0.0; return &v }()}},
				"currency":   {Type: "string", Required: true},
				"method":     {Type: "string", Required: true},
				"gateway":    {Type: "string", Required: true},
				"status":     {Type: "string", Required: true, Constraints: &Constraints{Enum: []string{"completed", "failed", "pending", "refunded"}}},
			},
			Required: []string{"payment_id", "user_id", "order_id", "amount", "currency", "method", "gateway", "status"},
		},
		// Subscription events
		{
			EventType: "subscription.created",
			Version:   "1.0",
			Fields: map[string]FieldSchema{
				"subscription_id": {Type: "number", Required: true},
				"user_id":         {Type: "number", Required: true},
				"plan_id":         {Type: "number", Required: true},
				"status":          {Type: "string", Required: true, Constraints: &Constraints{Enum: []string{"pending", "active", "paused", "cancelled", "expired"}}},
				"start_date":      {Type: "string", Required: true},
				"end_date":        {Type: "string", Required: false},
				"billing_cycle":   {Type: "string", Required: true, Constraints: &Constraints{Enum: []string{"monthly", "yearly", "one-time"}}},
			},
			Required: []string{"subscription_id", "user_id", "plan_id", "status", "start_date", "billing_cycle"},
		},
	}

	for _, schema := range schemas {
		if err := r.versionManager.RegisterSchema(schema); err != nil {
			return fmt.Errorf("failed to register schema for %s v%s: %w", schema.EventType, schema.Version, err)
		}
	}

	return nil
}

// GetVersionManager returns the underlying version manager
func (r *EventSchemaRegistry) GetVersionManager() *EventVersionManager {
	return r.versionManager
}
