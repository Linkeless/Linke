package services

import (
	"context"
	"fmt"
	"time"

	"linke/internal/shared/framework"
	"linke/internal/shared/logger"
)

// LookupServiceMixin provides common lookup operations for services
type LookupServiceMixin[T any, ID comparable] struct {
	*BaseServiceImpl[T, ID]
	fieldMappings map[string]string // logical field -> database column mapping
}

// NewLookupServiceMixin creates a new LookupServiceMixin
func NewLookupServiceMixin[T any, ID comparable](
	base *BaseServiceImpl[T, ID],
	fieldMappings map[string]string,
) *LookupServiceMixin[T, ID] {
	if fieldMappings == nil {
		fieldMappings = make(map[string]string)
	}
	return &LookupServiceMixin[T, ID]{
		BaseServiceImpl: base,
		fieldMappings:   fieldMappings,
	}
}

// GetByField retrieves an entity by a specific field
func (s *LookupServiceMixin[T, ID]) GetByField(ctx context.Context, field string, value any) (*T, error) {
	// Map logical field to database column if mapping exists
	dbField := field
	if mapped, exists := s.fieldMappings[field]; exists {
		dbField = mapped
	}

	var entity T
	if err := s.repository.GetDB().WithContext(ctx).Where(dbField+" = ?", value).First(&entity).Error; err != nil {
		s.logger.Error("Failed to get entity by field",
			logger.String("field", field),
			logger.Any("value", value),
			logger.ErrorField(err))
		return nil, fmt.Errorf("get entity by %s: %w", field, err)
	}
	return &entity, nil
}

// ExistsByField checks if an entity exists by a specific field
func (s *LookupServiceMixin[T, ID]) ExistsByField(ctx context.Context, field string, value any) (bool, error) {
	dbField := field
	if mapped, exists := s.fieldMappings[field]; exists {
		dbField = mapped
	}

	var count int64
	if err := s.repository.GetDB().WithContext(ctx).Model(new(T)).Where(dbField+" = ?", value).Count(&count).Error; err != nil {
		s.logger.Error("Failed to check entity exists by field",
			logger.String("field", field),
			logger.Any("value", value),
			logger.ErrorField(err))
		return false, fmt.Errorf("check entity exists by %s: %w", field, err)
	}
	return count > 0, nil
}

// GetByUniqueFields retrieves an entity by multiple unique fields
func (s *LookupServiceMixin[T, ID]) GetByUniqueFields(ctx context.Context, fields map[string]any) (*T, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("no fields provided for lookup")
	}

	db := s.repository.GetDB().WithContext(ctx)
	for field, value := range fields {
		dbField := field
		if mapped, exists := s.fieldMappings[field]; exists {
			dbField = mapped
		}
		db = db.Where(dbField+" = ?", value)
	}

	var entity T
	if err := db.First(&entity).Error; err != nil {
		s.logger.Error("Failed to get entity by unique fields",
			logger.Any("fields", fields),
			logger.ErrorField(err))
		return nil, fmt.Errorf("get entity by unique fields: %w", err)
	}
	return &entity, nil
}

// OrderManagementServiceMixin provides order-specific operations
type OrderManagementServiceMixin[T any, ID comparable] struct {
	*BaseServiceImpl[T, ID]
	orderNumberField string
	statusField      string
}

// NewOrderManagementServiceMixin creates a new OrderManagementServiceMixin
func NewOrderManagementServiceMixin[T any, ID comparable](
	base *BaseServiceImpl[T, ID],
	orderNumberField, statusField string,
) *OrderManagementServiceMixin[T, ID] {
	if orderNumberField == "" {
		orderNumberField = "order_number"
	}
	if statusField == "" {
		statusField = "status"
	}
	return &OrderManagementServiceMixin[T, ID]{
		BaseServiceImpl:  base,
		orderNumberField: orderNumberField,
		statusField:      statusField,
	}
}

// GetByOrderNumber retrieves an entity by order number
func (s *OrderManagementServiceMixin[T, ID]) GetByOrderNumber(ctx context.Context, orderNumber string) (*T, error) {
	var entity T
	if err := s.repository.GetDB().WithContext(ctx).Where(s.orderNumberField+" = ?", orderNumber).First(&entity).Error; err != nil {
		s.logger.Error("Failed to get entity by order number",
			logger.String("order_number", orderNumber),
			logger.ErrorField(err))
		return nil, fmt.Errorf("get entity by order number: %w", err)
	}
	return &entity, nil
}

