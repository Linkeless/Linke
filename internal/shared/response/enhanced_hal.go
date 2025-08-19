package response

import (
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// EnhancedHALCollection represents an enhanced HAL collection with additional features
type EnhancedHALCollection struct {
	Embedded   map[string]any    `json:"_embedded"`
	Links      HALLinks          `json:"_links"`
	Page       PageInfo          `json:"page"`
	Total      int64             `json:"total"`
	Metadata   CollectionMeta    `json:"_metadata,omitempty"`
	Facets     []Facet           `json:"facets,omitempty"`
	Operations []AllowedOp       `json:"_operations,omitempty"`
}

// CollectionMeta provides additional metadata about the collection
type CollectionMeta struct {
	Generated   time.Time              `json:"generated"`
	Query       map[string]interface{} `json:"query,omitempty"`
	Sort        SortInfo               `json:"sort,omitempty"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	Performance PerformanceMeta        `json:"performance,omitempty"`
}

// SortInfo describes the current sorting applied
type SortInfo struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
	Available []SortOption `json:"available,omitempty"`
}

// SortOption describes available sorting options
type SortOption struct {
	Field       string `json:"field"`
	Label       string `json:"label"`
	Direction   string `json:"direction"`
	Default     bool   `json:"default,omitempty"`
}

// Facet represents a facet for filtering
type Facet struct {
	Name    string       `json:"name"`
	Label   string       `json:"label"`
	Type    string       `json:"type"`
	Values  []FacetValue `json:"values"`
	Applied []string     `json:"applied,omitempty"`
}

// FacetValue represents a value in a facet
type FacetValue struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count int64  `json:"count"`
	URL   string `json:"url,omitempty"`
}

// AllowedOp represents an allowed operation on the collection
type AllowedOp struct {
	Name   string            `json:"name"`
	Method string            `json:"method"`
	URL    string            `json:"url"`
	Schema map[string]interface{} `json:"schema,omitempty"`
}

// PerformanceMeta provides performance information
type PerformanceMeta struct {
	QueryTime   time.Duration `json:"query_time"`
	TotalTime   time.Duration `json:"total_time"`
	CacheHit    bool          `json:"cache_hit,omitempty"`
	IndexUsed   []string      `json:"index_used,omitempty"`
}

// EnhancedHALBuilder provides advanced collection building capabilities
type EnhancedHALBuilder struct {
	*HALCollectionBuilder
	metadata         CollectionMeta
	facets          []Facet
	operations      []AllowedOp
	sortOptions     []SortOption
	enableMetadata  bool
	enableFacets    bool
	enableOps       bool
	queryStartTime  time.Time
}

// NewEnhancedHALBuilder creates a new enhanced HAL builder
func NewEnhancedHALBuilder(c *gin.Context) *EnhancedHALBuilder {
	baseBuilder := NewHALCollectionBuilder(c)
	
	return &EnhancedHALBuilder{
		HALCollectionBuilder: baseBuilder,
		metadata: CollectionMeta{
			Generated: time.Now(),
			Query:     make(map[string]interface{}),
			Filters:   make(map[string]interface{}),
		},
		queryStartTime:  time.Now(),
		enableMetadata: true,
		enableFacets:   false,
		enableOps:      false,
	}
}

// EnableMetadata enables collection metadata
func (b *EnhancedHALBuilder) EnableMetadata(enable bool) *EnhancedHALBuilder {
	b.enableMetadata = enable
	return b
}

// EnableFacets enables faceted search
func (b *EnhancedHALBuilder) EnableFacets(enable bool) *EnhancedHALBuilder {
	b.enableFacets = enable
	return b
}

// EnableOperations enables collection operations
func (b *EnhancedHALBuilder) EnableOperations(enable bool) *EnhancedHALBuilder {
	b.enableOps = enable
	return b
}

// WithPerformance adds performance metadata
func (b *EnhancedHALBuilder) WithPerformance(queryTime time.Duration, cacheHit bool, indexUsed []string) *EnhancedHALBuilder {
	if b.enableMetadata {
		b.metadata.Performance = PerformanceMeta{
			QueryTime: queryTime,
			TotalTime: time.Since(b.queryStartTime),
			CacheHit:  cacheHit,
			IndexUsed: indexUsed,
		}
	}
	return b
}

// WithSort adds sorting information
func (b *EnhancedHALBuilder) WithSort(field, direction string, options []SortOption) *EnhancedHALBuilder {
	if b.enableMetadata {
		b.metadata.Sort = SortInfo{
			Field:     field,
			Direction: direction,
			Available: options,
		}
		b.sortOptions = options
	}
	return b
}

// WithQueryInfo adds query information to metadata
func (b *EnhancedHALBuilder) WithQueryInfo(queryParams map[string]interface{}) *EnhancedHALBuilder {
	if b.enableMetadata {
		b.metadata.Query = queryParams
	}
	return b
}

// WithFilters adds filter information to metadata
func (b *EnhancedHALBuilder) WithFilters(filters map[string]interface{}) *EnhancedHALBuilder {
	if b.enableMetadata {
		b.metadata.Filters = filters
	}
	return b
}

// AddFacet adds a facet to the collection
func (b *EnhancedHALBuilder) AddFacet(facet Facet) *EnhancedHALBuilder {
	if b.enableFacets {
		// Generate URLs for facet values if not provided
		for i := range facet.Values {
			if facet.Values[i].URL == "" {
				facet.Values[i].URL = b.buildFacetURL(facet.Name, facet.Values[i].Value)
			}
		}
		b.facets = append(b.facets, facet)
	}
	return b
}

// AddOperation adds an allowed operation to the collection
func (b *EnhancedHALBuilder) AddOperation(op AllowedOp) *EnhancedHALBuilder {
	if b.enableOps {
		b.operations = append(b.operations, op)
	}
	return b
}

// buildFacetURL builds a URL for a facet value
func (b *EnhancedHALBuilder) buildFacetURL(facetName, facetValue string) string {
	u, _ := url.Parse(b.baseURL)
	q := u.Query()

	// Add current query parameters
	for key, value := range b.query {
		q.Set(key, value)
	}

	// Add pagination
	q.Set("page", "1") // Reset to first page when filtering
	q.Set("size", strconv.Itoa(b.size))

	// Add facet filter
	q.Set(facetName, facetValue)

	u.RawQuery = q.Encode()
	return u.String()
}

// BuildEnhanced creates an enhanced HAL collection
func (b *EnhancedHALBuilder) BuildEnhanced(items any) EnhancedHALCollection {
	collection := EnhancedHALCollection{
		Embedded: map[string]any{
			"items": items,
		},
		Links: b.BuildLinks(),
		Page:  b.BuildPageInfo(),
		Total: b.total,
	}

	if b.enableMetadata {
		collection.Metadata = b.metadata
	}

	if b.enableFacets && len(b.facets) > 0 {
		collection.Facets = b.facets
	}

	if b.enableOps && len(b.operations) > 0 {
		collection.Operations = b.operations
	}

	return collection
}

// SendEnhanced sends an enhanced HAL collection response
func (b *EnhancedHALBuilder) SendEnhanced(c *gin.Context, items any) {
	hm := NewHeaderManager(c)
	hm.SetHALContentType()

	// Add additional headers
	hm.SetPublicCache(300) // 5 minutes cache for collections
	hm.SetVary("Accept", "Authorization", "Accept-Language")

	collection := b.BuildEnhanced(items)

	// Add performance headers if available
	if b.enableMetadata && b.metadata.Performance.QueryTime > 0 {
		c.Header("X-Query-Time", b.metadata.Performance.QueryTime.String())
		c.Header("X-Total-Time", b.metadata.Performance.TotalTime.String())
		
		if b.metadata.Performance.CacheHit {
			c.Header("X-Cache", "HIT")
		} else {
			c.Header("X-Cache", "MISS")
		}
	}

	c.JSON(200, collection)
}

// Advanced pagination functions

// SendAdvancedPagination sends an advanced paginated response
func SendAdvancedPagination(c *gin.Context, items any, total int64, config PaginationConfig) {
	builder := NewEnhancedHALBuilder(c)
	builder.SetTotal(total)
	builder.EnableMetadata(config.EnableMetadata)
	builder.EnableFacets(config.EnableFacets)
	builder.EnableOperations(config.EnableOperations)

	// Add sort options if provided
	if len(config.SortOptions) > 0 {
		currentSort := c.Query("sort")
		currentOrder := c.DefaultQuery("order", "asc")
		builder.WithSort(currentSort, currentOrder, config.SortOptions)
	}

	// Add filters if provided
	if len(config.Filters) > 0 {
		builder.WithFilters(config.Filters)
	}

	// Add facets if provided
	if config.EnableFacets && len(config.Facets) > 0 {
		for _, facet := range config.Facets {
			builder.AddFacet(facet)
		}
	}

	// Add operations if provided
	if config.EnableOperations && len(config.Operations) > 0 {
		for _, op := range config.Operations {
			builder.AddOperation(op)
		}
	}

	// Add performance info if provided
	if config.PerformanceInfo != nil {
		builder.WithPerformance(
			config.PerformanceInfo.QueryTime,
			config.PerformanceInfo.CacheHit,
			config.PerformanceInfo.IndexUsed,
		)
	}

	builder.SendEnhanced(c, items)
}

// PaginationConfig configures advanced pagination
type PaginationConfig struct {
	EnableMetadata    bool
	EnableFacets      bool
	EnableOperations  bool
	SortOptions       []SortOption
	Filters           map[string]interface{}
	Facets            []Facet
	Operations        []AllowedOp
	PerformanceInfo   *PerformanceMeta
}

// DefaultPaginationConfig returns a default pagination configuration
func DefaultPaginationConfig() PaginationConfig {
	return PaginationConfig{
		EnableMetadata:   true,
		EnableFacets:     false,
		EnableOperations: false,
		Filters:          make(map[string]interface{}),
	}
}

// SendCachedPagination sends a cached paginated response
func SendCachedPagination(c *gin.Context, items any, total int64, cacheKey string, ttl time.Duration) {
	builder := NewEnhancedHALBuilder(c)
	builder.SetTotal(total)
	builder.EnableMetadata(true)
	builder.WithPerformance(0, true, nil) // Indicate cache hit

	// Add cache headers
	hm := NewHeaderManager(c)
	hm.SetPublicCache(int(ttl.Seconds()))
	hm.SetETag(cacheKey, false)

	builder.SendEnhanced(c, items)
}

// SendSearchPagination sends search results with advanced pagination
func SendSearchPagination(c *gin.Context, items any, total int64, searchMeta SearchMetadata) {
	builder := NewEnhancedHALBuilder(c)
	builder.SetTotal(total)
	builder.EnableMetadata(true)
	builder.EnableFacets(true)
	builder.WithQueryInfo(map[string]interface{}{
		"search_query":    searchMeta.Query,
		"search_time":     searchMeta.SearchTime,
		"highlighting":    searchMeta.Highlighting,
		"suggestions":     searchMeta.Suggestions,
	})

	// Add search facets if available
	if len(searchMeta.Facets) > 0 {
		for _, facet := range searchMeta.Facets {
			builder.AddFacet(facet)
		}
	}

	// Add search performance
	if searchMeta.PerformanceInfo != nil {
		builder.WithPerformance(
			searchMeta.PerformanceInfo.QueryTime,
			searchMeta.PerformanceInfo.CacheHit,
			searchMeta.PerformanceInfo.IndexUsed,
		)
	}

	// Enhance embedded content with search metadata
	collection := builder.BuildEnhanced(items)
	collection.Embedded["search_metadata"] = map[string]interface{}{
		"query":         searchMeta.Query,
		"total_matches": total,
		"search_time":   searchMeta.SearchTime,
		"highlighting":  searchMeta.Highlighting,
		"suggestions":   searchMeta.Suggestions,
	}

	hm := NewHeaderManager(c)
	hm.SetHALContentType()
	c.JSON(200, collection)
}

// SearchMetadata contains search-specific metadata
type SearchMetadata struct {
	Query           string
	SearchTime      time.Duration
	Highlighting    map[string]interface{}
	Suggestions     []string
	Facets          []Facet
	PerformanceInfo *PerformanceMeta
}

// SendResourceCollection sends a collection of resources with individual links
func SendResourceCollection[T any](c *gin.Context, items []T, total int64, linkFunc func(T) HALLinks) {
	builder := NewEnhancedHALBuilder(c)
	builder.SetTotal(total)
	builder.EnableMetadata(true)

	// Transform items to include links
	enhancedItems := make([]map[string]interface{}, len(items))
	for i, item := range items {
		// Convert item to map
		itemMap := make(map[string]interface{})
		
		// Use reflection or type assertion to populate itemMap
		// This is a simplified version - in practice, you'd use proper serialization
		if linkFunc != nil {
			links := linkFunc(item)
			itemMap["_links"] = links
		}
		
		enhancedItems[i] = itemMap
	}

	builder.SendEnhanced(c, enhancedItems)
}

// Specialized pagination functions

// SendAuditPagination sends audit log pagination with temporal facets
func SendAuditPagination(c *gin.Context, items any, total int64, timeRange TimeRange) {
	builder := NewEnhancedHALBuilder(c)
	builder.SetTotal(total)
	builder.EnableMetadata(true)
	builder.EnableFacets(true)

	// Add temporal facets
	timeFacet := Facet{
		Name:  "time_range",
		Label: "Time Range",
		Type:  "temporal",
		Values: []FacetValue{
			{Value: "last_hour", Label: "Last Hour", URL: builder.buildFacetURL("time_range", "last_hour")},
			{Value: "last_day", Label: "Last Day", URL: builder.buildFacetURL("time_range", "last_day")},
			{Value: "last_week", Label: "Last Week", URL: builder.buildFacetURL("time_range", "last_week")},
			{Value: "last_month", Label: "Last Month", URL: builder.buildFacetURL("time_range", "last_month")},
		},
	}
	
	if timeRange.Current != "" {
		timeFacet.Applied = []string{timeRange.Current}
	}

	builder.AddFacet(timeFacet)

	// Add metadata about time range
	builder.WithQueryInfo(map[string]interface{}{
		"time_range": map[string]interface{}{
			"start":   timeRange.Start,
			"end":     timeRange.End,
			"current": timeRange.Current,
		},
	})

	builder.SendEnhanced(c, items)
}

// TimeRange represents a time range for temporal queries
type TimeRange struct {
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
	Current string    `json:"current"`
}

// SendMetricsPagination sends metrics with aggregation metadata
func SendMetricsPagination(c *gin.Context, items any, total int64, metrics MetricsMetadata) {
	builder := NewEnhancedHALBuilder(c)
	builder.SetTotal(total)
	builder.EnableMetadata(true)
	builder.WithQueryInfo(map[string]interface{}{
		"aggregation": metrics.Aggregation,
		"interval":    metrics.Interval,
		"metrics":     metrics.AvailableMetrics,
	})

	// Add aggregation operations
	if metrics.EnableOperations {
		builder.EnableOperations(true)
		
		ops := []AllowedOp{
			{
				Name:   "aggregate",
				Method: "POST",
				URL:    builder.baseURL + "/aggregate",
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"metrics": map[string]interface{}{"type": "array"},
						"interval": map[string]interface{}{"type": "string"},
						"groupBy": map[string]interface{}{"type": "array"},
					},
				},
			},
		}
		
		for _, op := range ops {
			builder.AddOperation(op)
		}
	}

	builder.SendEnhanced(c, items)
}

// MetricsMetadata contains metrics-specific metadata
type MetricsMetadata struct {
	Aggregation      string
	Interval         string
	AvailableMetrics []string
	EnableOperations bool
}