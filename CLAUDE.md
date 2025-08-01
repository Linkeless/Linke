# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

Linke 是一个基于 Go 语言开发的综合性服务管理平台，提供订阅制计费、用户管理和服务器管理功能。平台集成了 OAuth2 认证、流量订阅管理、多网关支付、推荐系统和客户支持系统。

## 核心技术栈
- **Web 框架**: Gin
- **数据库**: MySQL + GORM
- **缓存**: Redis  
- **任务队列**: Asynq
- **认证**: JWT + OAuth2
- **文档**: Swagger
- **日志**: Zap
- **迁移**: golang-migrate

## 常用命令

### 构建和运行
```bash
# 构建项目
make build

# 运行项目
make run

# 开发模式运行（包含 Swagger 文档生成）
make dev

# 安全运行（包含安全检查）
make safe-run
make safe-dev

# 运行测试
make test

# 生成 Swagger 文档
make swagger

# 安装依赖和工具
make install

# 清理构建产物
make clean
```

### 数据库迁移
```bash
make migrate-up      # 运行迁移
make migrate-down    # 回滚迁移
make migrate-reset   # 重置数据库
make migrate-status  # 查看状态
make migrate-create NAME=name  # 创建迁移
```

### 开发工具
```bash
make security-check  # 安全检查
make test           # 运行所有测试  
make swagger        # 生成文档
make install        # 安装依赖和工具
make clean          # 清理构建产物

# Note: 以下命令目前在 Makefile 中未实现
# make test-unit      # 运行单元测试
# make test-integration # 运行集成测试
# make test-coverage  # 生成测试覆盖率报告
# make benchmark      # 运行性能基准测试
# make fmt            # 格式化代码
# make lint           # 代码静态分析
# make deps-check     # 检查依赖更新
```

## 架构设计

### 垂直切片架构 (DDD)

项目采用垂直切片架构与DDD结合的设计模式，按业务领域垂直切分：

```
internal/
├── user/               # 用户领域（包含认证）
├── subscription/       # 订阅领域  
├── payment/           # 支付领域
├── ticket/            # 工单领域
├── server/            # 服务器领域
├── invoice/           # 发票领域
├── coupon/            # 优惠券领域
└── shared/            # 跨领域共享
    ├── domain/        # 跨领域的领域概念
    └── infra/         # 共享基础设施
```

注意：当前项目缺少 `internal/shared/valueobject/` 目录，存在跨域值对象重复问题。

### 严格 DDD 领域内分层
```
domain/                    # 领域层
├── aggregate/            # 聚合根：业务一致性边界
├── entity/              # 实体：有唯一标识的业务对象
├── valueobject/         # 值对象：不可变的业务概念
├── event/               # 领域事件：业务状态变化通知  
├── repository/          # 仓储接口：数据访问抽象
└── service/             # 领域服务：跨实体业务逻辑

service/                  # 应用层
├── command/             # 命令处理器：写操作
└── query/               # 查询处理器：读操作

handler/                 # 接口层：HTTP处理器
infra/                   # 基础设施层
├── persistence/         # 数据持久化实现
├── event/              # 事件发布器实现
└── external/           # 外部服务集成
```

### 核心业务实体
- **用户系统**: User（包含认证信息）, OAuth 集成, Session 管理
- **订阅系统**: SubscriptionPlan, UserSubscription, SubscriptionOrder  
- **支付系统**: PaymentRecord, PaymentConfig
- **推荐系统**: Referral, ReferralCampaign, ReferralReward
- **支持系统**: Ticket, TicketMessage
- **服务器管理**: ServerGroup, ShadowsocksServer
- **优惠券系统**: Coupon, CouponUsage
- **发票系统**: Invoice（独立于订单）

### 业务流程

**标准商务流程** (参考 `docs/BUSINESS_FLOW_REDESIGN.md`)：
```
用户选择服务 → 创建订单 → 生成发票 → 用户付款 → 确认收款 → 激活服务
```

