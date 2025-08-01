# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**项目概述:**
Linke 是一个基于 Go 语言构建的综合服务管理平台，采用 VSA (Vertical Slice Architecture) + Clean Architecture 架构模式。系统提供订阅计费、用户管理、服务器管理等功能，支持 shadowsocks 服务器管理、多网关支付、推荐系统和工单客服系统。

**Go 开发规范:**
- 遵循 Go 官方代码规范和惯例
- 使用依赖注入模式进行组件解耦
- 接口优先设计，面向接口编程
- 统一的错误处理和日志记录
- 完整的单元测试和集成测试覆盖

## 开发命令

**构建和运行:**
- `make build` - 构建应用程序到 `bin/server`
- `make run` - 直接运行应用程序
- `make dev` - 开发模式运行 (先生成 Swagger 文档)
- `make safe-run` - 带安全预检查运行 (推荐)
- `make safe-dev` - 开发模式带安全检查运行 (推荐)
- `go run cmd/server/main.go app:serve` - 启动服务器
- `go run cmd/server/main.go -help` - 显示可用命令

**安全检查:**
- `make security-check` - 运行安全预检查验证
- `go run tools/generate-jwt-key/main.go` - 生成安全的 JWT 密钥

**测试和质量:**
- `make test` - 运行所有测试，详细输出 (`go test -v ./...`)
- `go test ./... -v -coverprofile cover.txt -coverpkg=./...` - 生成测试覆盖率报告
- `go tool cover -html=cover.txt -o index.html` - 生成 HTML 覆盖率报告
- `make lint-setup` - 配置代码检查和预提交钩子 (需要 Python3)
- 目标测试覆盖率: >80%

**文档生成:**
- `make swagger` - 在 `docs/` 目录生成 Swagger API 文档
- 服务器运行时可通过 `/swagger/*any` 访问 Swagger UI

**依赖管理:**
- `make install` - 安装依赖和必需工具 (swag)
- `make clean` - 清理构建产物和文档
- `go mod tidy` - 清理和同步模块依赖
- `go list -m -u all` - 检查所有模块更新
- `go get -u` - 更新所有依赖到最新版本
- `go get example.com/module@latest` - 更新特定模块

**数据库迁移 (集成式):**
- `make migrate-help` - 显示详细的迁移帮助和用法
- `make migrate-up` - 运行所有待执行的迁移
- `make migrate-down` - 回滚一个迁移
- `make migrate-status` - 显示当前迁移版本
- `make migrate-list` - 列出所有已应用的迁移
- `make migrate-reset` - 删除所有表并重新运行迁移 (危险操作!)
- `make migrate-create NAME=name` - 创建新的迁移文件
- `make migrate-force VERSION=N` - 强制设置迁移版本
- `make migrate-goto VERSION=N` - 迁移到指定版本
- `make migrate-steps STEPS=N` - 运行 N 步迁移 (正数向上，负数向下)
- `make migrate-fix-dirty VERSION=N` - 修复脏迁移状态 (使用错误消息中的版本号)
- `go run cmd/server/main.go -migrate` - 仅运行迁移后退出 (不启动服务器)

**迁移问题排查:**
- 如果看到 "Dirty database version X" 错误: `make migrate-fix-dirty VERSION=X`
- 或使用: `go run cmd/server/main.go -migrate-command=fix-dirty -migrate-version=X`

## 架构概述 (VSA + Clean Architecture)

**整体架构设计:**
本项目采用 VSA (Vertical Slice Architecture) + Clean Architecture 的混合架构模式，将业务按功能垂直切分成独立的领域模块，每个模块内部遵循 Clean Architecture 的四层结构。

**架构层次:**
```
├── cmd/server/                    # 应用程序入口点
├── internal/
│   ├── application/               # 应用层 - 跨领域业务协调
│   │   ├── services/              # 应用服务 - 领域模块协调器
│   │   ├── workflows/             # 业务工作流 - 复杂业务流程
│   │   └── handlers/              # 应用处理器 - 跨领域 HTTP 接口
│   ├── domains/                   # 领域层 - 业务功能垂直切片
│   │   ├── user/                  # 用户领域模块
│   │   ├── auth/                  # 认证领域模块
│   │   ├── subscription/          # 订阅领域模块
│   │   ├── payment/               # 支付领域模块
│   │   ├── coupon/                # 优惠券领域模块
│   │   ├── invoice/               # 发票领域模块
│   │   ├── ticket/                # 工单领域模块
│   │   ├── server/                # 服务器领域模块
│   │   └── referral/              # 推荐领域模块
│   └── shared/                    # 共享基础设施
│       ├── config/                # 配置管理
│       ├── database/              # 数据库连接
│       ├── logger/                # 日志系统
│       ├── middleware/            # 通用中间件
│       ├── response/              # 响应格式化
│       ├── events/                # 事件总线
│       └── queue/                 # 队列系统
```

