package model

import (
	"time"

	"linke/internal/server/domain/event"
	"linke/internal/server/domain/valueobject"
)

// ShadowsocksServer represents a shadowsocks server aggregate root
type ShadowsocksServer struct {
	id           valueobject.ServerID
	groupID      valueobject.ServerGroupID
	name         valueobject.ServerName
	host         valueobject.ServerHost
	port         valueobject.ServerPort
	serverPort   valueobject.ServerPort
	cipher       valueobject.Cipher
	rate         valueobject.Rate
	isVisible    bool
	tags         string
	obfs         string
	obfsSettings string
	excludes     string
	ips          string
	routeID      string
	parentID     *int
	sort         *int
	createdAt    time.Time
	updatedAt    time.Time
	events       []event.DomainEvent
}

// NewShadowsocksServer creates a new shadowsocks server
func NewShadowsocksServer(
	groupID valueobject.ServerGroupID,
	name valueobject.ServerName,
	host valueobject.ServerHost,
	port valueobject.ServerPort,
	serverPort valueobject.ServerPort,
	cipher valueobject.Cipher,
	rate valueobject.Rate,
) (*ShadowsocksServer, error) {
	server := &ShadowsocksServer{
		groupID:    groupID,
		name:       name,
		host:       host,
		port:       port,
		serverPort: serverPort,
		cipher:     cipher,
		rate:       rate,
		isVisible:  false, // Default to hidden
		createdAt:  time.Now(),
		updatedAt:  time.Now(),
		events:     make([]event.DomainEvent, 0),
	}
	
	return server, nil
}

// ReconstructShadowsocksServer reconstructs a shadowsocks server from persistence
func ReconstructShadowsocksServer(
	id valueobject.ServerID,
	groupID valueobject.ServerGroupID,
	name valueobject.ServerName,
	host valueobject.ServerHost,
	port valueobject.ServerPort,
	serverPort valueobject.ServerPort,
	cipher valueobject.Cipher,
	rate valueobject.Rate,
	isVisible bool,
	tags, obfs, obfsSettings, excludes, ips, routeID string,
	parentID, sort *int,
	createdAt, updatedAt time.Time,
) *ShadowsocksServer {
	return &ShadowsocksServer{
		id:           id,
		groupID:      groupID,
		name:         name,
		host:         host,
		port:         port,
		serverPort:   serverPort,
		cipher:       cipher,
		rate:         rate,
		isVisible:    isVisible,
		tags:         tags,
		obfs:         obfs,
		obfsSettings: obfsSettings,
		excludes:     excludes,
		ips:          ips,
		routeID:      routeID,
		parentID:     parentID,
		sort:         sort,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		events:       make([]event.DomainEvent, 0),
	}
}

// ID returns the server ID
func (ss *ShadowsocksServer) ID() valueobject.ServerID {
	return ss.id
}

// GroupID returns the group ID
func (ss *ShadowsocksServer) GroupID() valueobject.ServerGroupID {
	return ss.groupID
}

// Name returns the server name
func (ss *ShadowsocksServer) Name() valueobject.ServerName {
	return ss.name
}

// Host returns the server host
func (ss *ShadowsocksServer) Host() valueobject.ServerHost {
	return ss.host
}

// Port returns the server port
func (ss *ShadowsocksServer) Port() valueobject.ServerPort {
	return ss.port
}

// ServerPort returns the server port
func (ss *ShadowsocksServer) ServerPort() valueobject.ServerPort {
	return ss.serverPort
}

// Cipher returns the cipher
func (ss *ShadowsocksServer) Cipher() valueobject.Cipher {
	return ss.cipher
}

// Rate returns the rate
func (ss *ShadowsocksServer) Rate() valueobject.Rate {
	return ss.rate
}

// IsVisible returns whether the server is visible
func (ss *ShadowsocksServer) IsVisible() bool {
	return ss.isVisible
}

// Tags returns the tags
func (ss *ShadowsocksServer) Tags() string {
	return ss.tags
}

// Obfs returns the obfs
func (ss *ShadowsocksServer) Obfs() string {
	return ss.obfs
}

// ObfsSettings returns the obfs settings
func (ss *ShadowsocksServer) ObfsSettings() string {
	return ss.obfsSettings
}

