package constants

// 服务器可见性状态
const (
	ServerShowVisible = 1
	ServerShowHidden  = 0
)

// 节点类型
const (
	NodeTypeShadowsocks = "shadowsocks"
)

// UniProxy 配置
const (
	DefaultPushInterval = 60
	DefaultPullInterval = 60
)

// 安全限制
const (
	MinServerTokenLength = 20
)

// 表名常量
const (
	TableServerGroups       = "server_groups"
	TableShadowsocksServers = "shadowsocks_servers"
)

// 字段长度限制
const (
	MaxNameLength         = 255
	MaxHostLength         = 255
	MaxTagsLength         = 255
	MaxRouteIDLength      = 255
	MaxObfsLength         = 11
	MaxObfsSettingsLength = 255
	MaxIPsLength          = 255
	MaxExcludesLength     = 500 // TEXT 类型字段的合理长度限制
)

// 端口范围
const (
	MinPort = 1
	MaxPort = 65535
)

// 服务器组ID特殊值
const (
	ServerGroupAllAccess = 0 // 表示访问所有服务器组
)

// 状态消息
const (
	MsgServerNotFound      = "服务器未找到"
	MsgServerGroupNotFound = "服务器组未找到"
	MsgInvalidNodeType     = "无效的节点类型"
	MsgTokenTooShort       = "令牌长度太短"
)