**核心实体关系**：
```
User → Order → Invoice → Payment → Subscription
  ↓       ↓        ↓        ↓         ↓
订阅者   购买意向   付款请求   资金流转   服务履行
```

注意：项目当前处于业务流程重构阶段，从 "订单 → 付款 → 发票" 的错误流程迁移到标准商务流程。

## API 设计

### 路由组织
- `/api/v1/user/auth/*` - 用户认证相关（登录、注册、OAuth）
- `/api/v1/admin/*` - 管理端 API
- `/api/v1/user/*` - 用户端 API  
- `/api/v1/server/*` - 服务器 API

### 认证方式
- **用户认证**: JWT Bearer Token
- **服务器认证**: API Token  
- **OAuth2**: Google, GitHub 第三方登录

### 响应格式
统一使用 `internal/response/` 中定义的响应格式：
- 成功: `APIResponse`, `ListResponse`
- 错误: 统一使用 `Error()` 函数和预定义的错误响应函数
- 分页: `SuccessList()` 和相关分页函数

## 开发规范

### 严格 DDD 设计原则

#### 聚合根设计
- **一致性边界**: 聚合根是事务一致性的唯一边界
- **单一事务**: 一次事务只能修改一个聚合根实例
- **身份标识**: 聚合根拥有全局唯一标识符
- **状态管理**: 聚合根负责管理内部实体和值对象的状态
- **业务规则**: 所有业务不变性在聚合根中维护和验证

#### 实体与值对象
- **实体特征**: 具有唯一标识符，生命周期内可变
- **值对象特征**: 无标识符，不可变，通过值比较相等性
- **类型安全**: 使用值对象包装原始类型，避免原始类型滥用
- **业务语义**: 值对象承载业务含义和验证规则

#### 领域事件驱动
- **状态变化**: 聚合根状态变化时产生领域事件
- **最终一致性**: 跨聚合的一致性通过事件实现
- **副作用隔离**: 业务副作用通过事件处理器异步执行
- **解耦通信**: 聚合间通过事件进行松耦合通信

#### 依赖方向控制
- **领域独立**: 领域层不依赖任何外层
- **接口抽象**: 通过仓储接口抽象数据访问
- **依赖注入**: 基础设施实现通过依赖注入提供
- **薄应用层**: 应用层协调调用，不包含业务逻辑

### 严格 DDD 开发流程

#### 领域建模阶段
1. **业务分析** - 识别业务用例和领域概念
2. **聚合识别** - 确定一致性边界和聚合根
3. **实体建模** - 设计实体的身份标识和生命周期
4. **值对象设计** - 封装业务概念和验证规则
5. **事件设计** - 定义领域状态变化事件

#### 实现阶段
6. **值对象实现** - 优先实现不可变值对象
7. **实体实现** - 实现具有标识的业务实体
8. **聚合根实现** - 封装业务不变性和状态管理
9. **领域服务** - 实现跨实体的复杂业务逻辑
10. **仓储接口** - 定义数据访问抽象

#### 应用层实现
11. **命令处理器** - 处理写操作和业务流程协调
12. **查询处理器** - 处理读操作和数据投影
13. **事件处理器** - 处理领域事件的副作用

#### 基础设施实现  
14. **仓储实现** - 数据持久化和查询实现
15. **事件发布器** - 领域事件分发机制
16. **外部服务** - 第三方系统集成

#### 接口层实现
17. **HTTP处理器** - API端点和数据传输对象
18. **集成测试** - 端到端业务流程验证

### 测试策略
```
           [E2E Tests]          ← 少量，完整业务流程
         [Integration Tests]     ← 适量，跨层协作测试  
    [Unit Tests - Application]   ← 较多，用例测试
  [Unit Tests - Domain Layer]    ← 最多，领域逻辑测试
```

## 代码规范

### DDD 实现规范