**领域模块内部结构 (Clean Architecture):**
每个领域模块按 Clean Architecture 四层结构组织：

```
domains/[domain]/
├── entities/                      # 实体层 - 业务实体和核心业务规则
│   ├── [domain].go               # 核心业务实体
│   ├── value_objects.go          # 值对象
│   └── dto.go                    # 数据传输对象
├── usecases/                     # 用例层 - 业务逻辑和应用服务
│   ├── interfaces/               # 接口定义
│   │   ├── [domain]_repository.go
│   │   └── [domain]_service.go
│   └── implementations/          # 接口实现
│       ├── [domain]_service_impl.go
│       └── mock_[domain]_service.go
├── adapters/                     # 适配器层 - 外部接口适配
│   └── repositories/             # 数据访问实现
│       └── [domain]_repository.go
└── module.go                     # 模块组装和依赖注入
```

**关键架构原则:**

1. **领域驱动设计 (DDD)**: 按业务领域垂直切分，每个领域具有清晰的边界和职责
2. **依赖倒置**: 内层不依赖外层，通过接口进行解耦
3. **单一职责**: 每个模块、服务、实体都有明确的单一职责
4. **开闭原则**: 对扩展开放，对修改关闭
5. **接口隔离**: 使用小而专一的接口，避免臃肿的接口

**Go 语言特定原则:**
- **接口优先设计**: 定义小而专一的接口，便于测试和扩展
- **构造函数注入**: 使用 `NewXxx()` 构造函数进行依赖注入
- **错误处理**: 明确的错误返回和包装，使用 `pkg/errors` 增强错误信息
- **并发安全**: 合理使用 goroutine 和 channel，避免竞态条件
- **模块化**: 使用 `fx.Module` 进行依赖注入管理

## 核心业务领域模块

**User 领域模块** (`internal/domains/user/`):
- **职责**: 用户生命周期管理、用户信息维护、用户状态管理
- **核心实体**: User, UserProfile, UserPreferences
- **主要功能**: 用户注册、资料更新、状态变更、多提供商账号关联
- **依赖关系**: 被 Auth 领域依赖，依赖共享基础设施

**Auth 领域模块** (`internal/domains/auth/`):
- **职责**: 身份认证、授权管理、会话管理
- **核心实体**: AuthToken, Session, OAuthProvider, LoginAttempt
- **主要功能**: 多提供商 OAuth2、JWT 令牌管理、登录/注销、密码重置
- **依赖关系**: 依赖 User 领域获取用户信息

**Subscription 领域模块** (`internal/domains/subscription/`):
- **职责**: 订阅计划管理、用户订阅生命周期、计费周期管理
- **核心实体**: SubscriptionPlan, UserSubscription, SubscriptionOrder
- **主要功能**: 计划创建、订阅购买、续费、取消、试用期管理
- **依赖关系**: 与 Payment 领域协作处理付费订阅

**Payment 领域模块** (`internal/domains/payment/`):
- **职责**: 支付处理、多网关集成、支付记录管理
- **核心实体**: Payment, PaymentMethod, PaymentGateway, PaymentConfig
- **主要功能**: 支付创建、状态跟踪、网关回调处理、退款管理
- **依赖关系**: 被多个领域依赖进行支付处理

**Coupon 领域模块** (`internal/domains/coupon/`):
- **职责**: 优惠券系统、折扣计算、使用限制管理
- **核心实体**: Coupon, CouponUsage, DiscountRule
- **主要功能**: 优惠券创建、验证、使用、统计分析
- **依赖关系**: 与 Subscription 和 Payment 集成提供折扣

**Invoice 领域模块** (`internal/domains/invoice/`):
- **职责**: 发票生成、管理、PDF 生成、通知发送
- **核心实体**: Invoice, InvoiceItem, InvoiceTemplate
- **主要功能**: 发票创建、PDF 生成、状态管理、发送通知
- **依赖关系**: 基于 Payment 数据生成发票

**Ticket 领域模块** (`internal/domains/ticket/`):
- **职责**: 客户支持、工单管理、SLA 跟踪
- **核心实体**: Ticket, TicketMessage, TicketAssignment
- **主要功能**: 工单创建、消息线程、分配处理、SLA 监控
- **依赖关系**: 需要用户信息和订阅上下文

