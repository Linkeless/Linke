package persistence

import (
	"time"

	"linke/internal/server/domain/model"
	"linke/internal/server/domain/valueobject"
)

// ServerGroupPO represents the persistent object for server groups
type ServerGroupPO struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:255;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"not null;index"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName returns the table name for ServerGroupPO
func (ServerGroupPO) TableName() string {
	return "server_groups"
}

// ToDomain converts ServerGroupPO to domain model
func (po *ServerGroupPO) ToDomain() (*model.ServerGroup, error) {
	id, err := valueobject.NewServerGroupID(po.ID)
	if err != nil {
		return nil, err
	}
	
	name, err := valueobject.NewServerGroupName(po.Name)
	if err != nil {
		return nil, err
	}
	
	return model.ReconstructServerGroup(id, name, po.CreatedAt, po.UpdatedAt), nil
}

// FromDomain converts domain model to ServerGroupPO
func (po *ServerGroupPO) FromDomain(group *model.ServerGroup) {
	if !group.ID().IsZero() {
		po.ID = group.ID().Value()
	}
	po.Name = group.Name().Value()
	po.CreatedAt = group.CreatedAt()
	po.UpdatedAt = group.UpdatedAt()
}