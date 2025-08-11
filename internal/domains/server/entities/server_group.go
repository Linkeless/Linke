package entities

import (
	"time"

	"linke/internal/domains/server/constants"
)

// ServerGroup represents a server group
type ServerGroup struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"size:255;not null;uniqueIndex" binding:"required" validate:"max=255"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`

	// Reverse relationship - shadowsocks servers in this group
	ShadowsocksServers []ShadowsocksServer `json:"shadowsocks_servers,omitempty" gorm:"foreignKey:GroupID;references:ID"`
}

// TableName returns the table name for ServerGroup model
func (ServerGroup) TableName() string {
	return constants.TableServerGroups
}

// ServerGroupResponse represents the server group data structure for API responses
type ServerGroupResponse struct {
	ID        uint   `json:"id" example:"1"`
	Name      string `json:"name" example:"Asia Pacific"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ToResponse converts ServerGroup to ServerGroupResponse
func (sg *ServerGroup) ToResponse() *ServerGroupResponse {
	return &ServerGroupResponse{
		ID:        sg.ID,
		Name:      sg.Name,
		CreatedAt: sg.CreatedAt.Format(time.RFC3339),
		UpdatedAt: sg.UpdatedAt.Format(time.RFC3339),
	}
}