**Server 领域模块** (`internal/domains/server/`):
- **职责**: 服务器管理、健康监控、负载均衡
- **核心实体**: Server, ServerGroup, HealthCheck
- **主要功能**: 服务器配置、状态监控、流量统计、组管理
- **依赖关系**: 与 Subscription 集成提供服务器访问权限

**Referral 领域模块** (`internal/domains/referral/`):
- **职责**: 推荐系统、邀请码管理、奖励计算
- **核心实体**: Referral, InviteCode, ReferralCampaign, ReferralReward
- **主要功能**: 邀请码生成、推荐跟踪、转化统计、奖励发放
- **依赖关系**: 与 User 和 Payment 集成处理推荐奖励

## 应用层协调

**ApplicationService** (`internal/application/services/application_service.go`):
- **职责**: 作为所有领域模块的协调器，提供统一的应用程序接口
- **主要功能**: 
  - 领域模块初始化和依赖注入
  - 跨领域业务流程协调
  - 事件总线管理
  - 系统健康检查

**业务工作流** (`internal/application/workflows/`):
- **SubscriptionWorkflow**: 完整的订阅购买流程，整合支付、优惠券、推荐
- **ReferralWorkflow**: 推荐邀请和奖励处理流程
- **其他复杂业务流程**: 涉及多个领域协作的业务场景

**应用处理器** (`internal/application/handlers/`):
- **ApplicationHandler**: 处理跨领域的 HTTP 请求
- **提供聚合的 API 接口**: 如用户仪表板、管理后台等

## 共享基础设施

**配置管理** (`internal/shared/config/`):
- 统一的配置加载和管理
- 支持环境变量、配置文件
- 提供类型安全的配置访问接口

**数据库管理** (`internal/shared/database/`):
- GORM 连接管理和配置
- 连接池优化
- 迁移管理集成

**日志系统** (`internal/shared/logger/`):
- 基于 Zap 的结构化日志
- 统一的日志格式和字段
- 多输出支持 (控制台、文件、远程)

**事件系统** (`internal/shared/events/`):
- 领域事件发布和订阅
- 异步事件处理
- 跨领域通信机制

**队列系统** (`internal/shared/queue/`):
- 基于 Redis 的任务队列
- 异步任务处理
- 重试和错误处理机制

**中间件** (`internal/shared/middleware/`):
- 认证和授权中间件
- 软删除查询中间件
- 请求日志和限流中间件

## 环境配置

**🚨 关键安全要求:**

**JWT 配置 (必需):**
- `JWT_SECRET` - **关键**: 必须设置安全随机密钥 (32+ 字符)
  - 生成方式: `openssl rand -hex 32`
  - 或使用: `go run tools/generate-jwt-key/main.go`
  - 生产环境中绝不使用默认值
- `JWT_EXPIRE_HOURS` - 令牌过期时间 (默认: 24)

**安全检查:**
- 部署前运行 `make security-check`
- 开发环境使用 `make safe-run` 或 `make safe-dev`
- 应用程序在不安全的 JWT 配置下会拒绝启动

**必需环境变量:**

**数据库配置:**
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`

**Redis 配置:**
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`