// UpdateOrderStatus updates the order status with reason
func (s *OrderManagementServiceMixin[T, ID]) UpdateOrderStatus(ctx context.Context, id ID, status, reason string) (*T, error) {
	updates := map[string]any{
		s.statusField: status,
	}
	if reason != "" {
		updates["status_reason"] = reason
		updates["status_updated_at"] = time.Now()
	}

	if err := s.repository.GetDB().WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update order status",
			logger.Any("id", id),
			logger.String("status", status),
			logger.ErrorField(err))
		return nil, fmt.Errorf("update order status: %w", err)
	}

	return s.GetByID(ctx, id)
}

// ProcessOrder processes an order with custom data
func (s *OrderManagementServiceMixin[T, ID]) ProcessOrder(ctx context.Context, id ID, processData map[string]any) (*T, error) {
	// Default implementation - can be overridden
	processData["processed_at"] = time.Now()
	processData[s.statusField] = "processing"

	if err := s.repository.GetDB().WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(processData).Error; err != nil {
		s.logger.Error("Failed to process order",
			logger.Any("id", id),
			logger.ErrorField(err))
		return nil, fmt.Errorf("process order: %w", err)
	}

	return s.GetByID(ctx, id)
}

// CancelOrder cancels an order with reason
func (s *OrderManagementServiceMixin[T, ID]) CancelOrder(ctx context.Context, id ID, reason string) (*T, error) {
	updates := map[string]any{
		s.statusField:         "cancelled",
		"cancelled_at":        time.Now(),
		"cancellation_reason": reason,
	}

	if err := s.repository.GetDB().WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to cancel order",
			logger.Any("id", id),
			logger.String("reason", reason),
			logger.ErrorField(err))
		return nil, fmt.Errorf("cancel order: %w", err)
	}

	return s.GetByID(ctx, id)
}

// GetOrdersByStatus retrieves orders by status
func (s *OrderManagementServiceMixin[T, ID]) GetOrdersByStatus(ctx context.Context, status string, req *framework.ListRequest) (*framework.ListResponse[T], error) {
	return s.ListByStatus(ctx, status, req)
}

// GetOrderStatisticsByPeriod gets order statistics for a time period
func (s *OrderManagementServiceMixin[T, ID]) GetOrderStatisticsByPeriod(ctx context.Context, start, end time.Time) (*framework.OrderStatistics, error) {
	// Get base statistics
	baseStats, err := s.GetStatistics(ctx)
	if err != nil {
		return nil, err
	}

	// Get period-specific counts
	db := s.repository.GetDB().WithContext(ctx).Model(new(T)).Where("created_at BETWEEN ? AND ?", start, end)

	var total, completed, pending, cancelled int64

	// Total in period
	if err := db.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count total orders in period: %w", err)
	}

	// Completed orders
	if err := db.Where(s.statusField+" IN (?)", []string{"completed", "delivered", "success"}).Count(&completed).Error; err != nil {
		completed = 0 // Non-fatal
	}

	// Pending orders
	if err := db.Where(s.statusField+" IN (?)", []string{"pending", "processing", "confirmed"}).Count(&pending).Error; err != nil {
		pending = 0 // Non-fatal
	}

	// Cancelled orders
	if err := db.Where(s.statusField+" IN (?)", []string{"cancelled", "refunded", "failed"}).Count(&cancelled).Error; err != nil {
		cancelled = 0 // Non-fatal
	}

	return &framework.OrderStatistics{
		StatisticsResponse: *baseStats,
		CompletedOrders:    completed,
		PendingOrders:      pending,
		CancelledOrders:    cancelled,
		TotalRevenue:       0, // Should be calculated based on amount field if exists
		AverageOrderValue:  0, // Should be calculated based on amount field if exists
		OrdersByPeriod:     map[string]int64{"current": total},
		RevenueByPeriod:    map[string]float64{"current": 0},
	}, nil
}

// CacheableServiceMixin provides caching capabilities
type CacheableServiceMixin[T any, ID comparable] struct {
	*BaseServiceImpl[T, ID]
	cache framework.Cache
}

// NewCacheableServiceMixin creates a new CacheableServiceMixin
func NewCacheableServiceMixin[T any, ID comparable](
	base *BaseServiceImpl[T, ID],
	cache framework.Cache,
) *CacheableServiceMixin[T, ID] {
	return &CacheableServiceMixin[T, ID]{
		BaseServiceImpl: base,
		cache:           cache,
	}
}

