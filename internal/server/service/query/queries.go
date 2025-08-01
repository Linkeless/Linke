package query

// Server Group Queries

// GetServerGroupQuery represents a query to get a server group by ID
type GetServerGroupQuery struct {
	ID uint `json:"id" validate:"required"`
}

// GetServerGroupsQuery represents a query to get server groups with pagination
type GetServerGroupsQuery struct {
	Limit  int `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset int `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// GetAllServerGroupsQuery represents a query to get all server groups
type GetAllServerGroupsQuery struct{}

// Shadowsocks Server Queries

// GetShadowsocksServerQuery represents a query to get a shadowsocks server by ID
type GetShadowsocksServerQuery struct {
	ID int `json:"id" validate:"required"`
}

// GetShadowsocksServersQuery represents a query to get shadowsocks servers with filters
type GetShadowsocksServersQuery struct {
	GroupID *uint  `json:"group_id,omitempty"`
	Show    *int   `json:"show,omitempty"`
	Tags    string `json:"tags,omitempty"`
	Cipher  string `json:"cipher,omitempty"`
	Limit   int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset  int    `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// GetShadowsocksServersByGroupQuery represents a query to get shadowsocks servers by group
type GetShadowsocksServersByGroupQuery struct {
	GroupID     uint `json:"group_id" validate:"required"`
	VisibleOnly bool `json:"visible_only"`
}

// GetVisibleShadowsocksServersQuery represents a query to get all visible shadowsocks servers
type GetVisibleShadowsocksServersQuery struct{}

// GetShadowsocksServerCountQuery represents a query to get shadowsocks server count
type GetShadowsocksServerCountQuery struct {
	GroupID *uint `json:"group_id,omitempty"`
}