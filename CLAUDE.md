# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述
Linke 是基于 Go 语言构建的现代化服务管理平台，采用 VSA (垂直切片架构) + 清洁架构模式。核心特性包括事件驱动架构、多层缓存系统、API版本管理、高性能支付处理、Telegram Bot 集成等。

**核心业务领域:**
- 用户管理 (`domains/user`) - 包含多级缓存的用户服务
- 身份认证 (`domains/auth`) - JWT黑名单、OAuth2、登录安全
- 订阅计费 (`domains/subscription`) - 订阅生命周期、暂停/恢复、使用量追踪
- 支付处理 (`domains/payment`) - 高性能多网关支付、内存优化、防重放攻击
- 发票管理 (`domains/invoice`) - PDF生成、缓存下载、安全验证
- 服务器管理 (`domains/server`) - 服务器群组、shadowsocks配置
- 工单系统 (`domains/ticket`) - 完整的客户支持工单系统，支持多消息、附件、内部注释
- Telegram Bot (`shared/telegram`) - 增强型 Telegram Bot，支持丰富的菜单系统和工单通知

**业务流程:**
```
用户注册/认证 → 选择订阅套餐 → 创建订单 → 生成发票 → 支付处理 → 激活服务 → 使用量追踪
                                                                      ↓
                                                               工单支持 ← Telegram Bot
```

## 核心命令

**开发环境:**
- `make dev` - 开发模式运行 (自动生成 Swagger 文档)
- `make safe-dev` - 带安全预检的开发模式 (推荐生产环境)
- `make build` - 构建可执行文件到 `bin/server`
- `make install` - 安装依赖工具 (swag, migrate等)

**测试:**
- `make test` - 运行所有测试，详细输出
- `go test ./internal/domains/user` - 测试特定领域模块
- `go test -run TestFunctionName` - 运行特定测试函数
- `go test -race ./...` - 竞态条件检测

**数据库迁移 (集成式):**
- `make migrate-up` - 运行所有待执行迁移
- `make migrate-down` - 回滚一个迁移
- `make migrate-status` - 检查当前迁移状态
- `make migrate-list` - 列出所有已应用的迁移
- `make migrate-create NAME=name` - 创建新迁移文件
- `make migrate-fix-dirty VERSION=N` - 修复脏迁移状态
- `make migrate-goto VERSION=N` - 迁移到指定版本
- `make migrate-steps STEPS=2` - 向前运行2步迁移
- `make migrate-steps STEPS=-1` - 回滚1步迁移
- `make migrate-reset` - 重置数据库（删除所有表并重新运行迁移）

**开发工具:**
- `make swagger` - 手动生成 API 文档 (自动访问 `/swagger/index.html`)
- `make security-check` - 运行安全配置验证
- `go run tools/generate-jwt-key/main.go` - 生成安全的 JWT 密钥

**Swagger 文档配置:**
- BasePath 设置为 `/api/v1` (在 `cmd/server/main.go` 中配置)
- **重要**: Handler 中的 `@Router` 注释应使用相对路径，不要包含 `/api/v1` 前缀
- 示例: `@Router /user/bindings [get]` 而不是 `@Router /api/v1/user/bindings [get]`
- 最终生成路径: `basePath` + `@Router路径` = `/api/v1/user/bindings`

## RESTful API 重构 (2024年重大更新)

**系统已完成全面的RESTful API重构，所有API现在严格遵循RESTful设计原则：**

**响应格式变更:**
- **移除包装结构**: 不再使用 `{code, message, data}` 包装格式
- **直接返回资源**: API直接返回资源数据，如 `{"id": 1, "name": "user"}` 
- **标准HTTP状态码**: 通过HTTP状态码传达操作结果，无需自定义code字段

**错误响应 (RFC 9457 Problem JSON):**
```json
{
  "type": "/problems/not-found",
  "title": "Not Found", 
  "status": 404,
  "detail": "The user with id 123 was not found",
  "instance": "/api/v1/users/123"
}
```

