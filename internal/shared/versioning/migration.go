package versioning

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"linke/internal/shared/logger"
)

// MigrationFunc represents a function that can migrate data between versions
type MigrationFunc func(from, to Version, data any) (any, error)

// FieldMapping defines how to map fields between versions
type FieldMapping struct {
	FromField    string `json:"from_field"`
	ToField      string `json:"to_field"`
	Transform    string `json:"transform,omitempty"` // transformation type
	DefaultValue any    `json:"default_value,omitempty"`
	Required     bool   `json:"required"`
}

// VersionMigration defines how to migrate between two versions
type VersionMigration struct {
	FromVersion Version        `json:"from_version"`
	ToVersion   Version        `json:"to_version"`
	Mappings    []FieldMapping `json:"mappings"`
	CustomFunc  MigrationFunc  `json:"-"` // custom migration function
}

// MigrationRegistry manages version migrations
type MigrationRegistry struct {
	migrations map[string]VersionMigration // key: "from_version->to_version"
	logger     logger.Logger
}

// NewMigrationRegistry creates a new migration registry
func NewMigrationRegistry(log logger.Logger) *MigrationRegistry {
	return &MigrationRegistry{
		migrations: make(map[string]VersionMigration),
		logger:     log,
	}
}

// migrationKey generates a key for migration lookup
func (mr *MigrationRegistry) migrationKey(from, to Version) string {
	return fmt.Sprintf("%s->%s", from.String(), to.String())
}

// RegisterMigration registers a migration between two versions
func (mr *MigrationRegistry) RegisterMigration(migration VersionMigration) {
	key := mr.migrationKey(migration.FromVersion, migration.ToVersion)
	mr.migrations[key] = migration

	mr.logger.Info("Registered version migration",
		logger.String("from_version", migration.FromVersion.String()),
		logger.String("to_version", migration.ToVersion.String()),
		logger.Int("mapping_count", len(migration.Mappings)),
		logger.Bool("has_custom_func", migration.CustomFunc != nil),
	)
}

// GetMigration retrieves a migration between two versions
func (mr *MigrationRegistry) GetMigration(from, to Version) (*VersionMigration, bool) {
	key := mr.migrationKey(from, to)
	migration, exists := mr.migrations[key]
	return &migration, exists
}

// Migrate performs migration from one version to another
func (mr *MigrationRegistry) Migrate(from, to Version, data any) (any, error) {
	// If versions are the same, return data as-is
	if from.Compare(to) == 0 {
		return data, nil
	}

	migration, exists := mr.GetMigration(from, to)
	if !exists {
		return nil, fmt.Errorf("no migration found from version %s to %s", from.String(), to.String())
	}

	// Use custom migration function if available
	if migration.CustomFunc != nil {
		return migration.CustomFunc(from, to, data)
	}

	// Use field mappings for migration
	return mr.migrateWithMappings(*migration, data)
}

// migrateWithMappings performs migration using field mappings
func (mr *MigrationRegistry) migrateWithMappings(migration VersionMigration, data any) (any, error) {
	// Convert data to map for easier manipulation
	dataMap, err := mr.toMap(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert data to map: %w", err)
	}

	result := make(map[string]any)

	// Apply field mappings
	for _, mapping := range migration.Mappings {
		value, exists := dataMap[mapping.FromField]

		// Handle missing required fields
		if !exists && mapping.Required {
			if mapping.DefaultValue != nil {
				value = mapping.DefaultValue
			} else {
				return nil, fmt.Errorf("required field %s is missing and no default value provided", mapping.FromField)
			}
		}

		// Use default value if field doesn't exist
		if !exists && mapping.DefaultValue != nil {
			value = mapping.DefaultValue
		}

		// Apply transformation if specified
		if mapping.Transform != "" && value != nil {
			transformedValue, err := mr.applyTransform(mapping.Transform, value)
			if err != nil {
				mr.logger.Warn("Failed to apply transformation",
					logger.String("transform", mapping.Transform),
					logger.String("field", mapping.FromField),
					logger.ErrorField(err),
				)
			} else {
				value = transformedValue
			}
		}

		// Set the mapped field
		if value != nil {
			result[mapping.ToField] = value
		}
	}

	// Copy any unmapped fields that don't conflict
	for key, value := range dataMap {
		if _, exists := result[key]; !exists {
			// Check if this field is being renamed
			isRenamed := false
			for _, mapping := range migration.Mappings {
				if mapping.FromField == key {
					isRenamed = true
					break
				}
			}

			// If not renamed, copy the field
			if !isRenamed {
				result[key] = value
			}
		}
	}

	return result, nil
}

// toMap converts any data type to map[string]any
func (mr *MigrationRegistry) toMap(data any) (map[string]any, error) {
	// If already a map, return as-is
	if dataMap, ok := data.(map[string]any); ok {
		return dataMap, nil
	}

	// Convert via JSON marshaling/unmarshaling
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	err = json.Unmarshal(jsonBytes, &result)
	return result, err
}

// applyTransform applies transformation to a value
func (mr *MigrationRegistry) applyTransform(transform string, value any) (any, error) {
	switch transform {
	case "string":
		return fmt.Sprintf("%v", value), nil
	case "int":
		return mr.toInt(value)
	case "float":
		return mr.toFloat(value)
	case "bool":
		return mr.toBool(value)
	case "uppercase":
		if str, ok := value.(string); ok {
			return strings.ToUpper(str), nil
		}
		return value, fmt.Errorf("uppercase transform requires string input")
	case "lowercase":
		if str, ok := value.(string); ok {
			return strings.ToLower(str), nil
		}
		return value, fmt.Errorf("lowercase transform requires string input")
	default:
		return value, fmt.Errorf("unknown transform: %s", transform)
	}
}