**OAuth2 提供商:**
- Google: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`
- GitHub: `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_REDIRECT_URL`
- Telegram: `TELEGRAM_BOT_TOKEN`, `TELEGRAM_REDIRECT_URL`

**日志配置:**
- `LOG_LEVEL` (debug, info, warn, error)
- `LOG_FORMAT` (text/json)
- `LOG_OUTPUT`

**服务器配置:**
- `SERVER_PORT` (默认: 8080)
- `SERVER_API_TOKEN` - UniProxy 服务器节点认证令牌 (最少 20 字符)

**数据库迁移:**
- `RUN_MIGRATION` (设置为 "true" 启用自动迁移)

## API 路由结构

**应用层路由** (`/api/v1/app`):
- `POST /app/register-with-invite` - 带邀请码的用户注册工作流
- `POST /app/subscription/purchase` - 完整订阅购买工作流
- `POST /app/referral/create-invite` - 创建推荐邀请工作流
- `GET /app/dashboard/user` - 用户仪表板聚合数据
- `GET /app/dashboard/admin` - 管理员仪表板聚合数据
- `GET /app/system/health` - 系统健康检查

**领域特定路由:**
每个领域模块都有自己的路由组，例如：
- `/api/v1/user` - 用户领域路由
- `/api/v1/auth` - 认证领域路由
- `/api/v1/subscription` - 订阅领域路由
- `/api/v1/payment` - 支付领域路由
- 等等...

## 开发最佳实践

**领域模块开发:**
1. **实体优先**: 先定义核心业务实体和业务规则
2. **接口设计**: 定义清晰的仓储和服务接口
3. **测试驱动**: 为每个用例编写单元测试
4. **Mock 实现**: 提供 Mock 服务便于测试和开发

**跨领域协作:**
1. **事件驱动**: 使用领域事件进行异步通信
2. **接口依赖**: 通过接口而不是具体实现进行依赖
3. **应用服务协调**: 复杂业务流程在应用层协调
4. **避免循环依赖**: 通过事件或应用层打破循环依赖

**代码质量 (Go 最佳实践):**
1. **命名约定**: 遵循 Go 官方命名规范
   - 包名使用小写，简短且有意义
   - 接口名通常以 -er 结尾 (如 `UserRepository`, `PaymentService`)
   - 私有成员使用小写开头，公有成员大写开头
2. **错误处理**: 
   - 使用预定义错误常量 (如 `errorz.ErrUnauthorized`)
   - 统一的错误响应格式 (`responses.HandleError`)
   - 错误包装和上下文传递
3. **依赖注入**: 
   - 使用构造函数 `NewXxx()` 注入依赖
   - 通过接口而非具体类型进行依赖
   - 使用 `fx.Module` 管理生命周期
4. **日志记录**: 
   - 结构化日志，包含请求上下文
   - 统一的日志格式和字段
   - 合理的日志级别 (debug, info, warn, error)
5. **测试**: 
   - 为每个公共方法编写单元测试
   - 使用 Mock 接口进行隔离测试
   - 集成测试验证完整业务流程

**性能考虑 (Go 特定优化):**
1. **数据库优化**: 
   - GORM 连接池配置和预编译语句
   - 合理的索引设计和查询优化
   - 使用 `BeforeCreate` 钩子自动生成 UUID
2. **缓存策略**: 
   - Redis 连接池和分片策略
   - 热点数据缓存和缓存失效策略
3. **并发处理**: 
   - 使用 goroutine 处理异步任务
   - channel 用于 goroutine 间通信
   - context.Context 传递请求上下文
4. **内存管理**: 
   - 避免内存泄漏，及时释放资源
   - 使用对象池减少 GC 压力
   - 批量操作减少网络开销

## 测试策略

**单元测试:**
- 每个领域服务都有对应的单元测试
- 使用 Mock 仓储进行业务逻辑测试
- 测试覆盖率目标 >80%

**集成测试:**
- 测试领域模块之间的集成
- 测试数据库操作的正确性
- 测试外部服务集成

**端到端测试:**
- 测试完整的业务工作流
- 验证 API 接口的正确性
- 模拟真实用户场景

## 部署和运维

**容器化部署:**
- 提供 Dockerfile 进行容器化部署
- 支持 Docker Compose 本地开发环境
- Kubernetes 部署配置

**监控和观测:**
- 结构化日志输出
- Prometheus 指标暴露
- 分布式链路追踪集成
- 健康检查端点

**扩展性考虑:**
- 水平扩展支持
- 数据库读写分离
- 缓存分片策略
- 微服务拆分准备

## 使用示例

**创建新的领域模块:**
```bash
# 1. 创建领域目录结构
mkdir -p internal/domains/newmodule/{entities,usecases/{interfaces,implementations},adapters/repositories}

# 2. 实现核心实体
# internal/domains/newmodule/entities/newmodule.go

# 3. 定义接口
# internal/domains/newmodule/usecases/interfaces/

# 4. 实现服务和仓储
# internal/domains/newmodule/usecases/implementations/
# internal/domains/newmodule/adapters/repositories/

