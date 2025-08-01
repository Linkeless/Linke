package persistence

import (
	"time"

	"linke/internal/server/domain/model"
	"linke/internal/server/domain/valueobject"
)

// ShadowsocksServerPO represents the persistent object for shadowsocks servers
type ShadowsocksServerPO struct {
	ID           int     `gorm:"primaryKey;autoIncrement"`
	GroupID      uint    `gorm:"column:group_id;not null;index"`
	RouteID      string  `gorm:"column:route_id;type:varchar(255)"`
	ParentID     *int    `gorm:"column:parent_id;type:int"`
	Tags         string  `gorm:"column:tags;type:varchar(255)"`
	Excludes     string  `gorm:"column:excludes;type:text"`
	IPs          string  `gorm:"column:ips;type:varchar(255)"`
	Name         string  `gorm:"column:name;type:varchar(255);not null"`
	Rate         float64 `gorm:"column:rate;type:decimal(10,2);not null"`
	Host         string  `gorm:"column:host;type:varchar(255);not null"`
	Port         int     `gorm:"column:port;type:int;not null"`
	ServerPort   int     `gorm:"column:server_port;type:int;not null"`
	Cipher       string  `gorm:"column:cipher;type:varchar(255);not null"`
	Obfs         string  `gorm:"column:obfs;type:char(11)"`
	ObfsSettings string  `gorm:"column:obfs_settings;type:varchar(255)"`
	Show         int     `gorm:"column:show;type:tinyint;not null;default:0"`
	Sort         *int    `gorm:"column:sort;type:int"`
	CreatedAt    int     `gorm:"column:created_at;type:int;not null"`
	UpdatedAt    int     `gorm:"column:updated_at;type:int;not null"`
}

// TableName returns the table name for ShadowsocksServerPO
func (ShadowsocksServerPO) TableName() string {
	return "shadowsocks_servers"
}

// ToDomain converts ShadowsocksServerPO to domain model
func (po *ShadowsocksServerPO) ToDomain() (*model.ShadowsocksServer, error) {
	id, err := valueobject.NewServerID(po.ID)
	if err != nil {
		return nil, err
	}
	
	groupID, err := valueobject.NewServerGroupID(po.GroupID)
	if err != nil {
		return nil, err
	}
	
	name, err := valueobject.NewServerName(po.Name)
	if err != nil {
		return nil, err
	}
	
	host, err := valueobject.NewServerHost(po.Host)
	if err != nil {
		return nil, err
	}
	
	port, err := valueobject.NewServerPort(po.Port)
	if err != nil {
		return nil, err
	}
	
	serverPort, err := valueobject.NewServerPort(po.ServerPort)
	if err != nil {
		return nil, err
	}
	
	cipher, err := valueobject.NewCipher(po.Cipher)
	if err != nil {
		return nil, err
	}
	
	rate, err := valueobject.NewRate(po.Rate)
	if err != nil {
		return nil, err
	}
	
	isVisible := po.Show == 1
	createdAt := time.Unix(int64(po.CreatedAt), 0)
	updatedAt := time.Unix(int64(po.UpdatedAt), 0)
	
	return model.ReconstructShadowsocksServer(
		id, groupID, name, host, port, serverPort, cipher, rate, isVisible,
		po.Tags, po.Obfs, po.ObfsSettings, po.Excludes, po.IPs, po.RouteID,
		po.ParentID, po.Sort, createdAt, updatedAt,
	), nil
}

// FromDomain converts domain model to ShadowsocksServerPO
func (po *ShadowsocksServerPO) FromDomain(server *model.ShadowsocksServer) {
	if !server.ID().IsZero() {
		po.ID = server.ID().Value()
	}
	
	po.GroupID = server.GroupID().Value()
	po.Name = server.Name().Value()
	po.Host = server.Host().Value()
	po.Port = server.Port().Value()
	po.ServerPort = server.ServerPort().Value()
	po.Cipher = server.Cipher().Value()
	po.Rate = server.Rate().Value()
	po.Tags = server.Tags()
	po.Obfs = server.Obfs()
	po.ObfsSettings = server.ObfsSettings()
	po.Excludes = server.Excludes()
	po.IPs = server.IPs()
	po.RouteID = server.RouteID()
	po.ParentID = server.ParentID()
	po.Sort = server.Sort()
	
	if server.IsVisible() {
		po.Show = 1
	} else {
		po.Show = 0
	}
	
	po.CreatedAt = int(server.CreatedAt().Unix())
	po.UpdatedAt = int(server.UpdatedAt().Unix())
}