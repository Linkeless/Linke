# API 迁移指南

本指南帮助客户端在 Linke API 的不同版本之间进行迁移。

## 概述

Linke API 使用语义化版本控制，并提供多种迁移策略以确保版本间的平滑过渡。本指南涵盖了版本间的主要差异，并提供逐步迁移说明。

## 版本历史

### 版本 1.0.0（已弃用）

- **状态**: 已弃用
- **下线日期**: 2025年12月31日
- **描述**: 具有基本功能的初始 API 版本

### 版本 2.0.0（当前版本）

- **状态**: 活跃
- **发布日期**: 2024年1月1日
- **描述**: 具有改进数据结构和新功能的增强 API

## 版本检测

API 会自动检测您请求的版本并提供适当的响应。您可以使用以下方式指定版本：

### URL 路径（推荐）

```
GET /api/v1/users    # 版本 1.0.0
GET /api/v2/users    # 版本 2.0.0
```

### HTTP 头部

```bash
GET /api/users
X-API-Version: v2
```

### 查询参数

```
GET /api/users?version=v2
```

## 从 v1 迁移到 v2

### 破坏性变更

#### 1. 用户数据结构

**v1 响应:**
```json
{
  "id": "123",
  "name": "John Doe",
  "email": "john@example.com"
}
```

**v2 响应:**
```json
{
  "user_id": "123",
  "full_name": "John Doe",
  "email": "john@example.com",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "profile": {
    "avatar_url": "https://example.com/avatar.jpg",
    "bio": "Software Engineer"
  }
}
```

**迁移步骤:**
1. 更新字段映射：`id` → `user_id`，`name` → `full_name`
2. 处理新字段：`created_at`、`updated_at`、`profile`
3. 更新客户端代码以解析嵌套的 `profile` 对象

#### 2. 订阅数据结构

**v1 响应:**
```json
{
  "id": "sub_123",
  "plan_id": "basic",
  "status": "active",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

**v2 响应:**
```json
{
  "subscription_id": "sub_123",
  "subscription_plan": {
    "plan_id": "basic",
    "plan_name": "Basic Plan",
    "price": 9.99,
    "currency": "USD",
    "billing_cycle": "monthly",
    "features": [
      "10GB traffic",
      "Basic support",
      "5 devices"
    ]
  },
  "subscription_status": "active",
  "expiry_date": "2024-12-31T23:59:59Z",
  "auto_renewal": true,
  "usage": {
    "current_usage": "1.2GB",
    "usage_limit": "10GB",
    "usage_percent": 12.0
  },
  "billing_history": [
    {
      "invoice_id": "inv_456",
      "amount": 9.99,
      "currency": "USD",
      "paid_at": "2024-01-01T00:00:00Z",
      "status": "paid"
    }
  ]
}
```

**迁移步骤:**
1. 更新字段映射：`id` → `subscription_id`，`status` → `subscription_status`，`expires_at` → `expiry_date`
2. 处理嵌套的 `subscription_plan` 对象而不是平面的 `plan_id`
3. 处理新字段：`auto_renewal`、`usage`、`billing_history`

#### 3. 新端点

**v2 专有端点:**
- `GET /api/v2/analytics` - 使用情况分析（v1 中不可用）
- `GET /api/v2/profile/preferences` - 用户偏好设置管理

### 非破坏性增强

#### 1. 服务器状态端点

**v1 响应:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "uptime": "99.9%"
}
```

**v2 响应（增强版）:**
```json
{
  "server_status": "healthy",
  "last_checked": "2024-01-01T12:00:00Z",
  "availability": "99.9%",
  "system_info": {
    "version": "2.1.0",
    "environment": "production",
    "region": "us-east-1"
  },
  "metrics": {
    "cpu_usage": "15.5%",
    "memory_usage": "342MB",
    "disk_usage": "45%"
  }
}
```

**迁移步骤:**
1. 更新增强格式的字段映射
2. 如需要，处理新的嵌套对象
3. 通过检查字段存在性来保持向后兼容

## 客户端实现示例

### JavaScript/TypeScript

#### 版本检测

```typescript
interface ApiClient {
  version: string;
  baseUrl: string;
}

class LinkeApiClient implements ApiClient {
  constructor(
    public version: string = 'v2',
    public baseUrl: string = 'https://api.linke.com'
  ) {}

  private getVersionedUrl(endpoint: string): string {
    return `${this.baseUrl}/api/${this.version}${endpoint}`;
  }

  async getUser(userId: string): Promise<User> {
    const response = await fetch(this.getVersionedUrl(`/users/${userId}`));
    const data = await response.json();
    
    if (this.version === 'v1') {
      return this.transformUserV1ToV2(data);
    }
    return data;
  }

  private transformUserV1ToV2(v1User: any): User {
    return {
      user_id: v1User.id,
      full_name: v1User.name,
      email: v1User.email,
      created_at: null, // v1 中不可用
      updated_at: null, // v1 中不可用
      profile: {
        avatar_url: null,
        bio: null
      }
    };
  }
}
```

