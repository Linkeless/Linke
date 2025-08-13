package response

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// HALPagination represents HAL-compliant pagination
type HALPagination struct {
	Page     int   `json:"page"`               // Current page (1-based)
	Size     int   `json:"size"`               // Page size
	Total    int64 `json:"total"`              // Total items
	Pages    int   `json:"pages"`              // Total pages
	First    bool  `json:"first"`              // Is first page
	Last     bool  `json:"last"`               // Is last page
	Previous *int  `json:"previous,omitempty"` // Previous page number
	Next     *int  `json:"next,omitempty"`     // Next page number
}

// HALCollectionBuilder helps build HAL collection responses
type HALCollectionBuilder struct {
	baseURL    string
	page       int
	size       int
	total      int64
	query      map[string]string
	sortFields []string
}

// NewHALCollectionBuilder creates a new HAL collection builder
func NewHALCollectionBuilder(c *gin.Context) *HALCollectionBuilder {
	builder := &HALCollectionBuilder{
		query: make(map[string]string),
	}

	// Extract pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100 // Max page size
	}

	builder.page = page
	builder.size = size

	// Build base URL from request
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	builder.baseURL = fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, c.Request.URL.Path)

	// Extract query parameters (excluding pagination)
	for key, values := range c.Request.URL.Query() {
		if key != "page" && key != "size" && len(values) > 0 {
			builder.query[key] = values[0]
		}
	}

	return builder
}

// SetTotal sets the total number of items
func (b *HALCollectionBuilder) SetTotal(total int64) *HALCollectionBuilder {
	b.total = total
	return b
}

// SetSort sets sorting fields for URLs
func (b *HALCollectionBuilder) SetSort(fields ...string) *HALCollectionBuilder {
	b.sortFields = fields
	return b
}

// AddQueryParam adds a query parameter
func (b *HALCollectionBuilder) AddQueryParam(key, value string) *HALCollectionBuilder {
	if value != "" {
		b.query[key] = value
	}
	return b
}

