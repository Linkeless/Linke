# Telegram Bot 增强版功能文档

## 概述

增强版 Telegram Bot 提供了更友好的用户交互体验，支持多种非命令操作方式，让用户能够通过点击按钮、自然语言输入等方式轻松管理订阅。

## 主要特性

### 1. 交互式菜单系统

#### 主菜单
- 📊 **我的订阅** - 查看和管理订阅
- 💎 **套餐商店** - 浏览和购买套餐
- 📈 **使用统计** - 查看流量使用情况
- ⚙️ **设置** - 管理账号设置
- 💬 **客服支持** - 获取帮助和支持
- ❓ **帮助** - 查看使用指南

#### 订阅管理菜单
- 📋 查看详情 - 显示完整订阅信息
- 📈 流量使用 - 详细流量统计
- 🔄 续费设置 - 管理自动续费
- ⏸️ 暂停订阅 - 临时暂停服务
- ⬆️ 升级套餐 - 升级到更高级套餐
- 📝 订阅历史 - 查看历史记录

### 2. 内联键盘（Inline Keyboard）

所有菜单都使用内联键盘实现，用户只需点击按钮即可：
- 无需记住命令
- 支持多级菜单导航
- 包含返回按钮，方便导航
- 动态生成按钮内容

### 3. 回复键盘（Reply Keyboard）

提供快捷操作键盘：
```
[📊 查看订阅] [📈 流量使用]
[💎 套餐价格] [💬 联系客服]
     [🏠 主菜单]
```

### 4. 自然语言理解

Bot 能够理解常用关键词，自动触发相应功能：
- "订阅"、"套餐" → 显示订阅菜单
- "流量"、"使用" → 显示使用统计
- "帮助"、"help" → 显示帮助信息
- "价格"、"购买" → 显示套餐商店

### 5. 视觉化数据展示

#### 流量使用进度条
```
已使用: 50.00 GB
总流量: 100.00 GB
剩余: 50.00 GB
█████░░░░░ 50.0%
```

#### 订阅状态图标
- ✅ 活跃
- ⏸️ 暂停
- ❌ 已取消
- ⏰ 已过期
- 🎁 试用中

### 6. 智能提醒

根据使用情况提供智能提醒：
- 流量使用超过 90%：⚠️ 流量即将用尽，建议升级套餐
- 流量使用超过 70%：💡 流量使用较多，请注意控制
- 订阅即将到期：提醒续费或升级

## 使用方法

### 1. 启动 Bot

发送 `/start` 或点击开始按钮，Bot 会显示主菜单。

### 2. 绑定账号

首次使用需要绑定账号：
1. 访问网站
2. 使用 Telegram 登录
3. 授权后自动完成绑定

### 3. 日常使用

#### 方式一：点击按钮
直接点击菜单中的按钮进行操作

#### 方式二：输入关键词
输入"订阅"、"流量"等关键词，Bot 自动识别意图

#### 方式三：使用快捷键盘
使用底部的快捷键盘快速访问常用功能

## 技术实现

### 架构设计

```go
type BotEnhanced struct {
    api                     *tgbotapi.BotAPI
    userService            userInterfaces.UserService
    subscriptionService    interfaces.UserSubscriptionService
    subscriptionPlanService interfaces.SubscriptionPlanService
    cfg                    *config.Config
    userStates             map[int64]string // 用户会话状态
}
```

### 核心功能模块

1. **更新处理器** (`handleUpdate`)
   - 处理回调查询（按钮点击）
   - 处理命令消息
   - 处理自然语言消息

2. **菜单系统**
   - 主菜单 (`showMainMenu`)
   - 订阅菜单 (`showSubscriptionMenu`)
   - 套餐菜单 (`showPlansMenu`)
   - 设置菜单 (`showSettingsMenu`)

3. **数据展示**
   - 订阅详情 (`showSubscriptionInfo`)
   - 使用统计 (`showUsageDetails`)
   - 套餐详情 (`showPlanDetails`)

4. **辅助功能**
   - 格式化函数（流量、状态、周期等）
   - 进度条生成
   - 错误处理

### 配置要求

```env
# Telegram Bot Token
TELEGRAM_BOT_TOKEN=your_bot_token_here
```

## 部署说明

### 1. 创建 Telegram Bot

1. 在 Telegram 中找到 @BotFather
2. 发送 `/newbot` 创建新机器人
3. 设置机器人名称和用户名
4. 获取 Bot Token

### 2. 配置 Bot

在 `.env` 文件中设置：
```env
TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz
```

### 3. 启动服务

```bash
make dev
```

Bot 会自动启动并开始监听消息。

## 未来改进

### 计划中的功能

1. **多语言支持**
   - 中文/英文切换
   - 自动检测用户语言

2. **支付集成**
   - 直接在 Bot 中完成支付
   - 支持多种支付方式

3. **通知系统**
   - 流量告警
   - 订阅到期提醒
   - 优惠活动通知

4. **高级统计**
   - 历史使用趋势图
   - 使用预测
   - 节点使用分析

5. **客服系统**
   - 工单创建和跟踪
   - 实时客服对话
   - FAQ 自动回复

## 常见问题

### Q: 如何重新显示主菜单？
A: 发送 `/menu` 命令或点击任意菜单中的"返回主菜单"按钮

### Q: Bot 无响应怎么办？
A: 
1. 检查 Bot Token 是否正确
2. 确认服务是否正常运行
3. 查看日志文件排查错误

### Q: 如何解绑账号？
A: 目前需要联系客服进行解绑操作

### Q: 支持群组使用吗？
A: 目前仅支持私聊使用，群组功能正在开发中

## 安全注意事项

1. **保护 Bot Token**
   - 不要将 Token 提交到代码仓库
   - 使用环境变量存储
   - 定期更换 Token

2. **用户验证**
   - 所有操作都需要验证用户身份
   - 使用 Telegram ID 进行用户关联
   - 敏感操作需要二次确认

3. **数据保护**
   - 不在消息中显示敏感信息
   - 使用 HTTPS 进行 API 通信
   - 定期清理会话数据

## 技术支持

如有问题或建议，请通过以下方式联系：
- Telegram: @your_support_bot
- Email: support@your-domain.com
- GitHub Issues: https://github.com/your-repo/issues