#### 渐进式迁移

```typescript
class MigrationAwareClient {
  private v1Client: LinkeApiClient;
  private v2Client: LinkeApiClient;

  constructor() {
    this.v1Client = new LinkeApiClient('v1');
    this.v2Client = new LinkeApiClient('v2');
  }

  async getUser(userId: string): Promise<User> {
    try {
      // 首先尝试 v2
      return await this.v2Client.getUser(userId);
    } catch (error) {
      console.warn('v2 失败，回退到 v1:', error);
      // 回退到 v1
      return await this.v1Client.getUser(userId);
    }
  }
}
```

### Python

#### 版本处理

```python
import requests
from typing import Dict, Any, Optional
from dataclasses import dataclass
from datetime import datetime

@dataclass
class User:
    user_id: str
    full_name: str
    email: str
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    profile: Optional[Dict[str, Any]] = None

class LinkeApiClient:
    def __init__(self, version: str = 'v2', base_url: str = 'https://api.linke.com'):
        self.version = version
        self.base_url = base_url

    def _get_versioned_url(self, endpoint: str) -> str:
        return f"{self.base_url}/api/{self.version}{endpoint}"

    def get_user(self, user_id: str) -> User:
        response = requests.get(self._get_versioned_url(f"/users/{user_id}"))
        response.raise_for_status()
        data = response.json()

        if self.version == 'v1':
            return self._transform_user_v1_to_v2(data)
        return User(**data)

    def _transform_user_v1_to_v2(self, v1_user: Dict[str, Any]) -> User:
        return User(
            user_id=v1_user['id'],
            full_name=v1_user['name'],
            email=v1_user['email'],
            created_at=None,
            updated_at=None,
            profile=None
        )
```

#### 响应头部处理

```python
class VersionAwareClient:
    def __init__(self):
        self.session = requests.Session()

    def make_request(self, endpoint: str, version: str = 'v2') -> Dict[str, Any]:
        url = f"https://api.linke.com/api/{version}{endpoint}"
        response = self.session.get(url)
        
        # 检查弃用警告
        if 'Warning' in response.headers:
            print(f"API 警告: {response.headers['Warning']}")
        
        if 'Sunset' in response.headers:
            print(f"API 下线日期: {response.headers['Sunset']}")
        
        if 'X-API-Sunset-Days' in response.headers:
            days = response.headers['X-API-Sunset-Days']
            print(f"API 将在 {days} 天后下线")
        
        response.raise_for_status()
        return response.json()
```

### Go

#### 版本处理

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type User struct {
    UserID    string    `json:"user_id"`
    FullName  string    `json:"full_name"`
    Email     string    `json:"email"`
    CreatedAt *time.Time `json:"created_at,omitempty"`
    UpdatedAt *time.Time `json:"updated_at,omitempty"`
    Profile   *Profile   `json:"profile,omitempty"`
}

type Profile struct {
    AvatarURL string `json:"avatar_url"`
    Bio       string `json:"bio"`
}

type ApiClient struct {
    BaseURL string
    Version string
    Client  *http.Client
}

func NewApiClient(version string) *ApiClient {
    return &ApiClient{
        BaseURL: "https://api.linke.com",
        Version: version,
        Client:  &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *ApiClient) GetUser(userID string) (*User, error) {
    url := fmt.Sprintf("%s/api/%s/users/%s", c.BaseURL, c.Version, userID)
    
    resp, err := c.Client.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, err
    }

    return &user, nil
}
```

## 测试迁移

### 测试策略

1. **并行测试**: 对两个版本进行测试
2. **响应比较**: 比较 v1 和 v2 响应
3. **性能测试**: 测量迁移开销
4. **错误处理**: 测试错误场景

### 示例测试用例

```typescript
describe('API 迁移测试', () => {
  const v1Client = new LinkeApiClient('v1');
  const v2Client = new LinkeApiClient('v2');

  test('用户数据迁移', async () => {
    const v1User = await v1Client.getUser('123');
    const v2User = await v2Client.getUser('123');

    // 核心数据应该匹配
    expect(v1User.id).toBe(v2User.user_id);
    expect(v1User.name).toBe(v2User.full_name);
    expect(v1User.email).toBe(v2User.email);

    // v2 应该有额外字段
    expect(v2User.created_at).toBeDefined();
    expect(v2User.profile).toBeDefined();
  });

  test('错误处理一致性', async () => {
    // 测试错误响应的一致性
    await expect(v1Client.getUser('invalid')).rejects.toThrow();
    await expect(v2Client.getUser('invalid')).rejects.toThrow();
  });
});
```

## 回退策略

如果在迁移过程中出现问题：

### 立即回退

1. **客户端**: 切换回 v1 端点
2. **更新版本**: 将版本参数更改为 'v1'
3. **监控**: 检查错误率和功能

### 渐进式回退

```typescript
class GradualRollbackClient {
  private useV2 = true;

