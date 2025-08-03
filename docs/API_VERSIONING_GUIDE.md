# API 版本管理指南

本文档为在 Linke 项目中使用 API 版本管理系统提供全面指导。

## 概述

Linke 项目实现了一个灵活的 API 版本管理系统，支持多种版本控制策略、自动版本协商、优雅的弃用机制和向后兼容性功能。

### 核心特性

- **多种版本控制策略**: URL 路径、HTTP 头部、查询参数和基于内容类型的版本控制
- **版本协商**: 自动检测和验证 API 版本
- **弃用支持**: 适当的弃用警告和下线头部信息
- **迁移助手**: 版本间的自动数据迁移
- **向后兼容**: 版本回退和兼容性检查
- **配置驱动**: 易于配置和自定义

## 版本控制策略

### 1. URL 路径版本控制（默认）

最常见且推荐的方法。

```
GET /api/v1/users
GET /api/v2/users
```

**配置:**
```bash
API_VERSION_STRATEGY=url_path
API_URL_PREFIX=/api
```

### 2. Header 版本控制

在 HTTP 头部中指定版本。

```bash
GET /api/users
X-API-Version: v2
```

**配置:**
```bash
API_VERSION_STRATEGY=header
API_VERSION_HEADER=X-API-Version
```

### 3. 查询参数版本控制

将版本指定为查询参数。

```
GET /api/users?version=v2
```

**配置:**
```bash
API_VERSION_STRATEGY=query
API_VERSION_QUERY_PARAM=version
```

### 4. Content-Type 版本控制

在 Accept 头部中指定版本。

```bash
GET /api/users
Accept: application/vnd.api+json;version=2
```

**配置:**
```bash
API_VERSION_STRATEGY=content_type
```

## 配置

### 环境变量

| 变量名 | 默认值 | 描述 |
|----------|---------|-------------|
| `API_VERSION_STRATEGY` | `url_path` | 版本控制策略 |
| `API_DEFAULT_VERSION` | `1.0.0` | 未指定版本时的默认版本 |
| `API_MIN_VERSION` | `1.0.0` | 支持的最小版本 |
| `API_MAX_VERSION` | `2.0.0` | 支持的最大版本 |
| `API_VERSION_HEADER` | `X-API-Version` | Header 策略的头部名称 |
| `API_VERSION_QUERY_PARAM` | `version` | 查询策略的查询参数 |
| `API_URL_PREFIX` | `/api` | 路径策略的 URL 前缀 |
| `API_ENABLE_DEPRECATION_HEADERS` | `true` | 启用弃用警告 |
| `API_ENABLE_AUTO_MIGRATION` | `false` | 启用自动版本迁移 |
| `API_SUNSET_V1_DATE` | `2025-12-31T23:59:59Z` | v1 版本下线日期 |

### .env 配置示例

```bash
# API 版本管理配置
API_VERSION_STRATEGY=url_path
API_DEFAULT_VERSION=2.0.0
API_MIN_VERSION=1.0.0
API_MAX_VERSION=2.0.0
API_URL_PREFIX=/api
API_ENABLE_DEPRECATION_HEADERS=true
API_ENABLE_AUTO_MIGRATION=false
API_SUNSET_V1_DATE=2025-12-31T23:59:59Z
```

## 版本特定处理器

### 基础版本注册

```go
func RegisterUserRoutes(versionRouter *versioning.VersionRouter) {
    v1 := versioning.NewVersion(1, 0, 0)
    v2 := versioning.NewVersion(2, 0, 0)
    
    // 为不同版本注册不同的处理器
    versionRouter.GET("/users/:id", v1, getUserV1)
    versionRouter.GET("/users/:id", v2, getUserV2)
}

func getUserV1(c *gin.Context) {
    // V1 实现，使用简单的响应格式
    user := map[string]any{
        "id":    c.Param("id"),
        "name":  "John Doe",
        "email": "john@example.com",
    }
    response.Success(c, user)
}

func getUserV2(c *gin.Context) {
    // V2 实现，使用增强的响应格式
    user := map[string]any{
        "user_id":    c.Param("id"),  // 字段名已更改
        "full_name":  "John Doe",     // 字段名已更改
        "email":      "john@example.com",
        "created_at": "2024-01-01T00:00:00Z", // 新字段
        "profile": map[string]any{    // 嵌套数据
            "avatar_url": "https://example.com/avatar.jpg",
            "bio":        "Software Engineer",
        },
    }
    response.Success(c, user)
}
```

### 版本感知处理器

