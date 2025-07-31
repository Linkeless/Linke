# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

Linke 是一个基于 Go 语言开发的综合性服务管理平台，提供订阅制计费、用户管理和服务器管理功能。平台集成了 OAuth2 认证、流量订阅管理、多网关支付、推荐系统和客户支持系统。

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

### 数据库迁移管理
```bash
# 查看所有迁移命令帮助
make migrate-help

# 运行所有待执行的迁移
make migrate-up

# 回滚一个迁移
make migrate-down

# 重置数据库（删除所有表并重新运行迁移）
make migrate-reset

# 查看当前迁移状态
make migrate-status

# 创建新的迁移文件
make migrate-create NAME=your_migration_name

# 强制设置迁移版本
make migrate-force VERSION=N

# 迁移到指定版本
make migrate-goto VERSION=N

# 执行指定步数的迁移
make migrate-steps STEPS=N   # 正数向上，负数向下

# 修复脏迁移状态
make migrate-fix-dirty VERSION=N
```

### 开发工具
```bash
# 运行安全检查
make security-check

# 生成 JWT 密钥
go run tools/generate-jwt-key/main.go
```

## 架构概览

### 核心技术栈
- **Web 框架**: Gin (HTTP 路由和中间件)
- **数据库**: MySQL + GORM (ORM)
- **缓存**: Redis
- **任务队列**: Asynq (基于 Redis)
- **认证**: JWT + OAuth2 (Google, GitHub)
- **文档**: Swagger/OpenAPI
- **日志**: Zap
- **迁移**: golang-migrate

### 项目结构模式

项目采用清洁架构设计，分层明确：

```
cmd/server/           # 应用程序入口点
config/              # 配置管理
internal/
├── handler/         # HTTP 处理层（控制器）
├── service/         # 业务逻辑层
├── repository/      # 数据访问层  
├── model/          # 数据模型
├── middleware/     # HTTP 中间件
├── modules/        # 模块管理器
├── routes/         # 路由配置
├── queue/          # 任务队列
├── migration/      # 数据库迁移
├── logger/         # 日志工具
└── response/       # 统一响应格式
```

### 模块化管理架构

管理端采用模块化架构，参考 `internal/handler/admin/user/` 目录结构：

```
admin/[module]/
├── README.md           # 模块文档
├── manager.go          # 统一管理器（向后兼容）
├── shared/             # 共享组件
│   ├── base.go        # 基础处理器
│   └── validator.go   # 参数验证器
├── management/         # CRUD 操作
├── query/             # 查询操作
├── operation/         # 业务操作
├── status/            # 状态管理
└── statistics/        # 统计功能
```

**已模块化的管理端功能**:
- `admin/user/` - 用户管理（完整模块化示例）
- `admin/coupon/` - 优惠券管理
- `admin/invoice/` - 发票管理  
- `admin/ticket/` - 工单管理
- `admin/order/` - 订单管理

### 依赖注入模式

通过 `internal/modules/manager_simple.go` 统一管理所有依赖注入：

1. **服务层依赖**: 按顺序初始化所有 Service
2. **处理器依赖**: 基于 Service 初始化所有 Handler
3. **模块化处理器**: 使用别名导入避免命名冲突

### 数据库设计

#### 核心业务实体
- **用户系统**: User, OAuth 集成
- **订阅系统**: SubscriptionPlan, UserSubscription, SubscriptionOrder
- **支付系统**: PaymentRecord, PaymentConfig, 退款管理
- **推荐系统**: Referral, ReferralCampaign, ReferralReward
- **支持系统**: Ticket, TicketMessage
- **服务器管理**: ServerGroup, ShadowsocksServer
- **优惠券系统**: Coupon, CouponUsage
- **发票系统**: Invoice（独立于订单）

#### 业务流程设计
遵循标准商务流程：
```
用户选择服务 → 创建订单 → 生成发票 → 用户付款 → 确认收款 → 激活服务
```

### API 设计模式

#### 路由组织
- `/api/v1/auth/*` - 认证相关
- `/api/v1/admin/*` - 管理端 API
- `/api/v1/user/*` - 用户端 API  
- `/api/v1/server/*` - 服务器 API

#### 响应格式
统一使用 `internal/response/` 中定义的响应格式：
- 成功响应: `StandardResponse`, `StandardListResponse`
- 错误响应: `BadRequestResponse`, `UnauthorizedResponse` 等

#### 认证方式
- **用户认证**: JWT Bearer Token
- **服务器认证**: API Token
- **OAuth2**: Google, GitHub 第三方登录

## 开发规范

### 模块化开发
当需要重构大型 handler 文件时，参考 `admin/user/` 的模块化模式：

1. **创建目录结构**: 按照标准模块化目录组织
2. **共享组件**: 实现 BaseHandler 和 Validator
3. **功能分组**: 按业务职责分解到不同子模块
4. **向后兼容**: 通过 Manager 提供兼容接口
5. **更新依赖**: 在 `manager_simple.go` 中更新引用

### 数据库迁移
- 迁移文件位于 `migrations/` 目录
- 使用 `make migrate-create NAME=xxx` 创建新迁移
- 迁移命令集成在主程序中，通过命令行参数执行
- 环境变量通过 `.env` 文件配置

### 错误处理
- 使用统一的日志格式（Zap）
- HTTP 错误通过 `response` 包统一处理
- 业务逻辑错误在 Service 层处理

### 安全考虑
- 所有敏感操作需要认证中间件
- 管理端操作需要管理员权限验证
- 提供安全检查脚本 `scripts/security-check.sh`
- OAuth2 状态验证防止 CSRF 攻击

## 环境配置

项目通过环境变量配置，主要配置项：
- 数据库连接: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- Redis 配置: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- OAuth2 密钥: `GOOGLE_CLIENT_ID`, `GITHUB_CLIENT_ID` 等
- JWT 配置: `JWT_SECRET_KEY`
- 服务器配置: `SERVER_PORT`, `API_SERVER_TOKEN`

配置文件通过 `config/config.go` 统一管理，支持 `.env` 文件自动加载。