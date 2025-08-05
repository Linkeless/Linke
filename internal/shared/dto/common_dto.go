package dto

import (
	"fmt"
	"strings"
	"time"
)

// BaseDTO represents common fields that all DTOs might have
type BaseDTO struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PaginationDTO represents pagination information
type PaginationDTO struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	Pages    int `json:"pages"`
}

// ErrorDTO represents error information
type ErrorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// =============================================================================
// Generic Request Structures
// =============================================================================

// PaginationRequest represents common pagination parameters for list requests
type PaginationRequest struct {
	Limit  int `form:"limit,omitempty" binding:"omitempty,min=1,max=100" example:"10" json:"limit,omitempty"`
	Offset int `form:"offset,omitempty" binding:"omitempty,min=0" example:"0" json:"offset,omitempty"`
}

// GetPage returns the page number (1-based) from offset and limit
func (p PaginationRequest) GetPage() int {
	if p.Limit == 0 {
		return 1
	}
	return (p.Offset / p.Limit) + 1
}

// GetLimit returns the limit with a default value if not set
func (p PaginationRequest) GetLimit() int {
	if p.Limit == 0 {
		return 10 // default limit
	}
	return p.Limit
}

// StatusFilterRequest represents common status filtering parameters
type StatusFilterRequest struct {
	Status string `form:"status,omitempty" example:"active" json:"status,omitempty"`
}

// UserFilterRequest represents common user filtering parameters
type UserFilterRequest struct {
	UserID uint `form:"user_id,omitempty" example:"1" json:"user_id,omitempty"`
}

// TimeRangeRequest represents common time range filtering parameters
type TimeRangeRequest struct {
	DateFrom string `form:"date_from,omitempty" example:"2024-01-01" json:"date_from,omitempty"`
	DateTo   string `form:"date_to,omitempty" example:"2024-12-31" json:"date_to,omitempty"`
}

// SearchRequest represents common search parameters
type SearchRequest struct {
	Query string `form:"query,omitempty" example:"search term" json:"query,omitempty"`
}

// SortRequest represents common sorting parameters
type SortRequest struct {
	SortBy    string `form:"sort_by,omitempty" example:"created_at" json:"sort_by,omitempty"`
	SortOrder string `form:"sort_order,omitempty" example:"desc" json:"sort_order,omitempty"`
}

// BaseListRequest represents a comprehensive list request structure
// that combines all common filtering, pagination, and search parameters
type BaseListRequest struct {
	PaginationRequest
	StatusFilterRequest
	UserFilterRequest
	TimeRangeRequest  
	SearchRequest
	SortRequest
}

// IDRequest represents a request that contains a single ID parameter
type IDRequest struct {
	ID uint `uri:"id" binding:"required,min=1" example:"1" json:"id"`
}

// IDsRequest represents a request that contains multiple ID parameters
type IDsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1" example:"[1,2,3]"`
}

// UUIDRequest represents a request that contains a UUID parameter
type UUIDRequest struct {
	UUID string `uri:"uuid" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000" json:"uuid"`
}

// =============================================================================
// Generic Response Structures
// =============================================================================

// PaginatedResponse represents a paginated response with generic data type
type PaginatedResponse[T any] struct {
	Data       []T           `json:"data"`
	Pagination PaginationDTO `json:"pagination"`
}

// APIResponse represents a generic API response wrapper
type APIResponse[T any] struct {
	Success bool      `json:"success"`
	Data    T         `json:"data,omitempty"`
	Error   *ErrorDTO `json:"error,omitempty"`
	Message string    `json:"message,omitempty"`
}

// BatchOperationResult represents the result of a batch operation
type BatchOperationResult[ID comparable] struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	FailedIDs    []ID     `json:"failed_ids,omitempty"`
	Errors       []string `json:"errors,omitempty"`
	Message      string   `json:"message,omitempty"`
}

// StatsResponse represents common statistics response structure
type StatsResponse struct {
	TotalCount   int64                  `json:"total_count"`
	StatusCounts map[string]int64       `json:"status_counts,omitempty"`
	DateCounts   map[string]int64       `json:"date_counts,omitempty"`
	CustomStats  map[string]interface{} `json:"custom_stats,omitempty"`
	Period       string                 `json:"period,omitempty"` // e.g., "daily", "monthly", "yearly"
}

// CountResponse represents a simple count response
type CountResponse struct {
	Count int64 `json:"count"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// StatusResponse represents a status response with optional details
type StatusResponse struct {
	Status  string                 `json:"status"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// =============================================================================
// Specialized Request Structures
// =============================================================================

// BulkUpdateRequest represents a request for bulk updates
type BulkUpdateRequest[T any] struct {
	Items []T `json:"items" binding:"required,min=1"`
}

// BulkCreateRequest represents a request for bulk creation
type BulkCreateRequest[T any] struct {
	Items []T `json:"items" binding:"required,min=1"`
}

// BulkDeleteRequest represents a request for bulk deletion
type BulkDeleteRequest struct {
	IDsRequest
	Force bool `json:"force,omitempty"` // Force delete even if there are dependencies
}

// UpdateStatusRequest represents a request to update status
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required" example:"active"`
	Reason string `json:"reason,omitempty" example:"User requested activation"`
}

// ToggleRequest represents a request to toggle a boolean field
type ToggleRequest struct {
	Enabled bool   `json:"enabled" example:"true"`
	Reason  string `json:"reason,omitempty" example:"Toggle reason"`
}

// =============================================================================
// Validation Helpers
// =============================================================================

// ValidateTimeRange validates that DateFrom is before DateTo if both are provided
func (tr TimeRangeRequest) ValidateTimeRange() error {
	if tr.DateFrom == "" || tr.DateTo == "" {
		return nil
	}
	
	from, err := time.Parse("2006-01-02", tr.DateFrom)
	if err != nil {
		return err
	}
	
	to, err := time.Parse("2006-01-02", tr.DateTo)
	if err != nil {
		return err
	}
	
	if from.After(to) {
		return fmt.Errorf("date_from must be before date_to")
	}
	
	return nil
}

// GetSortOrder returns normalized sort order (asc/desc)
func (sr SortRequest) GetSortOrder() string {
	switch strings.ToLower(sr.SortOrder) {
	case "desc", "descending":
		return "desc"
	default:
		return "asc"
	}
}

// =============================================================================
// Response Builders
// =============================================================================

// NewAPIResponse creates a new successful API response
func NewAPIResponse[T any](data T, message ...string) APIResponse[T] {
	resp := APIResponse[T]{
		Success: true,
		Data:    data,
	}
	if len(message) > 0 {
		resp.Message = message[0]
	}
	return resp
}

// NewAPIError creates a new error API response
func NewAPIError[T any](err ErrorDTO) APIResponse[T] {
	return APIResponse[T]{
		Success: false,
		Error:   &err,
	}
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse[T any](data []T, pagination PaginationDTO) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Data:       data,
		Pagination: pagination,
	}
}

// NewBatchOperationResult creates a new batch operation result
func NewBatchOperationResult[ID comparable](successCount int, failedIDs []ID, errors []string) BatchOperationResult[ID] {
	return BatchOperationResult[ID]{
		SuccessCount: successCount,
		FailedCount:  len(failedIDs),
		FailedIDs:    failedIDs,
		Errors:       errors,
	}
}