#### 聚合根实现规范
```go
// 聚合根必须包含以下元素
type SomeAggregate struct {
    // 1. 唯一标识符
    id valueobject.SomeID
    
    // 2. 业务状态字段
    status valueobject.SomeStatus
    
    // 3. 审计字段
    createdAt time.Time
    updatedAt time.Time
    deletedAt *time.Time
    
    // 4. 领域事件容器
    domainEvents []event.DomainEvent
    
    // 5. 管理的实体（可选）
    entities map[EntityID]*Entity
}

// 必须提供的方法
func (a *SomeAggregate) ID() valueobject.SomeID           // 标识符访问
func (a *SomeAggregate) DomainEvents() []event.DomainEvent // 事件访问
func (a *SomeAggregate) ClearDomainEvents()               // 事件清理
func (a *SomeAggregate) IsDeleted() bool                  // 软删除检查
```

#### 值对象实现规范
```go
// 值对象必须是不可变的
type SomeValueObject struct {
    value string // 私有字段，防止外部修改
}

// 构造函数进行验证
func NewSomeValueObject(value string) (SomeValueObject, error) {
    if value == "" {
        return SomeValueObject{}, fmt.Errorf("value cannot be empty")
    }
    return SomeValueObject{value: value}, nil
}

// 提供访问器，不提供修改器
func (vo SomeValueObject) Value() string { return vo.value }
func (vo SomeValueObject) String() string { return vo.value }
func (vo SomeValueObject) Equals(other SomeValueObject) bool {
    return vo.value == other.value
}
```

#### 实体实现规范
```go
// 实体具有标识符和可变状态
type SomeEntity struct {
    id     EntityID    // 唯一标识符
    // 业务状态字段...
    createdAt time.Time
    updatedAt time.Time
}

// 必须提供 ID 访问器
func (e *SomeEntity) ID() EntityID { return e.id }
```

### Go 命名规范
- **包名**: 小写，简洁，无下划线（如：`user`, `payment`）
- **接口**: 以 `er` 结尾（如：`UserRepository`, `EventPublisher`）
- **结构体**: 大驼峰命名（如：`UserService`）
- **常量**: 大写字母加下划线（如：`MAX_RETRY_COUNT`）
- **聚合根**: 领域名词（如：`Coupon`, `User`, `Order`）
- **值对象**: 领域概念（如：`CouponCode`, `Email`, `Money`）

### 错误处理规范

#### 领域错误定义
```go
// 领域特定错误
var (
    ErrSomeNotFound     = errors.New("some not found")
    ErrSomeAlreadyExists = errors.New("some already exists")
    ErrInvalidSomeState = errors.New("invalid some state")
)

// 业务规则违反错误
type BusinessRuleViolationError struct {
    Rule    string
    Message string
}

func (e BusinessRuleViolationError) Error() string {
    return fmt.Sprintf("business rule violation [%s]: %s", e.Rule, e.Message)
}
```

#### 错误处理最佳实践
- 聚合根方法返回明确的领域错误
- 使用 `errors.Is()` 检查特定错误类型
- 使用 `fmt.Errorf("context: %w", err)` 包装错误
- 应用层负责错误类型到 HTTP 状态码的转换

### 领域事件规范

#### 事件定义
```go
// 事件必须实现 DomainEvent 接口
type SomeEvent struct {
    eventID   string
    occurredAt time.Time
    // 事件数据字段...
}

func (e SomeEvent) EventID() string { return e.eventID }
func (e SomeEvent) OccurredAt() time.Time { return e.occurredAt }
func (e SomeEvent) EventType() string { return "SomeEvent" }
```

#### 事件发布模式
- 聚合根状态变化时添加事件到 `domainEvents` 切片
- 应用层在事务提交前发布事件
- 事件处理器异步处理副作用

### 安全规范
- 密码使用 bcrypt 加密存储
- JWT Token 有效期控制
- 输入验证使用 validator 包
- 使用 context 超时控制

### 日志规范

#### Zap 日志使用
- 使用结构化日志记录重要业务事件
- 日志级别：DEBUG < INFO < WARN < ERROR < FATAL
- 生产环境建议使用 INFO 级别以上