// Helper functions for type conversion
func (mr *MigrationRegistry) toInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

func (mr *MigrationRegistry) toFloat(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func (mr *MigrationRegistry) toBool(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// ResponseMigrator handles response migration between versions
type ResponseMigrator struct {
	registry *MigrationRegistry
	logger   logger.Logger
}

// NewResponseMigrator creates a new response migrator
func NewResponseMigrator(registry *MigrationRegistry, log logger.Logger) *ResponseMigrator {
	return &ResponseMigrator{
		registry: registry,
		logger:   log,
	}
}

// MigrateResponse migrates response data to the requested version
func (rm *ResponseMigrator) MigrateResponse(c *gin.Context, data any, responseVersion Version) (any, error) {
	versionCtx, exists := GetVersionFromContext(c)
	if !exists {
		return data, nil // No version context, return data as-is
	}

	requestedVersion := versionCtx.RequestedVersion

	// If versions match, no migration needed
	if responseVersion.Compare(requestedVersion) == 0 {
		return data, nil
	}

	// Perform migration
	migratedData, err := rm.registry.Migrate(responseVersion, requestedVersion, data)
	if err != nil {
		rm.logger.Error("Response migration failed",
			logger.String("from_version", responseVersion.String()),
			logger.String("to_version", requestedVersion.String()),
			logger.ErrorField(err),
		)
		return nil, err
	}

	// Add migration headers
	c.Header("X-API-Response-Migrated-From", responseVersion.String())
	c.Header("X-API-Response-Migrated-To", requestedVersion.String())

	rm.logger.Debug("Response migration successful",
		logger.String("from_version", responseVersion.String()),
		logger.String("to_version", requestedVersion.String()),
	)

	return migratedData, nil
}

// Middleware returns a Gin middleware that automatically migrates responses
func (rm *ResponseMigrator) Middleware(responseVersion Version) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Continue with the request
		c.Next()

		// Check if response needs migration
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// Note: This is a placeholder for response migration
			// In practice, you would need to capture and migrate the response body
			// which requires more complex middleware implementation

			c.Header("X-API-Response-Version", responseVersion.String())
		}
	}
}

// PrebuiltMigrations returns commonly used migrations
func PrebuiltMigrations(log logger.Logger) map[string]VersionMigration {
	v1 := NewVersion(1, 0, 0)
	v2 := NewVersion(2, 0, 0)

	return map[string]VersionMigration{
		"user_v1_to_v2": {
			FromVersion: v1,
			ToVersion:   v2,
			Mappings: []FieldMapping{
				{FromField: "name", ToField: "full_name", Required: true},
				{FromField: "email", ToField: "email", Required: true},
				{FromField: "id", ToField: "user_id", Required: true},
				{FromField: "", ToField: "created_at", DefaultValue: "2024-01-01T00:00:00Z"},
			},
		},
		"subscription_v1_to_v2": {
			FromVersion: v1,
			ToVersion:   v2,
			Mappings: []FieldMapping{
				{FromField: "plan_id", ToField: "subscription_plan_id", Required: true},
				{FromField: "status", ToField: "subscription_status", Required: true},
				{FromField: "expires_at", ToField: "expiry_date", Required: true},
				{FromField: "", ToField: "billing_cycle", DefaultValue: "monthly"},
			},
		},
	}
}

// AutoMigrationBuilder helps build migrations automatically based on struct tags
type AutoMigrationBuilder struct {
	logger logger.Logger
}

// NewAutoMigrationBuilder creates a new auto migration builder
func NewAutoMigrationBuilder(log logger.Logger) *AutoMigrationBuilder {
	return &AutoMigrationBuilder{
		logger: log,
	}
}

// BuildMigration builds migration mappings from struct tags
func (amb *AutoMigrationBuilder) BuildMigration(fromStruct, toStruct any, fromVersion, toVersion Version) VersionMigration {
	mappings := amb.extractMappings(fromStruct, toStruct)

	return VersionMigration{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Mappings:    mappings,
	}
}

// extractMappings extracts field mappings from struct tags
func (amb *AutoMigrationBuilder) extractMappings(fromStruct, toStruct any) []FieldMapping {
	var mappings []FieldMapping

	fromType := reflect.TypeOf(fromStruct)
	toType := reflect.TypeOf(toStruct)

	if fromType.Kind() == reflect.Ptr {
		fromType = fromType.Elem()
	}
	if toType.Kind() == reflect.Ptr {
		toType = toType.Elem()
	}

	// Map fields from source struct
	for i := 0; i < fromType.NumField(); i++ {
		fromField := fromType.Field(i)
		jsonTag := fromField.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fromFieldName := strings.Split(jsonTag, ",")[0]

		// Find corresponding field in target struct
		for j := 0; j < toType.NumField(); j++ {
			toField := toType.Field(j)
			toJsonTag := toField.Tag.Get("json")
			if toJsonTag == "" || toJsonTag == "-" {
				continue
			}

			toFieldName := strings.Split(toJsonTag, ",")[0]

			// Check if fields match by name or migration tag
			if fromFieldName == toFieldName || fromField.Tag.Get("migrate") == toFieldName {
				mapping := FieldMapping{
					FromField: fromFieldName,
					ToField:   toFieldName,
					Required:  strings.Contains(toField.Tag.Get("json"), "required"),
				}

				// Add transformation if specified
				if transform := fromField.Tag.Get("transform"); transform != "" {
					mapping.Transform = transform
				}

				mappings = append(mappings, mapping)
				break
			}
		}
	}

	return mappings
}
