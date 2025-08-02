# CLAUDE.md

此文件为 Claude Code (claude.ai/code) 在此代码库中工作提供指导。

## 项目概述
Linke 是基于 Go 语言构建的服务管理平台，采用 VSA (垂直切片架构) + 清洁架构模式。提供订阅计费、用户管理、服务器管理、shadowsocks 服务器、多网关支付、推荐系统和工单系统。

**核心业务流程:**
```
用户选择服务 → 创建订单 → 生成发票 → 用户付款 → 激活服务
```

## 核心命令

**开发:**
- `make dev` - 开发模式运行 (生成 Swagger，推荐)
- `make safe-dev` - 带安全检查的开发模式 (最安全)
- `make build` - 构建到 `bin/server`
- `go run cmd/server/main.go` - 直接启动服务器

**测试:**
- `make test` - 运行所有测试，详细输出
- `go test ./internal/domains/user` - 测试特定领域模块
- `go test -run TestFunctionName` - 运行特定测试函数
- `go test -race ./...` - 竞态条件检测

**数据库:**
- `make migrate-up` - 运行待执行迁移
- `make migrate-status` - 检查迁移状态
- `make migrate-create NAME=name` - 创建新迁移
- `make migrate-fix-dirty VERSION=N` - 修复脏迁移状态

**工具:**
- `make swagger` - 生成 API 文档 (访问 `/swagger/*any`)
- `make security-check` - 验证安全配置
- `go run tools/generate-jwt-key/main.go` - 生成安全的 JWT 密钥

## 架构结构

**目录结构:**
```
internal/
├── application/          # 跨领域协调
├── domains/             # 业务领域 (VSA)
│   ├── user/           # 用户生命周期
│   ├── auth/           # 身份认证
│   ├── subscription/   # 计费和套餐
│   ├── payment/        # 支付处理
│   ├── server/         # 服务器管理
│   └── [其他]/          # coupon, invoice, ticket, referral
└── shared/             # 基础设施
    ├── config/         # 配置管理
    ├── database/       # 数据库和迁移
    ├── logger/         # 结构化日志
    └── middleware/     # HTTP 中间件
```

**领域模式 (清洁架构):**
```
domains/[领域]/
├── entities/           # 业务实体
├── usecases/          # 业务逻辑
│   ├── interfaces/    # 接口契约
│   └── implementations/ # 服务实现
├── adapters/          # 外部接口
│   └── repositories/ # 数据访问
└── module.go          # 依赖注入
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

## 核心实现模式

**Go 代码风格 (Uber 指南):**
- 函数顺序: 类型 → 构造函数 → 方法 → 工具函数
- 错误包装: `fmt.Errorf("上下文: %w", err)`
- 结构体初始化: 总是指定字段名，使用 `&T{}` 而非 `new(T)`
- 空切片: 返回 `nil` 而非 `[]T{}`
- 互斥锁: 作为字段 `mu sync.Mutex`，不嵌入

**依赖管理:**
- 使用 `fx.Module` 进行依赖注入
- 每个领域都是自包含模块
- 应用层协调跨领域工作流

**测试方法:**
- 表驱动测试配合 `t.Run` 子测试
- 目标覆盖率 >80%
- 单元测试使用 Mock 接口

## API 结构
- `/health` - 健康检查
- `/swagger/*any` - API 文档
- `/api/v1/app` - 应用层端点
- `/api/v1/[领域]` - 领域特定端点

## 常见问题
- **脏迁移**: 使用 `make migrate-fix-dirty VERSION=X`
- **JWT 错误**: 检查 `JWT_SECRET` 长度和唯一性
- **测试失败**: 确保数据库正在运行 (集成测试)