  async getUser(userId: string): Promise<User> {
    if (this.useV2) {
      try {
        return await this.v2Client.getUser(userId);
      } catch (error) {
        console.error('v2 错误，回退到 v1:', error);
        this.useV2 = false; // 在此会话中禁用 v2
        return await this.v1Client.getUser(userId);
      }
    }
    return await this.v1Client.getUser(userId);
  }
}
```

## 监控和告警

### 客户端监控

```typescript
class MonitoredApiClient {
  async makeRequest(endpoint: string, version: string = 'v2'): Promise<any> {
    const startTime = Date.now();
    
    try {
      const response = await fetch(`/api/${version}${endpoint}`);
      const duration = Date.now() - startTime;
      
      // Monitor deprecation warnings
      if (response.headers.get('Warning')) {
        this.reportDeprecation(version, endpoint, response.headers.get('Warning'));
      }
      
      // Monitor performance
      this.reportPerformance(version, endpoint, duration);
      
      return await response.json();
    } catch (error) {
      this.reportError(version, endpoint, error);
      throw error;
    }
  }

  private reportDeprecation(version: string, endpoint: string, warning: string) {
    console.warn(`${version}${endpoint} 的弃用警告: ${warning}`);
    // 发送到监控服务
  }

  private reportPerformance(version: string, endpoint: string, duration: number) {
    // 发送指标到监控服务
    console.log(`API 调用 ${version}${endpoint} 耗时 ${duration}ms`);
  }

  private reportError(version: string, endpoint: string, error: any) {
    // 发送错误到监控服务
    console.error(`${version}${endpoint} 的 API 错误:`, error);
  }
}
```

## 最佳实践

### 1. 规划您的迁移

- 彻底阅读版本更新日志
- 首先在测试环境中测试
- 规划渐进式发布
- 准备好回退策略

### 2. 优雅地处理弃用

- 监控弃用头部
- 规划迁移时间表
- 在下线日期之前更新代码
- 与您的团队沟通

### 3. 版本固定

```typescript
// 好的做法：固定到特定版本
const client = new LinkeApiClient('v2');

// 坏的做法：自动使用最新版本
const client = new LinkeApiClient(); // 当 v3 发布时可能会损坏
```

### 4. 错误处理

```typescript
async function robustApiCall<T>(
  operation: () => Promise<T>,
  fallback?: () => Promise<T>
): Promise<T> {
  try {
    return await operation();
  } catch (error) {
    if (fallback && error.status >= 500) {
      console.warn('API 错误，尝试回退:', error);
      return await fallback();
    }
    throw error;
  }
}
```

### 5. 测试

- 在过渡期间测试两个版本
- 验证迁移后的数据完整性
- 使用真实数据进行性能测试
- 测试错误场景

## 支持和资源

### 文档

- [API 参考](/docs/api-reference)
- [版本更新日志](/docs/changelog)
- [错误代码](/docs/error-codes)

### 支持渠道

- GitHub Issues: [github.com/yourorg/linke/issues]
- Discord: [您的 Discord 频道]
- 邮箱: api-support@yourorg.com

### 迁移协助

针对复杂迁移或问题：

1. 创建详细问题，包含：
   - 当前版本
   - 目标版本
   - 受影响的特定端点
   - 错误消息
   - 期望与实际行为

2. 包含：
   - 客户端实现语言
   - 请求/响应示例
   - 迁移时间表限制

## 常见问题

### 问：我可以同时使用多个版本吗？

答：是的，您可以为不同的端点调用不同的版本，但要确保应用程序中的数据一致性。

### 问：旧版本支持多长时间？

答：已弃用的版本在弃用公告后至少支持12个月。请查看版本信息中的下线日期。

### 问：我的数据会自动迁移吗？

答：如果启用了自动迁移，API 可以自动迁移响应数据，但这可能影响性能。建议为生产环境规划显式迁移。

### 问：如果我在下线前没有迁移会怎样？

答：下线的版本会返回 HTTP 410 Gone。您的应用程序将停止使用这些端点。

### 问：我可以在 v2 中请求特定字段吗？

答：是的，v2 支持通过查询参数进行字段选择。详情请参阅 API 文档。

### 问：如何知道我目前使用的是哪个版本？

答：检查响应中的 `X-API-Version` 头部，或调用 `GET /api/version` 获取详细的版本信息。