# 5. 创建模块组装
# internal/domains/newmodule/module.go
```

**应用层工作流示例:**
```go
// 跨领域业务流程
func (s *ApplicationService) ProcessOrderWithReferral(ctx context.Context, req *OrderRequest) (*OrderResponse, error) {
    // 1. 验证用户
    user, err := s.userModule.GetUserService().GetUser(ctx, req.UserID)
    if err != nil {
        return nil, err
    }
    
    // 2. 应用优惠券
    var discount float64
    if req.CouponCode != "" {
        couponResult, err := s.couponModule.GetCouponService().ValidateAndApply(ctx, req.CouponCode, req.Amount)
        if err == nil {
            discount = couponResult.DiscountAmount
        }
    }
    
    // 3. 处理支付
    payment, err := s.paymentModule.GetPaymentService().ProcessPayment(ctx, &payment.CreatePaymentRequest{
        Amount: req.Amount - discount,
        // ...
    })
    if err != nil {
        return nil, err
    }
    
    // 4. 记录推荐转化
    if req.ReferralCode != "" {
        go s.referralModule.GetReferralService().RecordConversion(ctx, req.ReferralCode, req.Amount)
    }
    
    return &OrderResponse{
        Payment: payment,
        Discount: discount,
    }, nil
}
```

**技术特色:**
- **高内聚，低耦合**: 每个领域模块职责单一，模块间依赖最小化
- **接口驱动**: 清晰的接口边界，便于单元测试和集成测试
- **Go 惯例**: 遵循 Go 官方推荐的项目结构和代码规范
- **依赖注入**: 使用 Fx 框架进行生命周期管理

## 使用指南

**🚀 启动应用:**
```bash
# 基本启动
go run cmd/server/main.go app:serve

# 构建后运行
make build
./bin/server

# 开发模式 (启用热重载)
make dev

# 带安全检查的开发模式 (推荐)
make safe-dev

# 使用 Mock 服务进行开发
go run cmd/server/main.go app:serve -mock

# 指定端口和环境
go run cmd/server/main.go app:serve -port 9090 -env production

# 显示所有可用命令
go run cmd/server/main.go -help
```

**项目结构 (Go 最佳实践):**
```
项目根目录/
├── cmd/                   # 应用程序入口点
│   └── server/           # 服务器应用
├── internal/             # 私有应用代码
│   ├── domains/          # 业务领域 (垂直切片)
│   ├── application/      # 应用层协调
│   └── shared/           # 共享基础设施
├── pkg/                  # 可复用的库代码
├── docs/                 # 文档和 API 规范
├── migrations/           # 数据库迁移文件
├── tools/                # 工具和脚本
└── go.mod               # Go 模块定义
```

**Go 代码示例 (最佳实践):**

1. **领域服务实现:**
```go
// 用户服务接口定义
type UserService interface {
    CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
    GetUserByID(ctx context.Context, id string) (*User, error)
}

// 服务实现
type userServiceImpl struct {
    repo   UserRepository
    logger framework.Logger
}

// 构造函数注入
func NewUserService(repo UserRepository, logger framework.Logger) UserService {
    return &userServiceImpl{
        repo:   repo,
        logger: logger,
    }
}

func (s *userServiceImpl) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    s.logger.Info("Creating user", "email", req.Email)
    
    // 业务验证
    if err := req.Validate(); err != nil {
        return nil, errorz.ErrInvalidInput
    }
    
    // 创建用户
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        s.logger.Error("Failed to create user", "error", err)
        return nil, err
    }
    
    return user, nil
}
```

2. **统一错误处理:**
```go
// 预定义错误
var (
    ErrUserNotFound = &APIError{
        StatusCode: http.StatusNotFound,
        Message:    "User not found",
    }
    ErrInvalidInput = &APIError{
        StatusCode: http.StatusBadRequest,
        Message:    "Invalid input provided",
    }
)

// HTTP 处理器中使用
func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    user, err := h.service.GetUserByID(c.Request.Context(), userID)
    if err != nil {
        responses.HandleError(h.logger, c, err)
        return
    }
    
    responses.SuccessJSON(c, http.StatusOK, user)
}
```

3. **模块组装 (Fx):**
```go
// 领域模块定义
var Module = fx.Module("user",
    fx.Provide(
        NewUserRepository,
        NewUserService,
        NewUserController,
        NewUserRoute,
    ),
    fx.Invoke(RegisterUserRoutes),
)

// 主应用中使用
func main() {
    fx.New(
        user.Module,
        auth.Module,
        payment.Module,
        // ...
    ).Run()
}
```

**架构优势:**
- **可维护性**: 领域边界清晰，单一职责原则
- **可测试性**: 接口驱动设计，Mock 支持完善
- **可扩展性**: 新功能以领域为单位增加
- **团队协作**: 领域独立开发，减少冲突
- **Go 特性**: 充分利用 Go 的并发模型和类型系统

**部署和运维:**
- **容器化**: Docker 和 Docker Compose 支持
- **环境配置**: 统一的环境变量管理
- **监控**: Prometheus 指标暴露，结构化日志
- **健康检查**: `/health` 端点和服务状态监控
- **扩展性**: 水平扩展和负载均衡支持