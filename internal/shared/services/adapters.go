package services

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"linke/internal/shared/framework"
	"linke/internal/shared/logger"
)

// ServiceAdapter helps migrate legacy domain services to use generic frameworks
// while maintaining backward compatibility
type ServiceAdapter[T any, ID comparable] struct {
	*BaseServiceImpl[T, ID]
	legacyService any // The original service implementation
	methodMapping map[string]string // legacy method -> generic method mapping
}

// NewServiceAdapter creates a new ServiceAdapter
func NewServiceAdapter[T any, ID comparable](
	base *BaseServiceImpl[T, ID],
	legacyService any,
	methodMapping map[string]string,
) *ServiceAdapter[T, ID] {
	if methodMapping == nil {
		methodMapping = getDefaultMethodMapping()
	}
	
	return &ServiceAdapter[T, ID]{
		BaseServiceImpl: base,
		legacyService:   legacyService,
		methodMapping:   methodMapping,
	}
}

// getDefaultMethodMapping returns common legacy -> generic method mappings
func getDefaultMethodMapping() map[string]string {
	return map[string]string{
		"CreateUser":           "Create",
		"GetUserByID":          "GetByID", 
		"UpdateUser":           "Update",
		"DeleteUser":           "Delete",
		"SoftDeleteUser":       "SoftDelete",
		"RestoreUser":          "Restore",
		"HardDeleteUser":       "HardDelete",
		"ListUsers":            "List",
		"ListDeletedUsers":     "ListDeleted",
		"SearchUsers":          "Search",
		"UpdateUserStatus":     "UpdateStatus",
		"GetUserStats":         "GetStatistics",
		"BatchDeleteUsers":     "BatchDelete",
		"BatchRestoreUsers":    "BatchRestore",
		
		// Common patterns for other domains
		"CreateTicket":         "Create",
		"GetTicketByID":        "GetByID",
		"UpdateTicket":         "Update",
		"DeleteTicket":         "Delete",
		"ListTickets":          "List",
		"GetTickets":           "ListWithFilters",
		"UpdateTicketStatus":   "UpdateStatus",
		
		"CreateSubscriptionOrder": "Create",
		"GetSubscriptionOrder":    "GetByID",
		"CancelSubscriptionOrder": "Delete",
		"GetSubscriptionOrders":   "ListWithFilters",
		
		"CreatePaymentOrder":      "Create",
		"GetPaymentRecord":        "GetByID",
		"UpdatePaymentStatus":     "UpdateStatus",
		"GetUserPaymentRecords":   "ListByUser",
	}
}

// CallLegacyMethod dynamically calls legacy methods for methods not yet migrated
func (s *ServiceAdapter[T, ID]) CallLegacyMethod(methodName string, args ...any) (any, error) {
	legacyValue := reflect.ValueOf(s.legacyService)
	method := legacyValue.MethodByName(methodName)
	
	if !method.IsValid() {
		return nil, fmt.Errorf("legacy method %s not found", methodName)
	}

	// Convert args to reflect.Value
	argValues := make([]reflect.Value, len(args))
	for i, arg := range args {
		argValues[i] = reflect.ValueOf(arg)
	}

	// Call the method
	results := method.Call(argValues)
	
	// Handle results (simplified error handling)
	if len(results) == 0 {
		return nil, nil
	}
	
	// Check if last result is error
	if len(results) > 1 {
		lastResult := results[len(results)-1]
		if lastResult.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !lastResult.IsNil() {
				return results[0].Interface(), lastResult.Interface().(error)
			}
		}
	}

	return results[0].Interface(), nil
}

// DomainServiceWrapper wraps domain-specific services with generic capabilities
type DomainServiceWrapper[T any, ID comparable] struct {
	*BaseServiceImpl[T, ID]
	domainService any
	entityType    reflect.Type
	
	// Configuration for domain-specific behaviors
	config *DomainServiceConfig
}

// DomainServiceConfig configures domain-specific behaviors
type DomainServiceConfig struct {
	// Field mappings for lookups
	FieldMappings map[string]string `json:"field_mappings"`
	
	// Search configuration
	SearchableFields []string `json:"searchable_fields"`
	
	// Cache configuration
	CacheEnabled      bool              `json:"cache_enabled"`
	CacheTTL         map[string]int     `json:"cache_ttl"` // method -> TTL in seconds
	CacheKeyPatterns map[string]string  `json:"cache_key_patterns"`
	
	// Order management (if applicable)
	OrderNumberField string `json:"order_number_field"`
	StatusField      string `json:"status_field"`
	
	// Event configuration  
	EventsEnabled    bool     `json:"events_enabled"`
	EventTypes       []string `json:"event_types"`
	
	// Validation rules
	ValidationRules map[string]any `json:"validation_rules"`
	
	// Business logic hooks
	BeforeCreate []string `json:"before_create"`
	AfterCreate  []string `json:"after_create"`
	BeforeUpdate []string `json:"before_update"`
	AfterUpdate  []string `json:"after_update"`
	BeforeDelete []string `json:"before_delete"`
	AfterDelete  []string `json:"after_delete"`
}

