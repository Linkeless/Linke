# 管理端用户模块化架构

本目录包含了模块化的管理端用户管理功能，将原来的单一大文件（855行）按照业务功能拆分为多个小模块。

## 📁 目录结构

```
internal/handler/admin/user/
├── README.md              # 本文档
├── manager.go             # 统一管理器，提供向后兼容性
├── shared/                # 共享组件
│   ├── base.go           # 基础处理器，提供通用依赖
│   └── validator.go      # 验证工具，处理参数验证
├── management/           # 用户基础管理
│   └── user_crud.go     # 用户CRUD操作 (创建、获取、更新、补丁更新)
├── status/               # 用户状态管理
│   └── user_status.go   # 状态变更 (角色、状态、密码重置)
├── query/                # 用户查询
│   ├── user_list.go     # 列表和分页 (用户列表、已删除用户、按提供商过滤)
│   └── user_search.go   # 搜索功能
├── operation/            # 用户操作
│   ├── user_delete.go   # 删除管理 (软删除、硬删除、恢复)
│   └── user_batch.go    # 批量操作 (批量删除、批量恢复)
└── statistics/           # 用户统计
    └── user_stats.go    # 统计信息
```

## 🔧 模块职责

### 1. 共享组件 (`shared/`)
- **BaseHandler**: 提供通用依赖注入（UserService, AuthService, Validator）
- **UserValidator**: 统一的参数验证逻辑，避免重复代码

### 2. 用户基础管理 (`management/`)
- 创建用户 (`CreateUser`)
- 获取用户信息 (`GetUser`) 
- 更新用户完整信息 (`UpdateUser`)
- 部分更新用户信息 (`PatchUser`)

### 3. 用户状态管理 (`status/`)
- 更新用户角色 (`UpdateUserRole`)
- 更新用户状态 (`UpdateUserStatus`)
- 重置用户密码 (`ResetUserPassword`)

### 4. 用户查询 (`query/`)
- 用户列表分页 (`ListUsers`)
- 已删除用户列表 (`ListDeletedUsers`)
- 按OAuth提供商过滤 (`ListUsersByProvider`)
- 用户搜索 (`SearchUsers`)

### 5. 用户操作 (`operation/`)
- 软删除用户 (`SoftDeleteUser`)
- 恢复用户 (`RestoreUser`)
- 硬删除用户 (`HardDeleteUser`)
- 批量删除用户 (`BatchDeleteUsers`)
- 批量恢复用户 (`BatchRestoreUsers`)

### 6. 用户统计 (`statistics/`)
- 获取用户统计信息 (`GetUserStats`)

## 🔄 向后兼容性

`manager.go` 文件提供 `AdminUserManager` 结构体，它包含所有子模块并提供与原始 `AdminUserHandler` 相同的方法接口。这确保了：

1. **无缝迁移**: 现有代码无需修改即可使用新的模块化结构
2. **渐进式重构**: 可以逐步重构其他部分来直接使用子模块
3. **测试兼容**: 现有测试用例仍然有效

## 🎯 设计原则

### 单一职责原则
每个模块只负责一个特定的业务领域，降低了复杂性和耦合度。

### 依赖注入
通过 `BaseHandler` 统一管理依赖，便于测试和维护。

### 代码复用
`UserValidator` 提供了通用的验证逻辑，避免在各个模块中重复实现。

### 一致性
所有模块遵循相同的错误处理、日志记录和响应格式模式。

## 📈 优势

1. **可维护性**: 小文件更容易理解和修改
2. **可测试性**: 每个模块可以独立测试
3. **可扩展性**: 新功能可以作为新模块添加
4. **团队协作**: 不同开发者可以同时工作在不同模块上
5. **代码质量**: 更好的代码组织和职责分离

## 🚀 使用示例

### 直接使用子模块 (推荐新代码)
```go
// 创建管理器
userManager := user.NewAdminUserManager(userService, authService)

// 直接使用子模块
userManager.Management.CreateUser(c)
userManager.Status.UpdateUserRole(c) 
userManager.Query.ListUsers(c)
```

### 使用兼容接口 (现有代码)
```go
// 现有代码无需修改
userManager := user.NewAdminUserManager(userService, authService)
userManager.CreateUser(c) // 内部路由到 Management.CreateUser
```

## 📝 迁移说明

原始的 `user.go` 文件已重命名为 `user.go.bak` 作为备份。新的模块化结构完全替代了原有功能，并保持了 API 的向后兼容性。