#### 日志记录最佳实践
```go
// 记录业务操作
logger.Info("user login successful",
    zap.String("user_id", userID.String()),
    zap.String("ip", clientIP),
    zap.Duration("duration", time.Since(start)))

// 记录错误信息
logger.Error("failed to process payment", 
    zap.Error(err),
    zap.String("payment_id", paymentID),
    zap.String("user_id", userID.String()))
```

#### 敏感信息处理
- 禁止记录密码、Token 等敏感信息
- 个人信息需脱敏处理
- 支付信息仅记录订单号，不记录具体金额

## 环境配置

### 主要配置项
- **数据库**: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- **Redis**: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`
- **OAuth2**: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`
- **JWT**: `JWT_SECRET` (必须，至少32字符，强制安全检查)
- **服务器**: `SERVER_PORT`, `SERVER_API_TOKEN`
- **日志**: `LOG_LEVEL`, `LOG_FORMAT`, `LOG_OUTPUT`

配置通过 `config/config.go` 统一管理，支持 `.env` 文件加载。

### 安全要求
- `JWT_SECRET` 是必需的，不能为空，最少32个字符
- 禁止使用默认的弱密钥值
- 建议使用 `openssl rand -hex 32` 生成安全密钥

## 重要问题 - 跨域值对象重复

⚠️ **当前项目存在严重的跨域值对象重复问题，违反了 DDD 原则**

### 发现的重复值对象

#### 1. UserID - 5个不同实现
- `user/domain/valueobject/user_id.go` - UUID string 类型，最完整
- `ticket/domain/valueobject/user_id.go` - uint 类型，JSON 序列化
- `subscription/domain/valueobject/user_id.go` - uint 类型，返回指针+错误
- `coupon/domain/valueobject/user_id.go` - uint64 类型，零值 panic
- `payment/domain/valueobject/user_id.go` - uint 类型，返回值+错误

#### 2. Money - 3个不同实现
- `invoice/domain/valueobject/money.go` - Currency 结构体，完整操作
- `coupon/domain/valueobject/money.go` - Currency string，JSON 支持
- `payment/domain/valueobject/money.go` - 独立 Currency，加密货币

#### 3. Currency - 2个不同实现
- `subscription/domain/valueobject/currency.go` - string 类型
- `payment/domain/valueobject/currency.go` - 结构体，法币+加密货币

#### 4. InvoiceID - 2个不同实现
- `invoice/domain/valueobject/invoice_id.go` - 不验证零值
- `payment/domain/valueobject/invoice_id.go` - 验证零值

### 解决方案

**必须立即重构**：
1. **统一 UserID** - 迁移到 `internal/shared/valueobject/user_id.go`
2. **统一 Money/Currency** - 迁移到 `internal/shared/valueobject/money.go` 和 `currency.go`  
3. **统一 InvoiceID** - 迁移到 `internal/shared/valueobject/invoice_id.go`

**重构原则**：
- 选择功能最完整、设计最好的实现作为标准版本
- 所有领域统一引用 `internal/shared/valueobject` 中的值对象
- 移除各领域内的重复实现
- 确保类型兼容性和 JSON 序列化支持

这是一个**高优先级**的架构债务，必须在继续开发前解决。

## 项目当前状态

### Git 分支结构
- **Main 分支**: `main` - 生产分支
- **当前开发分支**: `dev` - 活跃开发分支

### 最近的重要变更
- 实现了用户工单处理的模块化结构
- 添加了工单统计和状态管理处理器
- 用户管理模块的管理员功能
- 业务流程重构正在进行中（参考 BUSINESS_FLOW_REDESIGN.md）

### 开发注意事项
1. **测试运行**: 使用 `go test -v ./...` 运行测试
2. **数据库迁移**: 迁移通过主程序命令行参数执行，而不是独立工具
3. **安全检查**: 运行前必须通过安全检查 (`make security-check`)
4. **文档生成**: 使用 swag 生成 Swagger 文档
5. **依赖管理**: 使用 Go modules，通过 `make install` 安装工具依赖