// NewDomainServiceWrapper creates a new DomainServiceWrapper
func NewDomainServiceWrapper[T any, ID comparable](
	base *BaseServiceImpl[T, ID],
	domainService any,
	config *DomainServiceConfig,
) *DomainServiceWrapper[T, ID] {
	if config == nil {
		config = &DomainServiceConfig{
			FieldMappings:    make(map[string]string),
			SearchableFields: []string{"name", "title", "description"},
			CacheEnabled:     true,
			CacheTTL:         make(map[string]int),
			CacheKeyPatterns: make(map[string]string),
			EventsEnabled:    true,
			ValidationRules:  make(map[string]any),
		}
	}

	var entityType reflect.Type
	if domainService != nil {
		entityType = reflect.TypeOf(domainService)
	}

	return &DomainServiceWrapper[T, ID]{
		BaseServiceImpl: base,
		domainService:   domainService,
		entityType:      entityType,
		config:          config,
	}
}

// Enhanced Create with domain-specific logic
func (w *DomainServiceWrapper[T, ID]) CreateWithDomainLogic(ctx context.Context, req *framework.CreateRequest[T]) (*T, error) {
	// Execute before-create hooks
	if err := w.executeHooks(ctx, "before_create", req.Data); err != nil {
		return nil, fmt.Errorf("before create hooks failed: %w", err)
	}

	// Call base create
	entity, err := w.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	// Execute after-create hooks  
	if err := w.executeHooks(ctx, "after_create", entity); err != nil {
		w.logger.Warn("After create hooks failed", logger.ErrorField(err))
	}

	return entity, nil
}

// Enhanced Update with domain-specific logic
func (w *DomainServiceWrapper[T, ID]) UpdateWithDomainLogic(ctx context.Context, id ID, req *framework.UpdateRequest[T]) (*T, error) {
	// Execute before-update hooks
	if err := w.executeHooks(ctx, "before_update", req.Data); err != nil {
		return nil, fmt.Errorf("before update hooks failed: %w", err)
	}

	// Call base update
	entity, err := w.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// Execute after-update hooks
	if err := w.executeHooks(ctx, "after_update", entity); err != nil {
		w.logger.Warn("After update hooks failed", logger.ErrorField(err))
	}

	return entity, nil
}

// executeHooks executes configured hooks
func (w *DomainServiceWrapper[T, ID]) executeHooks(ctx context.Context, hookType string, data any) error {
	var hooks []string
	switch hookType {
	case "before_create":
		hooks = w.config.BeforeCreate
	case "after_create":
		hooks = w.config.AfterCreate
	case "before_update":
		hooks = w.config.BeforeUpdate
	case "after_update":
		hooks = w.config.AfterUpdate
	case "before_delete":
		hooks = w.config.BeforeDelete
	case "after_delete":
		hooks = w.config.AfterDelete
	}

	for _, hook := range hooks {
		if err := w.executeHook(ctx, hook, data); err != nil {
			return fmt.Errorf("hook %s failed: %w", hook, err)
		}
	}
	return nil
}

// executeHook executes a single hook
func (w *DomainServiceWrapper[T, ID]) executeHook(ctx context.Context, hook string, data any) error {
	// Try to call hook method on domain service
	if w.domainService != nil {
		serviceValue := reflect.ValueOf(w.domainService)
		method := serviceValue.MethodByName(hook)
		
		if method.IsValid() {
			args := []reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(data),
			}
			
			results := method.Call(args)
			if len(results) > 0 && results[0].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				if !results[0].IsNil() {
					return results[0].Interface().(error)
				}
			}
		}
	}
	return nil
}

// GetDomainMethodNames returns available methods on the domain service
func (w *DomainServiceWrapper[T, ID]) GetDomainMethodNames() []string {
	if w.domainService == nil {
		return nil
	}

	serviceType := reflect.TypeOf(w.domainService)
	var methods []string
	
	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)
		if method.IsExported() {
			methods = append(methods, method.Name)
		}
	}
	
	return methods
}

// MigrationHelper provides utilities for migrating services
type MigrationHelper struct {
	logger framework.Logger
}

// NewMigrationHelper creates a new MigrationHelper
func NewMigrationHelper(logger framework.Logger) *MigrationHelper {
	return &MigrationHelper{logger: logger}
}

