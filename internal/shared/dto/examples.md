# Common DTO Usage Examples

This document demonstrates how to use the newly added common DTO structures to eliminate code duplication across the project.

## Request Structure Examples

### 1. Basic List Request

```go
// Instead of defining separate pagination, filtering, and search structures
// in each domain, use the unified BaseListRequest

// Old way (repeated in each domain):
type ListUsersRequest struct {
    Limit    int    `form:"limit,omitempty"`
    Offset   int    `form:"offset,omitempty"`
    Status   string `form:"status,omitempty"`
    Query    string `form:"query,omitempty"`
    DateFrom string `form:"date_from,omitempty"`
    DateTo   string `form:"date_to,omitempty"`
}

// New way (using common DTO):
type ListUsersRequest struct {
    dto.BaseListRequest
    // Add domain-specific fields if needed
    Role string `form:"role,omitempty" example:"admin"`
}
```

### 2. Individual Filter Requests

```go
// Use individual filter components when you only need specific filters

type ListActiveUsersRequest struct {
    dto.PaginationRequest
    dto.StatusFilterRequest
}

type SearchUsersRequest struct {
    dto.SearchRequest
    dto.UserFilterRequest
}
```

### 3. Batch Operations

```go
// Bulk operations with generic types
type BulkUpdateUsersRequest struct {
    dto.BulkUpdateRequest[UserUpdateData]
}

type BulkDeleteUsersRequest struct {
    dto.BulkDeleteRequest
}
```

## Response Structure Examples

### 1. Paginated Responses

```go
// Old way:
type ListUsersResponse struct {
    Users      []UserDTO     `json:"users"`
    Pagination PaginationDTO `json:"pagination"`
}

// New way:
type ListUsersResponse = dto.PaginatedResponse[UserDTO]

// Usage in handler:
func (h *UserHandler) ListUsers(c *gin.Context) {
    users := []UserDTO{...}
    pagination := dto.PaginationDTO{...}
    
    response := dto.NewPaginatedResponse(users, pagination)
    c.JSON(http.StatusOK, response)
}
```

### 2. API Response Wrapper

```go
// Unified API response format
func (h *UserHandler) GetUser(c *gin.Context) {
    user, err := h.userService.GetUser(userID)
    if err != nil {
        errorResp := dto.NewAPIError[UserDTO](dto.ErrorDTO{
            Code:    "USER_NOT_FOUND",
            Message: "User not found",
        })
        c.JSON(http.StatusNotFound, errorResp)
        return
    }
    
    response := dto.NewAPIResponse(user, "User retrieved successfully")
    c.JSON(http.StatusOK, response)
}
```

### 3. Batch Operation Results

```go
func (h *UserHandler) BulkDeleteUsers(c *gin.Context) {
    var req dto.BulkDeleteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // handle error
        return
    }
    
    successCount, failedIDs, errors := h.userService.BulkDelete(req.IDs, req.Force)
    
    result := dto.NewBatchOperationResult(
        successCount,
        failedIDs,
        errors,
    )
    
    c.JSON(http.StatusOK, result)
}
```

### 4. Statistics Responses

```go
func (h *UserHandler) GetUserStats(c *gin.Context) {
    stats := dto.StatsResponse{
        TotalCount: 1500,
        StatusCounts: map[string]int64{
            "active":   1200,
            "inactive": 250,
            "suspended": 50,
        },
        DateCounts: map[string]int64{
            "2024-01": 100,
            "2024-02": 150,
            "2024-03": 200,
        },
        CustomStats: map[string]interface{}{
            "average_age": 28.5,
            "premium_users": 300,
        },
        Period: "monthly",
    }
    
    c.JSON(http.StatusOK, stats)
}
```

## Validation Examples

### 1. Time Range Validation

```go
func (h *Handler) ListWithTimeRange(c *gin.Context) {
    var req struct {
        dto.TimeRangeRequest
        dto.PaginationRequest
    }
    
    if err := c.ShouldBindQuery(&req); err != nil {
        // handle binding error
        return
    }
    
    // Validate time range
    if err := req.TimeRangeRequest.ValidateTimeRange(); err != nil {
        c.JSON(http.StatusBadRequest, dto.NewAPIError[interface{}](dto.ErrorDTO{
            Code:    "INVALID_TIME_RANGE",
            Message: err.Error(),
        }))
        return
    }
    
    // Continue with business logic...
}
```

### 2. Using Helper Methods

```go
func (h *Handler) ListUsers(c *gin.Context) {
    var req dto.BaseListRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        // handle error
        return
    }
    
    // Use helper methods
    limit := req.GetLimit()           // Returns 10 if not set
    page := req.GetPage()             // Calculates page from offset/limit
    sortOrder := req.GetSortOrder()   // Normalizes to "asc" or "desc"
    
    // Use in service call
    users, total, err := h.userService.ListUsers(limit, req.Offset, req.Status, sortOrder)
    // ...
}
```

## Migration Guide

### Converting Existing Structures

1. **Identify repeated patterns** in your existing request/response structures
2. **Replace with common DTOs** where applicable
3. **Add domain-specific fields** by embedding common DTOs
4. **Update handlers** to use new helper methods and constructors

### Example Migration

```go
// Before:
type ListPaymentsRequest struct {
    Limit    int    `form:"limit,omitempty"`
    Offset   int    `form:"offset,omitempty"`
    UserID   uint   `form:"user_id,omitempty"`
    Status   string `form:"status,omitempty"`
    DateFrom string `form:"date_from,omitempty"`
    DateTo   string `form:"date_to,omitempty"`
}

// After:
type ListPaymentsRequest struct {
    dto.BaseListRequest
    // Add payment-specific fields
    PaymentMethod string  `form:"payment_method,omitempty" example:"credit_card"`
    Amount        float64 `form:"amount,omitempty" example:"99.99"`
}
```

This approach provides:
- **Type safety** through Go generics
- **Consistency** across all API endpoints  
- **Reduced code duplication**
- **Better maintainability**
- **Standardized validation**
- **Improved documentation** through unified Swagger tags