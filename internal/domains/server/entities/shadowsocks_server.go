package entities

// ShadowsocksServer represents a Shadowsocks proxy server node
// Updated to use proper foreign key relationship with ServerGroup
type ShadowsocksServer struct {
	ID           int     `json:"id" gorm:"primaryKey;autoIncrement"`
	GroupID      uint    `json:"group_id" gorm:"column:group_id;not null;index"`
	RouteID      string  `json:"route_id,omitempty" gorm:"column:route_id;type:varchar(255)"`
	ParentID     *int    `json:"parent_id,omitempty" gorm:"column:parent_id;type:int"`
	Tags         string  `json:"tags,omitempty" gorm:"column:tags;type:varchar(255)"`
	Excludes     string  `json:"excludes,omitempty" gorm:"column:excludes;type:text"`
	IPs          string  `json:"ips,omitempty" gorm:"column:ips;type:varchar(255)"`
	Name         string  `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Rate         float64 `json:"rate" gorm:"column:rate;type:decimal(10,2);not null"`
	Host         string  `json:"host" gorm:"column:host;type:varchar(255);not null"`
	Port         int     `json:"port" gorm:"column:port;type:int;not null"`
	ServerPort   int     `json:"server_port" gorm:"column:server_port;type:int;not null"`
	Cipher       string  `json:"cipher" gorm:"column:cipher;type:varchar(255);not null"`
	Obfs         string  `json:"obfs,omitempty" gorm:"column:obfs;type:char(11)"`
	ObfsSettings string  `json:"obfs_settings,omitempty" gorm:"column:obfs_settings;type:varchar(255)"`
	Show         int     `json:"show" gorm:"column:show;type:tinyint;not null;default:0"`
	Sort         *int    `json:"sort,omitempty" gorm:"column:sort;type:int"`
	CreatedAt    int     `json:"created_at" gorm:"column:created_at;type:int;not null"`
	UpdatedAt    int     `json:"updated_at" gorm:"column:updated_at;type:int;not null"`

	// Relationship
	ServerGroup *ServerGroup `json:"server_group,omitempty" gorm:"foreignKey:GroupID;references:ID"`
}

// TableName returns the table name for ShadowsocksServer model
func (ShadowsocksServer) TableName() string {
	return "shadowsocks_servers"
}

// IsVisible checks if the shadowsocks server should be shown to users
func (s *ShadowsocksServer) IsVisible() bool {
	return s.Show == 1
}

// ShadowsocksServerResponse represents the shadowsocks server data structure for API responses
type ShadowsocksServerResponse struct {
	ID           int     `json:"id" example:"1"`
	GroupID      uint    `json:"group_id" example:"1"`
	RouteID      string  `json:"route_id,omitempty" example:"route-1"`
	ParentID     *int    `json:"parent_id,omitempty" example:"1"`
	Name         string  `json:"name" example:"US-01"`
	Tags         string  `json:"tags,omitempty" example:"premium,fast"`
	Host         string  `json:"host" example:"us01.example.com"`
	Port         int     `json:"port" example:"443"`
	ServerPort   int     `json:"server_port" example:"8388"`
	Cipher       string  `json:"cipher" example:"aes-256-gcm"`
	Obfs         string  `json:"obfs,omitempty" example:"tls"`
	ObfsSettings string  `json:"obfs_settings,omitempty" example:"obfs=tls"`
	Excludes     string  `json:"excludes,omitempty" example:"192.168.0.0/16"`
	IPs          string  `json:"ips,omitempty" example:"0.0.0.0/0"`
	Rate         float64 `json:"rate" example:"1.0"`
	Show         int     `json:"show" example:"1"`
	Sort         *int    `json:"sort,omitempty" example:"1"`
	CreatedAt    int     `json:"created_at" example:"1640995200"`
	UpdatedAt    int     `json:"updated_at" example:"1640995200"`

	// Relationship data
	ServerGroup *ServerGroupResponse `json:"server_group,omitempty"`
}

// ToResponse converts ShadowsocksServer to ShadowsocksServerResponse
func (s *ShadowsocksServer) ToResponse() *ShadowsocksServerResponse {
	resp := &ShadowsocksServerResponse{
		ID:           s.ID,
		GroupID:      s.GroupID,
		RouteID:      s.RouteID,
		ParentID:     s.ParentID,
		Name:         s.Name,
		Tags:         s.Tags,
		Host:         s.Host,
		Port:         s.Port,
		ServerPort:   s.ServerPort,
		Cipher:       s.Cipher,
		Obfs:         s.Obfs,
		ObfsSettings: s.ObfsSettings,
		Excludes:     s.Excludes,
		IPs:          s.IPs,
		Rate:         s.Rate,
		Show:         s.Show,
		Sort:         s.Sort,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}

	// Include server group data if loaded
	if s.ServerGroup != nil {
		resp.ServerGroup = s.ServerGroup.ToResponse()
	}

	return resp
}

// ToPublicResponse converts ShadowsocksServer to a public response (hides sensitive info)
func (s *ShadowsocksServer) ToPublicResponse() *ShadowsocksServerResponse {
	return &ShadowsocksServerResponse{
		ID:         s.ID,
		GroupID:    s.GroupID,
		Name:       s.Name,
		Tags:       s.Tags,
		Host:       s.Host,
		Port:       s.Port,
		ServerPort: s.ServerPort,
		Cipher:     s.Cipher,
		Rate:       s.Rate,
		Show:       s.Show,
		Sort:       s.Sort,
	}
}