// AnalyzeService analyzes a service and provides migration recommendations
func (h *MigrationHelper) AnalyzeService(service any) *ServiceAnalysis {
	serviceType := reflect.TypeOf(service)
	analysis := &ServiceAnalysis{
		ServiceName:      serviceType.Name(),
		Methods:          []MethodInfo{},
		Recommendations: []string{},
	}

	// Analyze methods
	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)
		if !method.IsExported() {
			continue
		}

		methodInfo := MethodInfo{
			Name:         method.Name,
			NumArgs:      method.Type.NumIn(),
			NumReturns:   method.Type.NumOut(),
			IsGenericCandidate: h.isGenericCandidate(method.Name),
		}

		analysis.Methods = append(analysis.Methods, methodInfo)

		// Add recommendations
		if h.isGenericCandidate(method.Name) {
			analysis.Recommendations = append(analysis.Recommendations,
				fmt.Sprintf("Method '%s' can be replaced with generic method", method.Name))
		}
	}

	// Count potential migrations
	genericCandidates := 0
	for _, method := range analysis.Methods {
		if method.IsGenericCandidate {
			genericCandidates++
		}
	}

	if genericCandidates > 0 {
		analysis.Recommendations = append(analysis.Recommendations,
			fmt.Sprintf("Total %d methods can be migrated to generic framework", genericCandidates))
	}

	return analysis
}

// isGenericCandidate checks if a method can be replaced with generic framework
func (h *MigrationHelper) isGenericCandidate(methodName string) bool {
	genericPatterns := []string{
		"Create", "Get.*ByID", "Update", "Delete", "SoftDelete", "Restore", "HardDelete",
		"List.*", "Search.*", "Count.*", "Exists.*", "Batch.*", "UpdateStatus",
	}

	for _, pattern := range genericPatterns {
		if matched, _ := regexp.MatchString("^"+pattern, methodName); matched {
			return true
		}
	}
	return false
}

// ServiceAnalysis represents the analysis result of a service
type ServiceAnalysis struct {
	ServiceName     string       `json:"service_name"`
	Methods         []MethodInfo `json:"methods"`
	Recommendations []string     `json:"recommendations"`
}

// MethodInfo represents information about a service method
type MethodInfo struct {
	Name               string `json:"name"`
	NumArgs            int    `json:"num_args"`
	NumReturns         int    `json:"num_returns"`
	IsGenericCandidate bool   `json:"is_generic_candidate"`
}

// GenerateMigrationPlan generates a migration plan for a service
func (h *MigrationHelper) GenerateMigrationPlan(analysis *ServiceAnalysis) *MigrationPlan {
	plan := &MigrationPlan{
		ServiceName: analysis.ServiceName,
		Phases:      []MigrationPhase{},
	}

	// Phase 1: Core CRUD operations
	phase1 := MigrationPhase{
		Name:        "Core CRUD Migration",
		Description: "Migrate basic CRUD operations to generic framework",
		Priority:    "High",
		Risk:        "Low",
		Methods:     []string{},
	}

	crudMethods := []string{"Create", "GetByID", "Update", "Delete"}
	for _, method := range analysis.Methods {
		for _, crud := range crudMethods {
			if strings.HasPrefix(method.Name, crud) {
				phase1.Methods = append(phase1.Methods, method.Name)
				break
			}
		}
	}

	if len(phase1.Methods) > 0 {
		plan.Phases = append(plan.Phases, phase1)
	}

	// Phase 2: List and Search operations
	phase2 := MigrationPhase{
		Name:        "List and Search Migration", 
		Description: "Migrate list and search operations",
		Priority:    "Medium",
		Risk:        "Low",
		Methods:     []string{},
	}

	listMethods := []string{"List", "Search", "Count"}
	for _, method := range analysis.Methods {
		for _, list := range listMethods {
			if strings.HasPrefix(method.Name, list) {
				phase2.Methods = append(phase2.Methods, method.Name)
				break
			}
		}
	}

	if len(phase2.Methods) > 0 {
		plan.Phases = append(plan.Phases, phase2)
	}

	// Phase 3: Domain-specific operations
	phase3 := MigrationPhase{
		Name:        "Domain-Specific Migration",
		Description: "Migrate domain-specific operations using mixins",
		Priority:    "Low",
		Risk:        "Medium", 
		Methods:     []string{},
	}

	// Add remaining methods
	migratedMethods := make(map[string]bool)
	for _, phase := range plan.Phases {
		for _, method := range phase.Methods {
			migratedMethods[method] = true
		}
	}

	for _, method := range analysis.Methods {
		if !migratedMethods[method.Name] && method.IsGenericCandidate {
			phase3.Methods = append(phase3.Methods, method.Name)
		}
	}

	if len(phase3.Methods) > 0 {
		plan.Phases = append(plan.Phases, phase3)
	}

	return plan
}

// MigrationPlan represents a plan for migrating a service
type MigrationPlan struct {
	ServiceName string           `json:"service_name"`
	Phases      []MigrationPhase `json:"phases"`
}

// MigrationPhase represents a phase in the migration plan
type MigrationPhase struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"` // High, Medium, Low
	Risk        string   `json:"risk"`     // Low, Medium, High
	Methods     []string `json:"methods"`
}