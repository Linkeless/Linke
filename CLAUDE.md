# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述
Linke 是基于 Go 语言构建的现代化服务管理平台，采用 VSA (垂直切片架构) + 清洁架构模式。核心特性包括事件驱动架构、多层缓存系统、API版本管理、支付重试机制等。

**核心业务领域:**
- 用户管理 (`domains/user`) - 包含多级缓存的用户服务
- 身份认证 (`domains/auth`) - JWT黑名单、OAuth2、登录安全
- 订阅计费 (`domains/subscription`) - 订阅生命周期、暂停/恢复、使用量追踪
- 支付处理 (`domains/payment`) - 多网关支付、智能重试、防重放攻击
- 发票管理 (`domains/invoice`) - PDF生成、缓存下载、安全验证
- 服务器管理 (`domains/server`) - 服务器群组、shadowsocks配置

**业务流程:**
```
用户注册/认证 → 选择订阅套餐 → 创建订单 → 生成发票 → 支付处理 → 激活服务 → 使用量追踪
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
- `make migrate-status` - 检查当前迁移状态
- `make migrate-create NAME=name` - 创建新迁移文件
- `make migrate-fix-dirty VERSION=N` - 修复脏迁移状态
- `make migrate-steps STEPS=2` - 向前运行2步迁移
- `make migrate-steps STEPS=-1` - 回滚1步迁移

**开发工具:**
- `make swagger` - 手动生成 API 文档 (自动访问 `/swagger/index.html`)
- `make security-check` - 运行安全配置验证
- `go run tools/generate-jwt-key/main.go` - 生成安全的 JWT 密钥

## 架构结构

**VSA + 清洁架构设计:**
```
internal/
├── application/          # 跨领域协调层 (工作流编排)
├── domains/             # 业务领域 (垂直切片)
│   ├── auth/           # JWT管理、OAuth2、登录安全
│   ├── user/           # 用户生命周期、多级缓存
│   ├── subscription/   # 订阅管理、暂停/恢复、使用量追踪
│   ├── payment/        # 支付网关、重试机制、幂等性
│   ├── invoice/        # PDF生成、下载安全、缓存策略
│   ├── server/         # 服务器群组、shadowsocks配置
│   └── [其他]/          # coupon, ticket, referral
└── shared/             # 共享基础设施
    ├── cache/          # 多级缓存系统 (内存+Redis)
    ├── events/         # 事件驱动架构、发布订阅
    ├── versioning/     # API版本管理中间件
    ├── constants/      # 系统常量定义
    ├── entities/       # 基础实体类型
    └── handlers/       # 通用HTTP处理函数
```

**领域内部结构 (清洁架构):**
```
domains/[领域]/
├── entities/           # 业务实体 + 验证逻辑
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
- Google: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`
- GitHub: `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`
- Telegram: `TELEGRAM_BOT_TOKEN`

**安全要求:**
- 应用程序拒绝使用弱 JWT 配置启动
- 部署前运行 `make security-check`
- 开发环境使用 `make safe-dev`

## 核心架构模式

**事件驱动架构 (`shared/events`):**
- **事件总线**: 内存总线、Redis分布式总线、度量装饰器
- **事件存储**: 基于GORM的持久化事件存储，支持回放和查询
- **重要事件类型**: `subscription.created`, `payment.completed`, `invoice.generated`
- **可靠性**: 至少一次投递保证、去重处理、熔断器模式
- **监控**: 完整的事件发布/订阅指标收集

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

**支付重试机制 (`domains/payment`):**
- **智能重试**: 指数退避算法、基于失败类型的重试策略
- **失败分类**: 网络错误、网关错误、永久失败自动识别
- **幂等性**: 防重放攻击、基于缓存的去重检查
- **监控告警**: 重试成功率、处理延时、系统健康度指标

**Go 代码规范 (Uber指南):**
- 使用 `shared/entities/BaseEntity` 减少结构体重复
- 使用 `shared/constants` 统一错误消息和状态值
- 使用 `shared/handlers` 通用函数处理ID解析、JSON绑定
- 错误包装: `fmt.Errorf("context: %w", err)`
- 依赖注入: 每个领域独立的 `fx.Module` 配置

## 关键开发要点

**启动应用程序:**
- 应用程序使用uber/fx进行依赖注入，启动顺序: 配置 → 数据库 → 缓存 → 事件系统 → 业务领域 → HTTP路由
- 支持集成式数据库迁移: 使用 `-migrate-command` 参数直接运行迁移
- 支持多种运行模式: 普通模式、迁移模式、安全检查模式

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
- `/api/v1/subscriptions/*` - 订阅服务
- `/api/v1/payments/*` - 支付处理
- `/api/v1/admin/*` - 管理员功能

**常见问题解决:**
- **脏迁移**: `make migrate-fix-dirty VERSION=X`
- **JWT配置错误**: 检查 `JWT_SECRET` 长度 ≥32 字符
- **缓存连接失败**: 检查Redis连接配置和网络
- **支付重试失败**: 查看重试配置和失败分类规则
- **事件处理异常**: 检查事件总线配置和订阅者注册