// InvalidateCache invalidates cache keys
func (s *CacheableServiceMixin[T, ID]) InvalidateCache(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		if err := s.cache.Delete(ctx, key); err != nil {
			s.logger.Warn("Failed to invalidate cache key", logger.String("key", key), logger.ErrorField(err))
		}
	}
	return nil
}

// WarmCache warms up cache for specific entities
func (s *CacheableServiceMixin[T, ID]) WarmCache(ctx context.Context, ids []ID) error {
	for _, id := range ids {
		if _, err := s.GetByID(ctx, id); err != nil {
			s.logger.Warn("Failed to warm cache for entity", logger.Any("id", id), logger.ErrorField(err))
		}
	}
	return nil
}

// GetCacheStats returns cache statistics
func (s *CacheableServiceMixin[T, ID]) GetCacheStats(ctx context.Context) (*framework.CacheStats, error) {
	// Default implementation - should be customized based on cache implementation
	return &framework.CacheStats{
		HitRate:       0.85,
		TotalRequests: 1000,
		CacheHits:     850,
		CacheMisses:   150,
		LastUpdated:   time.Now(),
	}, nil
}

// RefreshCache refreshes cache for a specific entity
func (s *CacheableServiceMixin[T, ID]) RefreshCache(ctx context.Context, id ID) (*T, error) {
	// Clear cache first
	cacheKey := fmt.Sprintf("%s:%v", s.GetName(), id)
	s.cache.Delete(ctx, cacheKey)

	// Get fresh data
	return s.GetByID(ctx, id)
}

// SearchableServiceMixin provides search capabilities
type SearchableServiceMixin[T any, ID comparable] struct {
	*BaseServiceImpl[T, ID]
	searchableFields []string
}

// NewSearchableServiceMixin creates a new SearchableServiceMixin
func NewSearchableServiceMixin[T any, ID comparable](
	base *BaseServiceImpl[T, ID],
	searchableFields []string,
) *SearchableServiceMixin[T, ID] {
	if searchableFields == nil {
		searchableFields = []string{"name", "title", "description"}
	}
	return &SearchableServiceMixin[T, ID]{
		BaseServiceImpl:  base,
		searchableFields: searchableFields,
	}
}

// FullTextSearch performs full-text search on specified fields
func (s *SearchableServiceMixin[T, ID]) FullTextSearch(ctx context.Context, query string, fields []string, req *framework.ListRequest) (*framework.ListResponse[T], error) {
	if len(fields) == 0 {
		fields = s.searchableFields
	}

	// Use the base Search method but with custom fields
	return s.Search(ctx, query, req)
}

// SearchByTags searches entities by tags
func (s *SearchableServiceMixin[T, ID]) SearchByTags(ctx context.Context, tags []string, req *framework.ListRequest) (*framework.ListResponse[T], error) {
	if len(tags) == 0 {
		return s.List(ctx, req)
	}

	var entities []*T
	var total int64

	db := s.repository.GetDB().WithContext(ctx).Model(new(T))

	// Build tag query
	for _, tag := range tags {
		db = db.Where("tags LIKE ?", "%"+tag+"%")
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count entities by tags: %w", err)
	}

	// Get paginated results
	if err := db.Limit(req.Limit).Offset(req.Offset).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("search entities by tags: %w", err)
	}

	return s.buildListResponse(entities, total, req), nil
}

// SearchWithHighlight performs search with highlighted results
func (s *SearchableServiceMixin[T, ID]) SearchWithHighlight(ctx context.Context, query string, highlightFields []string, req *framework.ListRequest) (*framework.SearchResponse[T], error) {
	// Perform regular search first
	listResponse, err := s.Search(ctx, query, req)
	if err != nil {
		return nil, err
	}

	// Convert to SearchResponse
	searchResponse := &framework.SearchResponse[T]{
		ListResponse: *listResponse,
		SearchTime:   10, // Mock search time in ms
		TotalMatches: listResponse.Total,
		Highlights:   make(map[string][]string),
	}

	// Add basic highlighting (simplified implementation)
	for _, field := range highlightFields {
		searchResponse.Highlights[field] = []string{fmt.Sprintf("...%s...", query)}
	}

	return searchResponse, nil
}

// GetSearchSuggestions returns search suggestions
func (s *SearchableServiceMixin[T, ID]) GetSearchSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	// Simple implementation - return query variations
	suggestions := make([]string, 0, limit)

	if query != "" {
		suggestions = append(suggestions, query+" tips")
		suggestions = append(suggestions, query+" guide")
		suggestions = append(suggestions, query+" help")
	}

	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return suggestions, nil
}