```go
func getServerStatus(c *gin.Context) {
    versionCtx, exists := versioning.GetVersionFromContext(c)
    if !exists {
        response.ErrorJSON(c, http.StatusInternalServerError, response.ErrorResponse{
            Error:   "version_context_missing",
            Message: "版本上下文未找到",
        })
        return
    }
    
    version := versionCtx.ResolvedVersion
    
    // 根据版本调整响应
    if version.Major == 1 {
        // V1 格式：简单状态
        response.Success(c, map[string]any{
            "status":    "healthy",
            "timestamp": time.Now().Format(time.RFC3339),
        })
    } else {
        // V2+ 格式：增强状态
        response.Success(c, map[string]any{
            "server_status": "healthy",
            "last_checked":  time.Now().Format(time.RFC3339),
            "system_info": map[string]any{
                "version":     "2.1.0",
                "environment": "production",
            },
            "metrics": map[string]any{
                "cpu_usage":    "15.5%",
                "memory_usage": "342MB",
            },
        })
    }
}
```

## 版本迁移

### 自动迁移

启用自动迁移以实现向后兼容：

```bash
API_ENABLE_AUTO_MIGRATION=true
```

### 自定义迁移

```go
func registerMigrations(registry *versioning.MigrationRegistry) {
    v1 := versioning.NewVersion(1, 0, 0)
    v2 := versioning.NewVersion(2, 0, 0)
    
    // 注册从 v1 到 v2 的用户数据迁移
    migration := versioning.VersionMigration{
        FromVersion: v1,
        ToVersion:   v2,
        Mappings: []versioning.FieldMapping{
            {FromField: "name", ToField: "full_name", Required: true},
            {FromField: "id", ToField: "user_id", Required: true},
            {FromField: "", ToField: "created_at", DefaultValue: "2024-01-01T00:00:00Z"},
        },
    }
    
    registry.RegisterMigration(migration)
}
```

### 自定义迁移函数

```go
func customUserMigration(from, to versioning.Version, data any) (any, error) {
    userMap, ok := data.(map[string]any)
    if !ok {
        return nil, fmt.Errorf("用户迁移的数据类型无效")
    }
    
    // 自定义迁移逻辑
    migratedUser := map[string]any{
        "user_id":    userMap["id"],
        "full_name":  userMap["name"],
        "email":      userMap["email"],
        "created_at": time.Now().Format(time.RFC3339),
        "profile": map[string]any{
            "avatar_url": "",
            "bio":        "",
        },
    }
    
    return migratedUser, nil
}

// 注册自定义迁移
migration := versioning.VersionMigration{
    FromVersion: v1,
    ToVersion:   v2,
    CustomFunc:  customUserMigration,
}
```

## 响应头

版本管理系统会自动为响应添加多个头部信息：

### 标准头部

- `X-API-Version`: 使用的已解析 API 版本
- `X-API-Version-Strategy`: 使用的版本控制策略
- `X-API-Version-Requested`: 原始请求的版本（如果不同）
- `X-API-Version-Latest`: 最新可用版本
- `X-API-Versions-Supported`: 所有支持版本的列表

### 弃用头部

针对已弃用的版本：

- `Warning`: 弃用警告消息
- `Sunset`: 版本下线日期（RFC 1123 格式）
- `X-API-Sunset-Days`: 距离下线的天数
- `Link`: 指向后续版本的链接

### 迁移头部

当发生版本迁移时：

- `X-API-Version-Migrated-From`: 原始版本
- `X-API-Version-Migrated-To`: 目标版本
- `X-API-Response-Migrated-From`: 响应数据原始版本
- `X-API-Response-Migrated-To`: 响应数据目标版本

## API 端点

### 版本信息

```
GET /api/version
```

返回全面的版本信息：

```json
{
  "message": "API version information",
  "data": {
    "current_version": "2.0.0",
    "latest_version": "2.0.0",
    "min_version": "1.0.0",
    "max_version": "2.0.0",
    "supported_versions": [
      {
        "version": "1.0.0",
        "status": "deprecated",
        "sunset_date": "2025-12-31T23:59:59Z",
        "description": "Initial API version (deprecated)",
        "released": "2023-01-01T00:00:00Z"
      },
      {
        "version": "2.0.0",
        "status": "active",
        "description": "Enhanced API with improved features",
        "released": "2024-01-01T00:00:00Z"
      }
    ],
    "strategy": "url_path",
    "deprecation_policy": {
      "enable_deprecation_headers": true,
      "enable_auto_migration": false
    }
  }
}
```

### 带版本信息的健康检查

```
GET /api/health
```

返回带版本信息的健康状态：

