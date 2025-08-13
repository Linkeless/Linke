package response

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// QueryProcessor handles standardized query parameter parsing
type QueryProcessor struct {
	c *gin.Context
}

// NewQueryProcessor creates a new query processor
func NewQueryProcessor(c *gin.Context) *QueryProcessor {
	return &QueryProcessor{c: c}
}

// PaginationQuery represents pagination parameters
type PaginationQuery struct {
	Page   int `json:"page"`   // 1-based page number
	Size   int `json:"size"`   // Items per page
	Offset int `json:"offset"` // Calculated offset for database queries
}

// SortQuery represents sorting parameters
type SortQuery struct {
	Field string `json:"field"` // Field to sort by
	Order string `json:"order"` // "asc" or "desc"
}

// FilterQuery represents filtering parameters
type FilterQuery struct {
	Search    string            `json:"search,omitempty"`     // Full-text search
	Filters   map[string]string `json:"filters,omitempty"`    // Field-specific filters
	DateRange *DateRangeFilter  `json:"date_range,omitempty"` // Date range filter
}

// DateRangeFilter represents date range filtering
type DateRangeFilter struct {
	Field string     `json:"field"`          // Date field name
	From  *time.Time `json:"from,omitempty"` // Start date
	To    *time.Time `json:"to,omitempty"`   // End date
}

// FieldSelection represents field selection for partial responses
type FieldSelection struct {
	Fields  []string `json:"fields,omitempty"`  // Fields to include
	Exclude []string `json:"exclude,omitempty"` // Fields to exclude
}

// QueryOptions contains all parsed query parameters
type QueryOptions struct {
	Pagination   PaginationQuery `json:"pagination"`
	Sort         *SortQuery      `json:"sort,omitempty"`
	Filter       FilterQuery     `json:"filter"`
	FieldSelect  *FieldSelection `json:"field_selection,omitempty"`
	IncludeTotal bool            `json:"include_total"`
	IncludeCount bool            `json:"include_count"`
}

// ParsePagination extracts and validates pagination parameters
func (q *QueryProcessor) ParsePagination() PaginationQuery {
	page, _ := strconv.Atoi(q.c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(q.c.DefaultQuery("size", "20"))

	// Validate and apply constraints
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100 // Maximum page size limit
	}

	offset := (page - 1) * size

	return PaginationQuery{
		Page:   page,
		Size:   size,
		Offset: offset,
	}
}

// ParseSort extracts and validates sorting parameters
func (q *QueryProcessor) ParseSort(allowedFields []string) *SortQuery {
	sortParam := q.c.Query("sort")
	if sortParam == "" {
		return nil
	}

	// Parse sort parameter: "field" or "field:desc" or "-field"
	field := sortParam
	order := "asc"

	// Handle "-field" format (descending)
	if strings.HasPrefix(sortParam, "-") {
		field = strings.TrimPrefix(sortParam, "-")
		order = "desc"
	} else if strings.Contains(sortParam, ":") {
		// Handle "field:desc" format
		parts := strings.Split(sortParam, ":")
		if len(parts) == 2 {
			field = parts[0]
			if parts[1] == "desc" || parts[1] == "asc" {
				order = parts[1]
			}
		}
	}

	// Override with explicit order parameter
	if orderParam := q.c.Query("order"); orderParam != "" {
		if orderParam == "asc" || orderParam == "desc" {
			order = orderParam
		}
	}

	// Validate field against allowed list
	if len(allowedFields) > 0 {
		allowed := false
		for _, allowedField := range allowedFields {
			if field == allowedField {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil // Field not allowed
		}
	}

	return &SortQuery{
		Field: field,
		Order: order,
	}
}

// ParseFilters extracts filtering parameters
func (q *QueryProcessor) ParseFilters(allowedFilters []string) FilterQuery {
	filter := FilterQuery{
		Filters: make(map[string]string),
	}

	// Search parameter
	filter.Search = q.c.Query("search")

	// Individual filters
	for _, filterName := range allowedFilters {
		if value := q.c.Query(filterName); value != "" {
			filter.Filters[filterName] = value
		}
	}

	// Date range filters
	if from := q.c.Query("from"); from != "" {
		if to := q.c.Query("to"); to != "" {
			dateRange := q.parseDateRange("created_at", from, to)
			if dateRange != nil {
				filter.DateRange = dateRange
			}
		}
	}

	return filter
}

// parseDateRange parses date range parameters
func (q *QueryProcessor) parseDateRange(field, from, to string) *DateRangeFilter {
	dateRange := &DateRangeFilter{Field: field}

	// Parse from date
	if fromTime, err := time.Parse("2006-01-02", from); err == nil {
		dateRange.From = &fromTime
	} else if fromTime, err := time.Parse(time.RFC3339, from); err == nil {
		dateRange.From = &fromTime
	}

	// Parse to date
	if toTime, err := time.Parse("2006-01-02", to); err == nil {
		// Set to end of day for date-only format
		endOfDay := toTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		dateRange.To = &endOfDay
	} else if toTime, err := time.Parse(time.RFC3339, to); err == nil {
		dateRange.To = &toTime
	}

	// Validate date range
	if dateRange.From != nil && dateRange.To != nil && dateRange.From.After(*dateRange.To) {
		return nil // Invalid range
	}

	return dateRange
}

// ParseFieldSelection extracts field selection parameters
func (q *QueryProcessor) ParseFieldSelection() *FieldSelection {
	fieldsParam := q.c.Query("fields")
	excludeParam := q.c.Query("exclude")

	if fieldsParam == "" && excludeParam == "" {
		return nil
	}

	selection := &FieldSelection{}

	// Parse include fields
	if fieldsParam != "" {
		fields := q.parseFieldList(fieldsParam)
		if len(fields) > 0 {
			selection.Fields = fields
		}
	}

	// Parse exclude fields
	if excludeParam != "" {
		exclude := q.parseFieldList(excludeParam)
		if len(exclude) > 0 {
			selection.Exclude = exclude
		}
	}

	return selection
}

// parseFieldList parses comma-separated field list with support for nested fields
func (q *QueryProcessor) parseFieldList(fields string) []string {
	// Remove whitespace and split by comma
	fieldList := strings.Split(strings.ReplaceAll(fields, " ", ""), ",")

	var result []string
	for _, field := range fieldList {
		if field != "" && q.isValidFieldName(field) {
			result = append(result, field)
		}
	}

	return result
}

// isValidFieldName validates field names (alphanumeric, underscore, dot for nested)
func (q *QueryProcessor) isValidFieldName(field string) bool {
	// Allow alphanumeric, underscore, and dot for nested fields
	matched, _ := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_.]*$", field)
	return matched
}

