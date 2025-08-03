# Linke 服务管理平台

基于 Go 语言构建的现代化服务管理平台，采用 VSA (垂直切片架构) + 清洁架构模式。

## 核心功能

- 🔐 **用户管理** - 完整的用户注册、认证和授权系统
- 💳 **订阅计费** - 灵活的订阅服务和计费管理
- 💰 **多网关支付** - 支持多种支付方式和网关
- 🖥️ **服务器管理** - 服务器资源管理和监控
- 🎫 **工单系统** - 完整的客户支持工单系统
- 🎁 **推荐系统** - 用户推荐和奖励机制

## 技术特性

- 🏗️ **现代架构** - VSA + Clean Architecture 架构模式
- 🔄 **事件驱动** - 完整的事件发布/订阅机制
- 🚀 **缓存优化** - Redis 缓存层，多种缓存模式
- 📊 **API 版本管理** - 灵活的 API 版本控制策略
- 🔒 **安全加固** - 多层安全防护和数据保护
- ⚡ **高性能** - 优化的数据库连接池和查询

## 快速开始

### 开发环境

```bash
# 开发模式运行（推荐）
make dev

# 带安全检查的开发模式（最安全）
make safe-dev

# 直接构建运行
make build && ./bin/server
```

### 数据库迁移

```bash
# 运行待执行迁移
make migrate-up

# 检查迁移状态
make migrate-status

# 创建新迁移
make migrate-create NAME=migration_name
```

### 测试

```bash
# 运行所有测试
make test

# 测试特定模块
go test ./internal/domains/user

# 竞态条件检测
go test -race ./...
```

## 架构说明

项目采用现代化的 VSA + Clean Architecture 架构模式：

```
internal/
├── application/          # 跨领域协调层
├── domains/             # 业务领域（VSA）
│   ├── auth/           # 身份认证
│   ├── user/           # 用户管理
│   ├── payment/        # 支付处理
│   ├── subscription/   # 订阅服务
│   └── [其他领域]/
└── shared/             # 共享基础设施
    ├── config/         # 配置管理
    ├── cache/          # 缓存系统
    ├── events/         # 事件系统
    └── versioning/     # API版本管理
```

## 环境配置

项目需要配置以下关键环境变量：

```bash
# JWT 密钥（必须32+字符）
JWT_SECRET="your-super-secure-jwt-secret-key"

# 数据库配置
DB_HOST="localhost"
DB_PORT="3306"
DB_USER="linke_user"
DB_PASSWORD="your_password"
DB_NAME="linke_db"

# Redis 配置
REDIS_HOST="localhost"
REDIS_PORT="6379"
REDIS_PASSWORD=""
REDIS_DB="0"
```

## 文档

- 📖 [项目指南](./CLAUDE.md) - 完整的项目使用和开发指南
- 🏆 [最佳实践](./BEST_PRACTICES.md) - 开发最佳实践和架构指南
- 🔄 [API版本管理](./docs/API_VERSIONING_GUIDE.md) - API版本控制指南
- 💾 [缓存策略](./CACHING_BEST_PRACTICES.md) - 缓存实现最佳实践
- 📡 [事件驱动](./EVENT_DRIVEN_ARCHITECTURE.md) - 事件驱动架构说明
- 🔒 [支付安全](./docs/PAYMENT_SECURITY_GUIDE.md) - 支付安全指南

## API 文档

启动服务后访问 Swagger 文档：

```
http://localhost:8080/swagger/index.html
```

## 健康检查

```bash
curl http://localhost:8080/health
```

## 许可证

本项目采用私有许可证。

## 贡献

欢迎提交 Issue 和 Pull Request。请遵循项目的代码规范和最佳实践。