**分页响应 (HAL格式):**
```json
{
  "_embedded": {
    "items": [{"id": 1, "name": "user1"}, {"id": 2, "name": "user2"}]
  },
  "_links": {
    "self": {"href": "/api/v1/users?page=1&size=20"},
    "first": {"href": "/api/v1/users?page=1&size=20"},
    "next": {"href": "/api/v1/users?page=2&size=20"},
    "last": {"href": "/api/v1/users?page=5&size=20"}
  },
  "page": {
    "size": 20,
    "totalElements": 100,
    "totalPages": 5,
    "number": 0
  },
  "total": 100
}
```

**新响应函数:**
- `response.OK(c, data)` - 200响应，直接返回资源数据
- `response.Created(c, data)` - 201响应，返回创建的资源
- `response.NoContent(c)` - 204响应，无内容
- `response.BadRequest(c, detail)` - 400 Problem JSON响应
- `response.NotFound(c, detail)` - 404 Problem JSON响应
- `response.SendPaginatedResponse(c, items, total)` - HAL分页响应

**标准HTTP头部支持:**
- `ETag` 和条件请求 (If-Match, If-None-Match)
- `Last-Modified` 和 `If-Modified-Since`
- `Cache-Control`, `Vary`, `Content-Location`
- `Idempotency-Key` 用于防重复操作
- `API-Version`, `Deprecation`, `Sunset` 用于版本管理

**查询参数标准化:**
- `page` - 页码 (1-based)
- `size` - 页面大小 (默认20，最大100)
- `sort` - 排序字段，支持 `field` 或 `-field` (降序)
- `search` - 全文搜索
- `fields` - 字段选择，如 `fields=id,name,email`
- `from`, `to` - 日期范围过滤

## 架构结构

**VSA + 清洁架构设计:**
```
internal/
├── application/          # 应用层 (工作流编排和启动逻辑)
│   ├── bootstrap/       # Fx 依赖注入和应用启动
│   ├── server/          # HTTP 服务器配置
│   ├── handlers/        # 应用级处理器
│   ├── services/        # 应用服务
│   └── workflows/       # 跨领域工作流
├── domains/             # 业务领域 (垂直切片)
│   ├── auth/           # JWT管理、OAuth2、登录安全
│   ├── user/           # 用户生命周期、多级缓存
│   ├── subscription/   # 订阅管理、暂停/恢复、使用量追踪
│   ├── payment/        # 高性能支付网关、内存优化、幂等性
│   ├── invoice/        # PDF生成、下载安全、缓存策略
│   ├── server/         # 服务器群组、shadowsocks配置
│   ├── ticket/         # 工单系统、客户支持
│   └── [其他]/          # coupon, referral
└── shared/             # 共享基础设施
    ├── cache/          # 多级缓存系统 (内存+Redis)
    ├── events/         # 事件驱动架构、发布订阅
    ├── versioning/     # API版本管理中间件
    ├── database/       # 数据库连接和迁移CLI
    ├── router/         # HTTP路由配置
    ├── config/         # 配置管理
    ├── constants/      # 系统常量定义
    ├── entities/       # 基础实体类型
    ├── middleware/     # HTTP中间件
    ├── queue/          # 任务队列系统
    ├── telegram/       # Telegram Bot 集成
    └── stubs/          # 临时存根实现
```

**领域内部结构 (清洁架构):**
```
domains/[领域]/
├── constants/          # 领域常量定义 (状态、类型等)
├── dto/               # 数据传输对象和转换函数
├── entities/          # 业务实体 + 验证逻辑
├── usecases/
│   ├── interfaces/    # 服务接口契约
│   └── implementations/
│       ├── [service].go        # 基础服务实现
│       ├── cached_[service].go # 缓存装饰器
│       └── [service]_test.go   # 单元测试
├── adapters/
│   ├── handlers/      # HTTP路由处理器
│   └── repositories/  # 数据访问层 (GORM)
└── module.go          # 依赖注入配置 (uber/fx)
```

## 关键配置