```json
{
  "status": "healthy",
  "service": "linke-api",
  "version": "2.0.0",
  "version_info": {
    "version": "2.0.0",
    "status": "active",
    "description": "Enhanced API with improved features"
  }
}
```

## 错误响应

### 不支持的版本

```
HTTP/1.1 400 Bad Request
Content-Type: application/json
X-API-Recommendation: Use version 2.0.0

{
  "error": "unsupported_api_version",
  "message": "API version 3.0.0 is not supported",
  "details": {
    "requested_version": "3.0.0",
    "supported_versions": ["1.0.0", "2.0.0"],
    "min_version": "1.0.0",
    "max_version": "2.0.0",
    "latest_version": "2.0.0"
  }
}
```

### 已下线版本

```
HTTP/1.1 410 Gone
Content-Type: application/json
X-API-Migration: Migrate to version 2.0.0

{
  "error": "api_version_sunset",
  "message": "API version 1.0.0 has been sunset and is no longer available",
  "details": {
    "requested_version": "1.0.0",
    "sunset_date": "2025-01-01T00:00:00Z",
    "latest_version": "2.0.0",
    "migration_guide": "Please migrate to version 2.0.0"
  }
}
```

### 版本未实现

```
HTTP/1.1 501 Not Implemented
Content-Type: application/json

{
  "error": "version_not_implemented",
  "message": "Version 2.0.0 is not implemented for this endpoint",
  "details": {
    "requested_version": "2.0.0",
    "available_versions": ["1.0.0"],
    "endpoint": "GET /api/analytics",
    "migration_required": true
  }
}
```

## 最佳实践

### 1. 版本规划

- 仔细规划版本变更
- 记录破坏性变更
- 提供迁移指南
- 设置合理的下线时间表

### 2. 向后兼容

- 在主版本内保持向后兼容
- 使用语义化版本控制（major.minor.patch）
- 仅在主版本中引入破坏性变更
- 尽可能提供优雅降级

### 3. 弃用策略

- 提前带告弃用通知
- 一致地使用弃用头部
- 提供清晰的迁移路径
- 监控已弃用版本的使用情况

### 4. 测试

- 测试所有支持的版本
- 验证版本协商
- 测试迁移路径
- 监控性能影响

### 5. 文档

- 记录所有版本变更
- 维护 API 更新日志
- 提供迁移示例
- 保持弃用计划更新

## 迁移检查清单

引入新 API 版本时：

### 发布前

- [ ] 规划版本变更和破坏性变更
- [ ] 更新版本配置
- [ ] 实现新版本处理器
- [ ] 创建迁移映射
- [ ] 更新新版本测试
- [ ] 记录 API 变更
- [ ] 规划弃用时间表

### 发布

- [ ] 与现有版本并行部署新版本
- [ ] 更新 API 文档
- [ ] 向客户端宣布新版本
- [ ] 监控版本使用情况
- [ ] 提供迁移支持

### 发布后

- [ ] 监控错误率和性能
- [ ] 跟踪版本采用情况
- [ ] 支持客户端迁移
- [ ] 规划旧版本弃用
- [ ] 更新下线日期

### 弃用

- [ ] 带时间表宣布弃用
- [ ] 添加弃用头部
- [ ] 通知正在使用已弃用版本的客户端
- [ ] 提供迁移协助
- [ ] 监控使用量下降

### 下线

- [ ] 设置下线日期
- [ ] 发送最终迁移通知
- [ ] 为已下线版本返回 410 Gone
- [ ] 移除已弃用版本代码
- [ ] 更新文档

## 故障排除

### 常见问题

1. **版本未被检测到**
   - 检查版本控制策略配置
   - 验证 URL 模式或头部格式
   - 检查中间件顺序

2. **迁移失败**
   - 检查迁移映射
   - 检查数据类型和结构
   - 验证自定义迁移函数

3. **性能问题**
   - 监控版本协商开销
   - 优化迁移函数
   - 考虑缓存策略

4. **客户端兼容性**
   - 检查弃用时间表
   - 提供清晰的迁移指南
   - 监控客户端错误率

### 调试模式

启用版本管理的调试日志：

```bash
LOG_LEVEL=debug
```

这将记录版本协商详情、迁移尝试和遇到的任何问题。

## 安全考虑

- 验证版本输入以防止注入攻击
- 对版本相关端点进行限率
- 监控异常的版本协商模式
- 正确保护已下线版本端点
- 审计版本访问模式

## 性能考虑

- 版本协商添加的开销最小
- 迁移可能影响响应时间
- 考虑缓存已迁移的响应
- 监控版本特定的性能指标
- 优化常用的迁移路径