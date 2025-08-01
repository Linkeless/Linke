package command

// Server Group Commands

// CreateServerGroupCommand represents a command to create a server group
type CreateServerGroupCommand struct {
	Name string `json:"name" validate:"required,max=255"`
}

// UpdateServerGroupCommand represents a command to update a server group
type UpdateServerGroupCommand struct {
	ID   uint    `json:"id" validate:"required"`
	Name *string `json:"name,omitempty" validate:"omitempty,max=255"`
}

// DeleteServerGroupCommand represents a command to delete a server group
type DeleteServerGroupCommand struct {
	ID uint `json:"id" validate:"required"`
}

// Shadowsocks Server Commands

// CreateShadowsocksServerCommand represents a command to create a shadowsocks server
type CreateShadowsocksServerCommand struct {
	GroupID      uint    `json:"group_id" validate:"required"`
	RouteID      string  `json:"route_id,omitempty" validate:"max=255"`
	ParentID     *int    `json:"parent_id,omitempty"`
	Name         string  `json:"name" validate:"required,max=255"`
	Tags         string  `json:"tags,omitempty" validate:"max=255"`
	Host         string  `json:"host" validate:"required,max=255"`
	Port         int     `json:"port" validate:"required,min=1,max=65535"`
	ServerPort   int     `json:"server_port" validate:"required,min=1,max=65535"`
	Cipher       string  `json:"cipher" validate:"required,max=255"`
	Obfs         string  `json:"obfs,omitempty" validate:"max=11"`
	ObfsSettings string  `json:"obfs_settings,omitempty" validate:"max=255"`
	Excludes     string  `json:"excludes,omitempty"`
	IPs          string  `json:"ips,omitempty" validate:"max=255"`
	Rate         float64 `json:"rate" validate:"required,min=0.1"`
	Show         int     `json:"show" validate:"min=0,max=1"`
	Sort         *int    `json:"sort,omitempty"`
}

// UpdateShadowsocksServerCommand represents a command to update a shadowsocks server
type UpdateShadowsocksServerCommand struct {
	ID           int      `json:"id" validate:"required"`
	GroupID      *uint    `json:"group_id,omitempty"`
	RouteID      *string  `json:"route_id,omitempty" validate:"omitempty,max=255"`
	ParentID     *int     `json:"parent_id,omitempty"`
	Name         *string  `json:"name,omitempty" validate:"omitempty,max=255"`
	Tags         *string  `json:"tags,omitempty" validate:"omitempty,max=255"`
	Host         *string  `json:"host,omitempty" validate:"omitempty,max=255"`
	Port         *int     `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
	ServerPort   *int     `json:"server_port,omitempty" validate:"omitempty,min=1,max=65535"`
	Cipher       *string  `json:"cipher,omitempty" validate:"omitempty,max=255"`
	Obfs         *string  `json:"obfs,omitempty" validate:"omitempty,max=11"`
	ObfsSettings *string  `json:"obfs_settings,omitempty" validate:"omitempty,max=255"`
	Excludes     *string  `json:"excludes,omitempty"`
	IPs          *string  `json:"ips,omitempty" validate:"omitempty,max=255"`
	Rate         *float64 `json:"rate,omitempty" validate:"omitempty,min=0.1"`
	Show         *int     `json:"show,omitempty" validate:"omitempty,min=0,max=1"`
	Sort         *int     `json:"sort,omitempty"`
}

// DeleteShadowsocksServerCommand represents a command to delete a shadowsocks server
type DeleteShadowsocksServerCommand struct {
	ID int `json:"id" validate:"required"`
}

// ChangeShadowsocksServerVisibilityCommand represents a command to change server visibility
type ChangeShadowsocksServerVisibilityCommand struct {
	ID        int  `json:"id" validate:"required"`
	IsVisible bool `json:"is_visible"`
}

// MoveShadowsocksServerToGroupCommand represents a command to move a server to a different group
type MoveShadowsocksServerToGroupCommand struct {
	ID      int  `json:"id" validate:"required"`
	GroupID uint `json:"group_id" validate:"required"`
}