**必需环境变量:**
- `JWT_SECRET` - **关键**: 必须 32+ 字符，使用 `openssl rand -hex 32`
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`

**OAuth2 提供商:**
- Google: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`
- GitHub: `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_REDIRECT_URL`
- Telegram: `TELEGRAM_BOT_TOKEN`, `TELEGRAM_REDIRECT_URL`

**Telegram Bot 配置:**
- `TELEGRAM_BOT_TOKEN` - Bot Token (从 @BotFather 获取)
- `TELEGRAM_ADMIN_CHAT_IDS` - 管理员 Chat ID 列表（逗号分隔）

**安全要求:**
- 应用程序拒绝使用弱 JWT 配置启动
- 部署前运行 `make security-check`
- 开发环境使用 `make safe-dev`

## 核心架构模式

**Uber/Fx 依赖注入架构:**
- **应用启动**: `internal/application/bootstrap/app.go` 包含完整的 Fx 应用配置
- **模块化**: 每个领域都有独立的 `module.go` 文件定义 Fx 模块
- **启动顺序**: 配置 → 数据库 → 缓存 → 事件系统 → 业务领域 → HTTP路由
- **生命周期管理**: Fx hooks 管理组件启动和关闭

**事件驱动架构 (`shared/events`):**
- **事件总线**: 内存总线、Redis分布式总线、度量装饰器
- **事件存储**: 基于GORM的持久化事件存储，支持回放和查询
- **重要事件类型**: `subscription.created`, `payment.completed`, `invoice.generated`, `ticket.created`, `ticket.replied`, `ticket.resolved`
- **可靠性**: 至少一次投递保证、去重处理、熔断器模式
- **监控**: 完整的事件发布/订阅指标收集
- **Telegram 集成**: 工单事件自动触发 Telegram 通知

**多级缓存系统 (`shared/cache`):**
- **内存缓存**: 高性能本地缓存，适用于热点数据
- **Redis缓存**: 分布式缓存，支持集群环境
- **缓存键管理**: 统一的键生成策略，支持模式匹配删除
- **缓存装饰器**: 用户服务、支付服务、配置服务的缓存封装
- **TTL策略**: 短期(1分钟)、中期(15分钟)、长期(1小时)缓存

**API版本管理 (`shared/versioning`):**
- **版本策略**: 头部版本、路径版本、查询参数版本
- **版本中间件**: 自动版本协商和兼容性检查
- **弃用管理**: 版本日落警告、迁移指南、自动升级

**高性能支付处理 (`domains/payment`):**
- **内存优化**: 结构体字段对齐减少内存占用、对象池减少GC压力
- **并发优化**: 原子操作处理通知计数、HTTP连接池复用
- **幂等性**: 防重放攻击、基于哈希的去重检查
- **安全加固**: 增强的错误处理、详细日志记录、参数验证

**集成式数据库迁移:**
- **命令行集成**: 迁移命令直接集成到主应用中，通过命令行参数调用
- **迁移CLI**: `internal/shared/database/migration_cli.go` 处理所有迁移逻辑
- **支持命令**: up, down, reset, status, list, force, goto, steps, fix-dirty
- **环境变量**: 使用相同的数据库配置，无需额外配置

**Go 代码规范 (Uber指南):**
- **main.go 最小化**: 仅 69 行，只负责命令行解析和应用启动
- 使用 `shared/entities/BaseEntity` 减少结构体重复
- 使用 `shared/constants` 统一错误消息和状态值
- 使用 `shared/handlers` 通用函数处理ID解析、JSON绑定
- 错误包装: `fmt.Errorf("context: %w", err)`
- 依赖注入: 每个领域独立的 `fx.Module` 配置

## 关键开发要点

**启动应用程序:**
- **bootstrap 模式**: `internal/application/bootstrap.NewApplication()` 创建 Fx 应用
- **依赖注入**: 所有组件通过 Fx 自动注入，包括数据库、缓存、事件系统等
- **路由配置**: HTTP 路由在 `internal/shared/router/router.go` 中统一配置
- **迁移集成**: 支持通过命令行参数直接运行数据库迁移

