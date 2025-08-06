# 通用 DTO 使用示例

本文档演示如何使用新添加的通用 DTO 结构来消除项目中的代码重复。

## 请求结构示例

### 1. 基本列表请求

```go
// 不要在每个领域中定义单独的分页、过滤和搜索结构，
// 使用统一的 BaseListRequest

// 旧方式（在每个领域中重复）：
type ListUsersRequest struct {
    Limit    int    `form:"limit,omitempty"`
    Offset   int    `form:"offset,omitempty"`
    Status   string `form:"status,omitempty"`
    Query    string `form:"query,omitempty"`
    DateFrom string `form:"date_from,omitempty"`
    DateTo   string `form:"date_to,omitempty"`
}

// 新方式（使用通用 DTO）：
type ListUsersRequest struct {
    dto.BaseListRequest
    // 如果需要，添加领域特定字段
    Role string `form:"role,omitempty" example:"admin"`
}
```

### 2. 个别过滤器请求

```go
// 当您只需要特定过滤器时，使用个别过滤器组件

type ListActiveUsersRequest struct {
    dto.PaginationRequest
    dto.StatusFilterRequest
}

type SearchUsersRequest struct {
    dto.SearchRequest
    dto.UserFilterRequest
}
```

### 3. 批量操作

```go
// 使用泛型类型的批量操作
type BulkUpdateUsersRequest struct {
    dto.BulkUpdateRequest[UserUpdateData]
}

type BulkDeleteUsersRequest struct {
    dto.BulkDeleteRequest
}
```

## 响应结构示例

### 1. 分页响应

```go
// 旧方式：
type ListUsersResponse struct {
    Users      []UserDTO     `json:"users"`
    Pagination PaginationDTO `json:"pagination"`
}

// 新方式：
type ListUsersResponse = dto.PaginatedResponse[UserDTO]

// 在处理器中使用：
func (h *UserHandler) ListUsers(c *gin.Context) {
    users := []UserDTO{...}
    pagination := dto.PaginationDTO{...}
    
    response := dto.NewPaginatedResponse(users, pagination)
    c.JSON(http.StatusOK, response)
}
```

### 2. API 响应封装器

```go
// 统一的 API 响应格式
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

### 3. 批量操作结果

```go
func (h *UserHandler) BulkDeleteUsers(c *gin.Context) {
    var req dto.BulkDeleteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // 处理错误
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

### 4. 统计响应

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

## 验证示例

### 1. 时间范围验证

```go
func (h *Handler) ListWithTimeRange(c *gin.Context) {
    var req struct {
        dto.TimeRangeRequest
        dto.PaginationRequest
    }
    
    if err := c.ShouldBindQuery(&req); err != nil {
        // 处理绑定错误
        return
    }
    
    // 验证时间范围
    if err := req.TimeRangeRequest.ValidateTimeRange(); err != nil {
        c.JSON(http.StatusBadRequest, dto.NewAPIError[interface{}](dto.ErrorDTO{
            Code:    "INVALID_TIME_RANGE",
            Message: err.Error(),
        }))
        return
    }
    
    // 继续业务逻辑...
}
```

### 2. 使用辅助方法

```go
func (h *Handler) ListUsers(c *gin.Context) {
    var req dto.BaseListRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        // 处理错误
        return
    }
    
    // 使用辅助方法
    limit := req.GetLimit()           // 如果未设置则返回 10
    page := req.GetPage()             // 从 offset/limit 计算页码
    sortOrder := req.GetSortOrder()   // 标准化为 "asc" 或 "desc"
    
    // 在服务调用中使用
    users, total, err := h.userService.ListUsers(limit, req.Offset, req.Status, sortOrder)
    // ...
}
```

## 迁移指南

### 转换现有结构

1. **识别重复模式** 在您现有的请求/响应结构中
2. **用通用 DTO 替换** 适用的部分
3. **添加领域特定字段** 通过嵌入通用 DTO
4. **更新处理器** 使用新的辅助方法和构造函数

### 迁移示例

```go
// 之前：
type ListPaymentsRequest struct {
    Limit    int    `form:"limit,omitempty"`
    Offset   int    `form:"offset,omitempty"`
    UserID   uint   `form:"user_id,omitempty"`
    Status   string `form:"status,omitempty"`
    DateFrom string `form:"date_from,omitempty"`
    DateTo   string `form:"date_to,omitempty"`
}

// 之后：
type ListPaymentsRequest struct {
    dto.BaseListRequest
    // 添加支付特定字段
    PaymentMethod string  `form:"payment_method,omitempty" example:"credit_card"`
    Amount        float64 `form:"amount,omitempty" example:"99.99"`
}
```

这种方法提供：
- **类型安全** 通过 Go 泛型
- **一致性** 跨所有 API 端点
- **减少代码重复**
- **更好的可维护性**
- **标准化验证**
- **改进的文档** 通过统一的 Swagger 标签