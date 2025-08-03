# API 版本管理系统

为 Linke 项目提供的全面 API 版本管理系统，支持多种版本控制策略、自动版本协商、优雅弃用和向后兼容性。

## 快速开始

### 基本用法

```go
// 导入版本管理包
import "linke/internal/shared/versioning"

// 版本管理系统通过主配置自动配置
// 并通过依赖注入启用

// 在您的处理器中，访问版本信息：
func getUserHandler(c *gin.Context) {
    versionCtx, exists := versioning.GetVersionFromContext(c)
    if exists {
        version := versionCtx.ResolvedVersion
        // 处理版本特定逻辑
    }
    // ... 处理器的其余部分
}
```

### 版本特定路由

```go
// 为不同版本注册不同的处理器
func RegisterUserRoutes(versionRouter *versioning.VersionRouter) {
    v1 := versioning.NewVersion(1, 0, 0)
    v2 := versioning.NewVersion(2, 0, 0)
    
    versionRouter.GET("/users/:id", v1, getUserV1)
    versionRouter.GET("/users/:id", v2, getUserV2)
}
```

## 配置

设置环境变量：

```bash
API_VERSION_STRATEGY=url_path        # url_path, header, query, content_type
API_DEFAULT_VERSION=2.0.0            # 未指定时的默认版本
API_MIN_VERSION=1.0.0                # 支持的最小版本
API_MAX_VERSION=2.0.0                # 支持的最大版本
API_ENABLE_DEPRECATION_HEADERS=true  # 启用弃用警告
API_ENABLE_AUTO_MIGRATION=false      # 启用自动版本迁移
```

## 功能特性

### ✅ 多种版本控制策略

- **URL 路径**: `/api/v1/users`（默认）
- **HTTP 头部**: `X-API-Version: v2`
- **查询参数**: `/api/users?version=v2`
- **Content-Type**: `Accept: application/vnd.api+json;version=2`

### ✅ 版本协商

- 自动版本检测和验证
- 未指定时回退到默认版本
- 为不支持的版本提供适当的错误响应

### ✅ 弃用支持

- 通过 HTTP 头部的弃用警告
- 下线日期管理
- 渐进式迁移支持

### ✅ 数据迁移

- 版本间的自动字段映射
- 自定义迁移函数
- 响应格式转换

### ✅ 监控和头部

- 响应头部中的版本信息
- 弃用警告
- 迁移跟踪头部

## 客户端使用示例

### URL 路径版本控制

```bash
# 版本 1
curl https://api.linke.com/api/v1/users/123

# 版本 2
curl https://api.linke.com/api/v2/users/123
```

### HTTP 头部版本控制

```bash
curl -H "X-API-Version: v2" https://api.linke.com/api/users/123
```

### 获取版本信息

```bash
curl https://api.linke.com/api/version
```

## 响应头部信息

系统会自动添加版本相关的头部：

- `X-API-Version`: 使用的已解析版本
- `X-API-Version-Latest`: 最新可用版本
- `Warning`: 弃用警告（针对已弃用的版本）
- `Sunset`: 下线日期（针对已弃用的版本）

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                    版本管理系统                        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ 提取器     │  │ 中间件     │  │ 版本路由器        │  │
│  │             │  │             │  │                     │  │
│  │ • URL 路径  │  │ • 协商     │  │ • 路由分发        │  │
│  │ • 头部      │──▶│ • 验证     │──▶│ • 处理器选择      │  │
│  │ • 查询      │  │ • 上下文   │  │ • 回退逻辑        │  │
│  │ • 内容类型  │  │ • 头部     │  │                     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│                                                             │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │               迁移系统                          │  │
│  │                                                         │  │
│  │ • 字段映射          • 自定义函数                │  │
│  │ • 类型转换          • 响应迁移              │  │
│  │ • 默认值            • 向后兼容          │  │
│  └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 文件结构

```
internal/shared/versioning/
├── types.go          # 核心类型和版本处理
├── extractors.go     # 版本提取策略
├── middleware.go     # 版本协商中间件
├── router.go         # 版本感知路由
├── migration.go      # 数据迁移工具
├── examples.go       # 示例实现
├── module.go         # 依赖注入模块
├── *_test.go         # 单元测试
└── README.md         # 本文件
```

## 测试

运行版本管理测试：

```bash
go test ./internal/shared/versioning/...
```

测试覆盖：
- 版本解析和比较
- 所有提取策略
- 中间件版本协商
- 路由分发逻辑
- 迁移功能

## 文档

- **[API 版本管理指南](../../../docs/API_VERSIONING_GUIDE.md)** - 全面使用指南
- **[API 迁移指南](../../../docs/API_MIGRATION_GUIDE.md)** - 客户端迁移说明

## 示例

查看 `examples.go` 获取完整的实现示例，包括：
- 版本特定处理器
- 带版本适配的共享处理器
- 向后兼容模式
- 迁移演示

## 集成

版本管理系统通过以下方式自动集成到主应用程序中：

1. **配置**: 通过 `internal/shared/config`
2. **依赖注入**: 通过 `module.go` 中的 `fx.Module`
3. **中间件**: 应用于所有 API 路由
4. **路由**: 为版本特定处理器增强的路由器

## 贡献

添加新功能时：

1. 更新 `types.go` 中的相关类型
2. 为新功能添加测试
3. 必要时更新示例
4. 更新文档

## 性能

版本管理系统添加的开销最小：
- 版本提取：每请求 ~1-2μs
- 中间件处理：每请求 ~5-10μs
- 数据迁移：可变（取决于数据大小和映射）

使用内置的日志和指标监控性能。