**缓存失效策略:**
- 用户数据更新时自动失效相关缓存 (ID、邮箱、用户名)
- 支付记录缓存仅保存非敏感字段，敏感数据不缓存
- 使用缓存标签 (`CacheTagUser`, `CacheTagPayment`) 进行批量失效

**测试和调试:**
- 表驱动测试: 使用 `t.Run` 创建子测试，目标覆盖率 >80%
- 集成测试需要数据库连接，确保测试环境数据库可用
- 使用 `go test -race` 检测竞态条件

**安全实现:**
- JWT令牌黑名单机制防止令牌重放
- 支付通知使用幂等键防重放攻击
- 文件下载包含安全验证和访问控制
- 所有密码使用 bcrypt 加密存储

**API路由结构:**
- `/health` - 系统健康检查
- `/swagger/index.html` - API文档 (开发环境)
- `/api/v1/auth/*` - 身份认证相关
- `/api/v1/users/*` - 用户管理
- `/api/v1/user/bindings/*` - 第三方账号绑定 (Google, GitHub, Telegram)
- `/api/v1/subscriptions/*` - 订阅服务
- `/api/v1/payments/*` - 支付处理
- `/api/v1/invoices/*` - 发票管理
- `/api/v1/tickets/*` - 工单管理
- `/api/v1/admin/*` - 管理员功能

**Telegram Bot 命令:**
- `/start` - 开始使用 Bot
- `/menu` - 显示主菜单
- `/subscription` - 查看订阅信息
- `/tickets` - 管理工单
- `/admin` - 管理面板（仅管理员）
- `/help` - 使用帮助

**常见问题解决:**
- **脏迁移**: `make migrate-fix-dirty VERSION=X`
- **JWT配置错误**: 检查 `JWT_SECRET` 长度 ≥32 字符
- **缓存连接失败**: 检查Redis连接配置和网络
- **支付处理异常**: 检查网关配置、连接池状态和防重放设置
- **事件处理异常**: 检查事件总线配置和订阅者注册
- **Fx 依赖注入错误**: 检查模块间的接口依赖和循环引用
- **Telegram Bot 启动失败**: 检查 `TELEGRAM_BOT_TOKEN` 是否正确
- **工单通知失败**: 检查用户是否已绑定 Telegram ID

## 重要开发注意事项

**添加新功能时:**
1. **领域优先**: 在对应的 `domains/` 目录下创建实体、用例和适配器
2. **架构分离**: 
   - `constants/` - 领域特定的常量 (状态、类型等)
   - `dto/` - 数据传输对象和转换函数
   - `entities/` - 纯业务实体，不包含常量或DTO
   - `usecases/interfaces/` - 服务契约，使用dto包类型
3. **依赖注入**: 在领域的 `module.go` 中注册 Fx 提供者
4. **路由注册**: 在 `shared/router/router.go` 中添加 HTTP 路由
5. **缓存考虑**: 对于频繁访问的数据，创建缓存装饰器
6. **事件集成**: 考虑是否需要发布领域事件
7. **Telegram 通知**: 对于重要操作考虑添加 Telegram 通知
8. **RESTful响应**: 使用新的响应函数，遵循RESTful标准

**RESTful Handler 开发指导:**
- **GET**: 使用 `response.OK(c, resource)` 返回资源数据
- **POST**: 使用 `response.Created(c, resource)` 返回创建的资源
- **PUT/PATCH**: 使用 `response.OK(c, updatedResource)` 返回更新后的资源
- **DELETE**: 使用 `response.NoContent(c)` 表示删除成功
- **列表接口**: 使用 `response.SendPaginatedResponse(c, items, total)` 
- **搜索接口**: 使用 `response.SendSearchResults(c, items, total, query)`
- **错误处理**: 使用相应的Problem JSON函数 (`BadRequest`, `NotFound`, `Conflict`等)
- **Swagger注释**: 直接引用资源类型，错误响应使用 `ProblemJSONResponse`