// buildURL builds a URL with pagination and query parameters
func (b *HALCollectionBuilder) buildURL(page int) string {
	u, _ := url.Parse(b.baseURL)
	q := u.Query()

	// Add pagination
	q.Set("page", strconv.Itoa(page))
	q.Set("size", strconv.Itoa(b.size))

	// Add other query parameters
	for key, value := range b.query {
		q.Set(key, value)
	}

	// Add sort fields
	if len(b.sortFields) > 0 {
		q.Set("sort", strings.Join(b.sortFields, ","))
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// BuildLinks creates HAL navigation links
func (b *HALCollectionBuilder) BuildLinks() HALLinks {
	totalPages := int((b.total + int64(b.size) - 1) / int64(b.size))
	if totalPages < 1 {
		totalPages = 1
	}

	links := HALLinks{
		Self: &HALLink{
			Href: b.buildURL(b.page),
		},
		First: &HALLink{
			Href: b.buildURL(1),
		},
		Last: &HALLink{
			Href: b.buildURL(totalPages),
		},
	}

	// Previous page link
	if b.page > 1 {
		links.Prev = &HALLink{
			Href: b.buildURL(b.page - 1),
		}
	}

	// Next page link
	if b.page < totalPages {
		links.Next = &HALLink{
			Href: b.buildURL(b.page + 1),
		}
	}

	return links
}

// BuildPageInfo creates pagination metadata
func (b *HALCollectionBuilder) BuildPageInfo() PageInfo {
	totalPages := int((b.total + int64(b.size) - 1) / int64(b.size))
	if totalPages < 1 {
		totalPages = 1
	}

	return PageInfo{
		Size:          b.size,
		TotalElements: b.total,
		TotalPages:    totalPages,
		Number:        b.page - 1, // HAL uses 0-based page numbers
	}
}

// BuildPagination creates HAL-style pagination info
func (b *HALCollectionBuilder) BuildPagination() HALPagination {
	totalPages := int((b.total + int64(b.size) - 1) / int64(b.size))
	if totalPages < 1 {
		totalPages = 1
	}

	pagination := HALPagination{
		Page:  b.page,
		Size:  b.size,
		Total: b.total,
		Pages: totalPages,
		First: b.page == 1,
		Last:  b.page == totalPages,
	}

	if b.page > 1 {
		prev := b.page - 1
		pagination.Previous = &prev
	}

	if b.page < totalPages {
		next := b.page + 1
		pagination.Next = &next
	}

	return pagination
}

// Build creates the complete HAL collection response
func (b *HALCollectionBuilder) Build(items any) HALCollection {
	return HALCollection{
		Embedded: map[string]any{
			"items": items,
		},
		Links: b.BuildLinks(),
		Page:  b.BuildPageInfo(),
		Total: b.total,
	}
}

// SendCollection sends a HAL collection response
func (b *HALCollectionBuilder) SendCollection(c *gin.Context, items any) {
	// Set HAL content type
	hm := NewHeaderManager(c)
	hm.SetHALContentType()

	collection := b.Build(items)
	c.JSON(200, collection)
}

// Helper functions for common pagination patterns

// SendPaginatedResponse sends a paginated response using HAL format
func SendPaginatedResponse(c *gin.Context, items any, total int64) {
	builder := NewHALCollectionBuilder(c).SetTotal(total)
	builder.SendCollection(c, items)
}

// SendPaginatedWithSort sends a paginated response with sorting
func SendPaginatedWithSort(c *gin.Context, items any, total int64, sortFields ...string) {
	builder := NewHALCollectionBuilder(c).SetTotal(total).SetSort(sortFields...)
	builder.SendCollection(c, items)
}

// SendFilteredCollection sends a filtered and paginated collection
func SendFilteredCollection(c *gin.Context, items any, total int64, filters map[string]string) {
	builder := NewHALCollectionBuilder(c).SetTotal(total)

	// Add filter parameters
	for key, value := range filters {
		builder.AddQueryParam(key, value)
	}

	builder.SendCollection(c, items)
}

// SendSearchResults sends search results with pagination and query info
func SendSearchResults(c *gin.Context, items any, total int64, searchQuery string) {
	builder := NewHALCollectionBuilder(c).
		SetTotal(total).
		AddQueryParam("search", searchQuery)

	// Set content type and send
	hm := NewHeaderManager(c)
	hm.SetHALContentType()

	collection := builder.Build(items)

	// Add search context to the collection
	if searchQuery != "" {
		collection.Embedded["search"] = map[string]any{
			"query":         searchQuery,
			"total_matches": total,
		}
	}

	c.JSON(200, collection)
}

// Pagination utilities

// ExtractPaginationParams extracts and validates pagination parameters
func ExtractPaginationParams(c *gin.Context) (page, size, offset int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ = strconv.Atoi(c.DefaultQuery("size", "20"))

	// Validate and set defaults
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100 // Maximum page size
	}

	offset = (page - 1) * size
	return page, size, offset
}

// ExtractSortParams extracts sorting parameters
func ExtractSortParams(c *gin.Context, allowedFields []string) (sortBy string, sortOrder string) {
	sortBy = c.Query("sort")
	sortOrder = strings.ToLower(c.DefaultQuery("order", "asc"))

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	// Validate sort field
	if sortBy != "" && len(allowedFields) > 0 {
		allowed := false
		for _, field := range allowedFields {
			if field == sortBy {
				allowed = true
				break
			}
		}
		if !allowed {
			sortBy = "" // Reset to default if not allowed
		}
	}

	return sortBy, sortOrder
}

// ExtractFilterParams extracts filter parameters
func ExtractFilterParams(c *gin.Context, allowedFilters []string) map[string]string {
	filters := make(map[string]string)

	for _, filter := range allowedFilters {
		if value := c.Query(filter); value != "" {
			filters[filter] = value
		}
	}

	return filters
}

// BuildSelfLink creates a self link for a resource
func BuildSelfLink(c *gin.Context, resourcePath ...string) string {
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}

	path := c.Request.URL.Path
	if len(resourcePath) > 0 {
		path = strings.Join(resourcePath, "/")
	}

	return fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, path)
}

// SendHALResource sends a single HAL resource with links
func SendHALResource(c *gin.Context, resource any, links ...HALLinks) {
	hm := NewHeaderManager(c)
	hm.SetHALContentType()

	if len(links) > 0 {
		// Add links to resource if it's a map
		if resourceMap, ok := resource.(map[string]any); ok {
			resourceMap["_links"] = links[0]
			c.JSON(200, resourceMap)
			return
		}
	}

	// Send resource directly if no links or not a map
	c.JSON(200, resource)
}
