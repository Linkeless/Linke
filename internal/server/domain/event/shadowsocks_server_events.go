package event

import (
	"time"

	"linke/internal/server/domain/valueobject"
)

// ShadowsocksServerCreated is emitted when a shadowsocks server is created
type ShadowsocksServerCreated struct {
	BaseDomainEvent
	ServerID valueobject.ServerID
	GroupID  valueobject.ServerGroupID
	Name     valueobject.ServerName
	Host     valueobject.ServerHost
	Port     valueobject.ServerPort
}

// NewShadowsocksServerCreated creates a new ShadowsocksServerCreated event
func NewShadowsocksServerCreated(
	serverID valueobject.ServerID,
	groupID valueobject.ServerGroupID,
	name valueobject.ServerName,
	host valueobject.ServerHost,
	port valueobject.ServerPort,
) *ShadowsocksServerCreated {
	return &ShadowsocksServerCreated{
		BaseDomainEvent: BaseDomainEvent{
			occurredOn: time.Now(),
			eventType:  "ShadowsocksServerCreated",
		},
		ServerID: serverID,
		GroupID:  groupID,
		Name:     name,
		Host:     host,
		Port:     port,
	}
}

// AggregateID returns the aggregate ID
func (e ShadowsocksServerCreated) AggregateID() string {
	return e.ServerID.String()
}

// ShadowsocksServerUpdated is emitted when a shadowsocks server is updated
type ShadowsocksServerUpdated struct {
	BaseDomainEvent
	ServerID valueobject.ServerID
	GroupID  valueobject.ServerGroupID
	Name     valueobject.ServerName
	Host     valueobject.ServerHost
	Port     valueobject.ServerPort
}

// NewShadowsocksServerUpdated creates a new ShadowsocksServerUpdated event
func NewShadowsocksServerUpdated(
	serverID valueobject.ServerID,
	groupID valueobject.ServerGroupID,
	name valueobject.ServerName,
	host valueobject.ServerHost,
	port valueobject.ServerPort,
) *ShadowsocksServerUpdated {
	return &ShadowsocksServerUpdated{
		BaseDomainEvent: BaseDomainEvent{
			occurredOn: time.Now(),
			eventType:  "ShadowsocksServerUpdated",
		},
		ServerID: serverID,
		GroupID:  groupID,
		Name:     name,
		Host:     host,
		Port:     port,
	}
}

// AggregateID returns the aggregate ID
func (e ShadowsocksServerUpdated) AggregateID() string {
	return e.ServerID.String()
}

// ShadowsocksServerDeleted is emitted when a shadowsocks server is deleted
type ShadowsocksServerDeleted struct {
	BaseDomainEvent
	ServerID valueobject.ServerID
}

// NewShadowsocksServerDeleted creates a new ShadowsocksServerDeleted event
func NewShadowsocksServerDeleted(serverID valueobject.ServerID) *ShadowsocksServerDeleted {
	return &ShadowsocksServerDeleted{
		BaseDomainEvent: BaseDomainEvent{
			occurredOn: time.Now(),
			eventType:  "ShadowsocksServerDeleted",
		},
		ServerID: serverID,
	}
}

// AggregateID returns the aggregate ID
func (e ShadowsocksServerDeleted) AggregateID() string {
	return e.ServerID.String()
}

// ShadowsocksServerStatusChanged is emitted when a shadowsocks server's visibility status changes
type ShadowsocksServerStatusChanged struct {
	BaseDomainEvent
	ServerID  valueobject.ServerID
	IsVisible bool
}

// NewShadowsocksServerStatusChanged creates a new ShadowsocksServerStatusChanged event
func NewShadowsocksServerStatusChanged(serverID valueobject.ServerID, isVisible bool) *ShadowsocksServerStatusChanged {
	return &ShadowsocksServerStatusChanged{
		BaseDomainEvent: BaseDomainEvent{
			occurredOn: time.Now(),
			eventType:  "ShadowsocksServerStatusChanged",
		},
		ServerID:  serverID,
		IsVisible: isVisible,
	}
}

// AggregateID returns the aggregate ID
func (e ShadowsocksServerStatusChanged) AggregateID() string {
	return e.ServerID.String()
}