**架构重构准则 (重要):**
- 所有领域必须遵循统一的包结构: `constants/`, `dto/`, `entities/`, `usecases/`, `adapters/`
- 实体(entities)应该是纯业务对象，不包含常量定义或DTO
- 常量应该集中在 `constants/` 包中，按功能分组
- DTO和转换函数应该独立在 `dto/` 包中
- 服务接口应该使用 `dto` 包的类型，而不是在接口文件中重复定义
- 已重构领域: payment, coupon, referral, ticket, invoice - 作为标准参考

**修改现有代码时:**
- 遵循现有的 VSA + Clean Architecture 模式
- 保持 `main.go` 的简洁性，避免在其中添加业务逻辑
- 使用事件驱动模式处理跨领域交互
- 优先使用现有的共享基础设施组件
- Telegram Bot 相关修改集中在 `shared/telegram` 目录
- 工单系统修改需同步更新 Telegram 通知逻辑
- **支付领域优化**: 使用对象池、原子操作和优化的HTTP客户端，参考payment领域的性能优化实践

## Telegram Bot 开发指南

**Bot 架构:**
- `shared/telegram/bot_enhanced.go` - 主要 Bot 实现，包含菜单系统和命令处理
- `shared/telegram/ticket_event_handler.go` - 工单事件处理和通知
- `shared/telegram/ticket_notification.go` - 通知数据结构定义
- `shared/telegram/module.go` - Fx 依赖注入配置

**核心功能:**
- **菜单系统**: 使用内联键盘实现多级菜单导航
- **工单集成**: 支持创建、查看、回复工单，支持多消息缓冲
- **用户绑定**: 通过 Telegram ID 与系统用户关联
- **管理面板**: 管理员专属功能，包括批量操作和系统监控
- **通知系统**: 自动发送工单状态更新通知

**开发注意事项:**
- 所有用户操作前需验证 Telegram ID 绑定状态
- 使用 `ticketReplyBuffer` 管理多消息回复
- 管理员权限通过 `isUserAdmin` 方法验证
- 使用 MarkdownV2 格式时注意特殊字符转义
- API 限流：避免频繁调用 setMyName 等配置 API

## 支付领域性能优化 (2025年更新)

**支付领域已完成基于Go最佳实践的全面性能优化，显著提升系统性能和稳定性：**

### 内存优化
- **结构体字段对齐**: PaymentRecord结构体重新排列字段顺序，减少内存填充
- **对象池模式**: 为频繁使用的DTO（PaymentRecordResponse、CreatePaymentOrderRequest、NotifyData）实现sync.Pool
- **原子操作**: NotifyCount字段使用int32类型配合atomic操作，避免锁竞争

### 并发性能优化
- **HTTP连接池**: 实现单例模式的优化HTTP客户端，支持连接复用和合理的超时设置
- **原子计数器**: 通知计数使用原子操作，线程安全且性能更优
- **并发安全**: 支付通知处理中的关键操作使用原子增量

### 代码质量提升
- **增强错误处理**: 详细的参数验证、上下文传播和错误信息
- **结构化日志**: 添加关键操作的调试信息和性能监控点
- **代码规范**: 使用gofmt格式化，保持一致的代码风格

### 架构简化
- **移除重试机制**: 按业务需求移除支付重试相关代码，简化系统复杂度
- **保留核心功能**: 保持防重放攻击、幂等性检查等关键安全特性
- **清洁架构**: 维护VSA+清洁架构的设计原则

### 性能提升效果
- **内存使用**: 通过字段对齐和对象池显著减少内存分配
- **GC压力**: 对象池复用减少垃圾回收频率
- **并发性能**: 原子操作和连接池提升高并发场景表现
- **错误定位**: 增强的日志记录提高问题诊断效率

**使用建议:**
- 在高频支付场景中，使用DTO对象池函数（GetPaymentRecordResponse/PutPaymentRecordResponse）
- 监控原子操作的通知计数性能表现
- 利用增强的错误日志进行问题诊断和性能调优