// Excludes returns the excludes
func (ss *ShadowsocksServer) Excludes() string {
	return ss.excludes
}

// IPs returns the IPs
func (ss *ShadowsocksServer) IPs() string {
	return ss.ips
}

// RouteID returns the route ID
func (ss *ShadowsocksServer) RouteID() string {
	return ss.routeID
}

// ParentID returns the parent ID
func (ss *ShadowsocksServer) ParentID() *int {
	return ss.parentID
}

// Sort returns the sort order
func (ss *ShadowsocksServer) Sort() *int {
	return ss.sort
}

// CreatedAt returns when the server was created
func (ss *ShadowsocksServer) CreatedAt() time.Time {
	return ss.createdAt
}

// UpdatedAt returns when the server was last updated
func (ss *ShadowsocksServer) UpdatedAt() time.Time {
	return ss.updatedAt
}

// UpdateConfiguration updates server configuration
func (ss *ShadowsocksServer) UpdateConfiguration(
	name valueobject.ServerName,
	host valueobject.ServerHost,
	port valueobject.ServerPort,
	serverPort valueobject.ServerPort,
	cipher valueobject.Cipher,
	rate valueobject.Rate,
) error {
	ss.name = name
	ss.host = host
	ss.port = port
	ss.serverPort = serverPort
	ss.cipher = cipher
	ss.rate = rate
	ss.updatedAt = time.Now()
	
	// Only raise event if ID is set (persisted aggregate)
	if !ss.id.IsZero() {
		ss.raiseEvent(event.NewShadowsocksServerUpdated(ss.id, ss.groupID, ss.name, ss.host, ss.port))
	}
	
	return nil
}

// ChangeVisibility changes the server visibility
func (ss *ShadowsocksServer) ChangeVisibility(isVisible bool) {
	if ss.isVisible == isVisible {
		return // No change required
	}
	
	ss.isVisible = isVisible
	ss.updatedAt = time.Now()
	
	// Only raise event if ID is set (persisted aggregate)
	if !ss.id.IsZero() {
		ss.raiseEvent(event.NewShadowsocksServerStatusChanged(ss.id, ss.isVisible))
	}
}

// MoveToGroup moves the server to a different group
func (ss *ShadowsocksServer) MoveToGroup(newGroupID valueobject.ServerGroupID) {
	if ss.groupID.Equals(newGroupID) {
		return // No change required
	}
	
	ss.groupID = newGroupID
	ss.updatedAt = time.Now()
	
	// Only raise event if ID is set (persisted aggregate)
	if !ss.id.IsZero() {
		ss.raiseEvent(event.NewShadowsocksServerUpdated(ss.id, ss.groupID, ss.name, ss.host, ss.port))
	}
}

// UpdateMetadata updates server metadata
func (ss *ShadowsocksServer) UpdateMetadata(tags, obfs, obfsSettings, excludes, ips, routeID string, parentID, sort *int) {
	ss.tags = tags
	ss.obfs = obfs
	ss.obfsSettings = obfsSettings
	ss.excludes = excludes
	ss.ips = ips
	ss.routeID = routeID
	ss.parentID = parentID
	ss.sort = sort
	ss.updatedAt = time.Now()
}

// MarkAsCreated marks the server as created (called after persistence)
func (ss *ShadowsocksServer) MarkAsCreated(id valueobject.ServerID) {
	ss.id = id
	ss.raiseEvent(event.NewShadowsocksServerCreated(ss.id, ss.groupID, ss.name, ss.host, ss.port))
}

// MarkAsDeleted marks the server as deleted
func (ss *ShadowsocksServer) MarkAsDeleted() {
	if !ss.id.IsZero() {
		ss.raiseEvent(event.NewShadowsocksServerDeleted(ss.id))
	}
}

// DomainEvents returns all domain events
func (ss *ShadowsocksServer) DomainEvents() []event.DomainEvent {
	return ss.events
}

// ClearEvents clears all domain events
func (ss *ShadowsocksServer) ClearEvents() {
	ss.events = make([]event.DomainEvent, 0)
}

// raiseEvent adds a domain event
func (ss *ShadowsocksServer) raiseEvent(evt event.DomainEvent) {
	ss.events = append(ss.events, evt)
}