// ParseQueryOptions parses all query parameters into a structured format
func (q *QueryProcessor) ParseQueryOptions(config QueryParsingConfig) QueryOptions {
	options := QueryOptions{
		Pagination:   q.ParsePagination(),
		Filter:       q.ParseFilters(config.AllowedFilters),
		IncludeTotal: q.c.Query("include_total") != "false", // Default true
		IncludeCount: q.c.Query("include_count") == "true",  // Default false
	}

	// Parse sorting if allowed
	if len(config.AllowedSortFields) > 0 {
		options.Sort = q.ParseSort(config.AllowedSortFields)
	}

	// Parse field selection if enabled
	if config.AllowFieldSelection {
		options.FieldSelect = q.ParseFieldSelection()
	}

	return options
}

// QueryParsingConfig configures query parameter parsing
type QueryParsingConfig struct {
	AllowedSortFields   []string // Fields that can be used for sorting
	AllowedFilters      []string // Filters that are allowed
	AllowFieldSelection bool     // Whether field selection is enabled
	DefaultPageSize     int      // Default page size (if not 20)
	MaxPageSize         int      // Maximum page size (if not 100)
}

// StandardConfig returns a standard configuration for most APIs
func StandardConfig() QueryParsingConfig {
	return QueryParsingConfig{
		AllowedSortFields:   []string{"id", "name", "created_at", "updated_at"},
		AllowedFilters:      []string{"status", "type", "category"},
		AllowFieldSelection: true,
		DefaultPageSize:     20,
		MaxPageSize:         100,
	}
}

// BuildQueryString rebuilds a query string from options (useful for pagination links)
func (options *QueryOptions) BuildQueryString() string {
	params := url.Values{}

	// Add pagination
	params.Set("page", strconv.Itoa(options.Pagination.Page))
	params.Set("size", strconv.Itoa(options.Pagination.Size))

	// Add sorting
	if options.Sort != nil {
		sortValue := options.Sort.Field
		if options.Sort.Order == "desc" {
			sortValue = "-" + sortValue
		}
		params.Set("sort", sortValue)
	}

	// Add search
	if options.Filter.Search != "" {
		params.Set("search", options.Filter.Search)
	}

	// Add filters
	for key, value := range options.Filter.Filters {
		params.Set(key, value)
	}

	// Add date range
	if options.Filter.DateRange != nil {
		if options.Filter.DateRange.From != nil {
			params.Set("from", options.Filter.DateRange.From.Format("2006-01-02"))
		}
		if options.Filter.DateRange.To != nil {
			params.Set("to", options.Filter.DateRange.To.Format("2006-01-02"))
		}
	}

	// Add field selection
	if options.FieldSelect != nil {
		if len(options.FieldSelect.Fields) > 0 {
			params.Set("fields", strings.Join(options.FieldSelect.Fields, ","))
		}
		if len(options.FieldSelect.Exclude) > 0 {
			params.Set("exclude", strings.Join(options.FieldSelect.Exclude, ","))
		}
	}

	return params.Encode()
}

// Helper functions for common query patterns

// ParseListQuery parses query parameters for list endpoints
func ParseListQuery(c *gin.Context, config QueryParsingConfig) QueryOptions {
	processor := NewQueryProcessor(c)
	return processor.ParseQueryOptions(config)
}

// ParseSearchQuery parses query parameters specifically for search endpoints
func ParseSearchQuery(c *gin.Context, allowedFilters []string) QueryOptions {
	config := QueryParsingConfig{
		AllowedSortFields:   []string{"relevance", "created_at", "updated_at", "name"},
		AllowedFilters:      allowedFilters,
		AllowFieldSelection: true,
	}

	processor := NewQueryProcessor(c)
	options := processor.ParseQueryOptions(config)

	// Search endpoints require a search term
	if options.Filter.Search == "" {
		options.Filter.Search = c.Query("q") // Alternative search parameter
	}

	return options
}

// ValidateRequiredFilters validates that required filters are present
func ValidateRequiredFilters(options QueryOptions, required []string) []string {
	var missing []string

	for _, requiredFilter := range required {
		if _, exists := options.Filter.Filters[requiredFilter]; !exists {
			if requiredFilter == "search" && options.Filter.Search == "" {
				missing = append(missing, requiredFilter)
			} else if requiredFilter != "search" {
				missing = append(missing, requiredFilter)
			}
		}
	}

	return missing
}
