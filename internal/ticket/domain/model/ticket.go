package model

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	sharedvo "linke/internal/shared/valueobject"
	"linke/internal/ticket/domain/event"
	"linke/internal/ticket/domain/valueobject"
)

// Ticket represents the ticket aggregate root
type Ticket struct {
	// Identity
	id           valueobject.TicketID
	ticketNumber valueobject.TicketNumber
	
	// Basic Information
	title       string
	description string
	category    valueobject.TicketCategory
	priority    valueobject.TicketPriority
	status      valueobject.TicketStatus
	
	// User Information
	userID sharedvo.UserID
	
	// Assignment Information
	assignedToID *sharedvo.UserID
	assignedAt   *time.Time
	
	// Resolution Information
	resolvedByID *sharedvo.UserID
	resolvedAt   *time.Time
	resolution   string
	
	// Timing Information
	firstResponseAt *time.Time
	lastResponseAt  *time.Time
	closedAt        *time.Time
	
	// Metadata
	tags     []string
	metadata map[string]interface{}
	
	// Messages (entities within the aggregate)
	messages []TicketMessage
	
	// Domain Events
	domainEvents []event.DomainEvent
	
	// Timestamps
	createdAt time.Time
	updatedAt time.Time
	version   int
}

// NewTicket creates a new ticket aggregate
func NewTicket(
	id valueobject.TicketID,
	userID sharedvo.UserID,
	title, description string,
	category valueobject.TicketCategory,
	priority valueobject.TicketPriority,
) (*Ticket, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, fmt.Errorf("ticket ID cannot be zero")
	}
	
	if userID.IsZero() {
		return nil, fmt.Errorf("user ID cannot be zero")
	}
	
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("ticket title cannot be empty")
	}
	
	if strings.TrimSpace(description) == "" {
		return nil, fmt.Errorf("ticket description cannot be empty")
	}
	
	if len(title) < 5 || len(title) > 255 {
		return nil, fmt.Errorf("ticket title must be between 5 and 255 characters")
	}
	
	if len(description) < 10 || len(description) > 5000 {
		return nil, fmt.Errorf("ticket description must be between 10 and 5000 characters")
	}
	
	// Generate ticket number
	ticketNumber := generateTicketNumber()
	
	now := time.Now()
	
	ticket := &Ticket{
		id:           id,
		ticketNumber: ticketNumber,
		title:        strings.TrimSpace(title),
		description:  strings.TrimSpace(description),
		category:     category,
		priority:     priority,
		status:       valueobject.DefaultTicketStatus(),
		userID:       userID,
		tags:         make([]string, 0),
		metadata:     make(map[string]interface{}),
		messages:     make([]TicketMessage, 0),
		domainEvents: make([]event.DomainEvent, 0),
		createdAt:    now,
		updatedAt:    now,
		version:      1,
	}
	
	// Raise domain event
	ticket.raiseEvent(event.NewTicketCreated(
		id, ticketNumber, userID, title, description, category, priority,
	))
	
	return ticket, nil
}

// ID returns the ticket ID
func (t *Ticket) ID() valueobject.TicketID {
	return t.id
}

// TicketNumber returns the ticket number
func (t *Ticket) TicketNumber() valueobject.TicketNumber {
	return t.ticketNumber
}

// Title returns the ticket title
func (t *Ticket) Title() string {
	return t.title
}

// Description returns the ticket description
func (t *Ticket) Description() string {
	return t.description
}

// Category returns the ticket category
func (t *Ticket) Category() valueobject.TicketCategory {
	return t.category
}

// Priority returns the ticket priority
func (t *Ticket) Priority() valueobject.TicketPriority {
	return t.priority
}

// Status returns the ticket status
func (t *Ticket) Status() valueobject.TicketStatus {
	return t.status
}

// UserID returns the user ID
func (t *Ticket) UserID() sharedvo.UserID {
	return t.userID
}

// AssignedToID returns the assigned user ID
func (t *Ticket) AssignedToID() *sharedvo.UserID {
	return t.assignedToID
}

// AssignedAt returns the assignment time
func (t *Ticket) AssignedAt() *time.Time {
	return t.assignedAt
}

// ResolvedByID returns the resolver user ID
func (t *Ticket) ResolvedByID() *sharedvo.UserID {
	return t.resolvedByID
}

// ResolvedAt returns the resolution time
func (t *Ticket) ResolvedAt() *time.Time {
	return t.resolvedAt
}

// Resolution returns the resolution text
func (t *Ticket) Resolution() string {
	return t.resolution
}

// FirstResponseAt returns the first response time
func (t *Ticket) FirstResponseAt() *time.Time {
	return t.firstResponseAt
}

// LastResponseAt returns the last response time
func (t *Ticket) LastResponseAt() *time.Time {
	return t.lastResponseAt
}

// ClosedAt returns the closure time
func (t *Ticket) ClosedAt() *time.Time {
	return t.closedAt
}

// Tags returns the tags
func (t *Ticket) Tags() []string {
	return t.tags
}

// Metadata returns the metadata
func (t *Ticket) Metadata() map[string]interface{} {
	return t.metadata
}

// Messages returns the messages
func (t *Ticket) Messages() []TicketMessage {
	return t.messages
}

// CreatedAt returns the creation time
func (t *Ticket) CreatedAt() time.Time {
	return t.createdAt
}

// UpdatedAt returns the last update time
func (t *Ticket) UpdatedAt() time.Time {
	return t.updatedAt
}

// Version returns the aggregate version
func (t *Ticket) Version() int {
	return t.version
}

// DomainEvents returns the domain events
func (t *Ticket) DomainEvents() []event.DomainEvent {
	return t.domainEvents
}

// ClearDomainEvents clears the domain events
func (t *Ticket) ClearDomainEvents() {
	t.domainEvents = make([]event.DomainEvent, 0)
}

// UpdateTitle updates the ticket title
func (t *Ticket) UpdateTitle(title string) error {
	title = strings.TrimSpace(title)
	
	if title == "" {
		return fmt.Errorf("ticket title cannot be empty")
	}
	
	if len(title) < 5 || len(title) > 255 {
		return fmt.Errorf("ticket title must be between 5 and 255 characters")
	}
	
	t.title = title
	t.markAsUpdated()
	return nil
}

// UpdateDescription updates the ticket description
func (t *Ticket) UpdateDescription(description string) error {
	description = strings.TrimSpace(description)
	
	if description == "" {
		return fmt.Errorf("ticket description cannot be empty")
	}
	
	if len(description) < 10 || len(description) > 5000 {
		return fmt.Errorf("ticket description must be between 10 and 5000 characters")
	}
	
	t.description = description
	t.markAsUpdated()
	return nil
}

// UpdateCategory updates the ticket category
func (t *Ticket) UpdateCategory(category valueobject.TicketCategory) error {
	t.category = category
	t.markAsUpdated()
	return nil
}

// UpdatePriority updates the ticket priority
func (t *Ticket) UpdatePriority(priority valueobject.TicketPriority) error {
	t.priority = priority
	t.markAsUpdated()
	return nil
}

// ChangeStatus changes the ticket status
func (t *Ticket) ChangeStatus(newStatus valueobject.TicketStatus, changedBy sharedvo.UserID) error {
	if !t.status.CanTransitionTo(newStatus) {
		return fmt.Errorf("cannot transition from %s to %s", t.status.String(), newStatus.String())
	}
	
	oldStatus := t.status
	t.status = newStatus
	
	// Update timing fields based on status
	now := time.Now()
	switch {
	case newStatus.IsResolved():
		t.resolvedAt = &now
		t.resolvedByID = &changedBy
	case newStatus.IsClosed():
		t.closedAt = &now
	}
	
	t.markAsUpdated()
	
	// Raise domain event
	t.raiseEvent(event.NewTicketStatusChanged(t.id, oldStatus, newStatus, changedBy))
	
	return nil
}

// AssignTo assigns the ticket to a user
func (t *Ticket) AssignTo(assignedToID sharedvo.UserID) error {
	if assignedToID.IsZero() {
		return fmt.Errorf("assigned user ID cannot be zero")
	}
	
	now := time.Now()
	t.assignedToID = &assignedToID
	t.assignedAt = &now
	
	// If ticket is open, move to in progress
	if t.status.IsOpen() {
		inProgressStatus := valueobject.MustTicketStatus(valueobject.TicketStatusInProgress)
		t.status = inProgressStatus
	}
	
	t.markAsUpdated()
	
	// Raise domain event
	t.raiseEvent(event.NewTicketAssigned(t.id, assignedToID, now))
	
	return nil
}

// Resolve resolves the ticket
func (t *Ticket) Resolve(resolvedBy sharedvo.UserID, resolution string) error {
	if resolvedBy.IsZero() {
		return fmt.Errorf("resolver user ID cannot be zero")
	}
	
	resolution = strings.TrimSpace(resolution)
	if resolution == "" {
		return fmt.Errorf("resolution cannot be empty")
	}
	
	if len(resolution) < 10 || len(resolution) > 5000 {
		return fmt.Errorf("resolution must be between 10 and 5000 characters")
	}
	
	resolvedStatus := valueobject.MustTicketStatus(valueobject.TicketStatusResolved)
	if !t.status.CanTransitionTo(resolvedStatus) {
		return fmt.Errorf("cannot resolve ticket with status %s", t.status.String())
	}
	
	now := time.Now()
	t.status = resolvedStatus
	t.resolvedByID = &resolvedBy
	t.resolvedAt = &now
	t.resolution = resolution
	
	t.markAsUpdated()
	
	// Raise domain event
	t.raiseEvent(event.NewTicketResolved(t.id, resolvedBy, now, resolution))
	
	return nil
}

// Close closes the ticket
func (t *Ticket) Close() error {
	closedStatus := valueobject.MustTicketStatus(valueobject.TicketStatusClosed)
	if !t.status.CanTransitionTo(closedStatus) {
		return fmt.Errorf("cannot close ticket with status %s", t.status.String())
	}
	
	now := time.Now()
	t.status = closedStatus
	t.closedAt = &now
	
	t.markAsUpdated()
	
	// Raise domain event
	t.raiseEvent(event.NewTicketClosed(t.id, now))
	
	return nil
}

// AddMessage adds a message to the ticket
func (t *Ticket) AddMessage(message TicketMessage) error {
	// Validate message belongs to this ticket
	if !message.TicketID().Equals(t.id) {
		return fmt.Errorf("message does not belong to this ticket")
	}
	
	t.messages = append(t.messages, message)
	
	// Update last response time
	now := time.Now()
	t.lastResponseAt = &now
	
	// If this is the first admin response, update first response time
	if t.firstResponseAt == nil && message.IsFromAdmin() {
		t.firstResponseAt = &now
	}
	
	t.markAsUpdated()
	
	// Raise domain event
	t.raiseEvent(event.NewTicketMessageAdded(
		t.id,
		message.ID(),
		message.UserID(),
		message.Content(),
		string(message.MessageType()),
		message.IsInternal(),
	))
	
	return nil
}

// AddTag adds a tag to the ticket
func (t *Ticket) AddTag(tag string) error {
	tag = strings.TrimSpace(strings.ToLower(tag))
	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}
	
	// Check if tag already exists
	for _, existingTag := range t.tags {
		if existingTag == tag {
			return nil // Tag already exists, no error
		}
	}
	
	t.tags = append(t.tags, tag)
	t.markAsUpdated()
	return nil
}

// RemoveTag removes a tag from the ticket
func (t *Ticket) RemoveTag(tag string) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	for i, existingTag := range t.tags {
		if existingTag == tag {
			t.tags = append(t.tags[:i], t.tags[i+1:]...)
			t.markAsUpdated()
			break
		}
	}
}

// SetTags sets all tags
func (t *Ticket) SetTags(tags []string) {
	cleanTags := make([]string, 0, len(tags))
	seen := make(map[string]bool)
	
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag != "" && !seen[tag] {
			cleanTags = append(cleanTags, tag)
			seen[tag] = true
		}
	}
	
	t.tags = cleanTags
	t.markAsUpdated()
}

// SetMetadata sets a metadata value
func (t *Ticket) SetMetadata(key string, value interface{}) {
	t.metadata[key] = value
	t.markAsUpdated()
}

// GetMetadata gets a metadata value
func (t *Ticket) GetMetadata(key string) (interface{}, bool) {
	value, exists := t.metadata[key]
	return value, exists
}

// IsAssigned checks if ticket is assigned
func (t *Ticket) IsAssigned() bool {
	return t.assignedToID != nil && !t.assignedToID.IsZero()
}

// IsResolved checks if ticket is resolved
func (t *Ticket) IsResolved() bool {
	return t.status.IsResolved()
}

// IsClosed checks if ticket is closed
func (t *Ticket) IsClosed() bool {
	return t.status.IsClosed()
}

// IsActive checks if ticket is in active state
func (t *Ticket) IsActive() bool {
	return t.status.IsActive()
}

// HasMessages checks if ticket has messages
func (t *Ticket) HasMessages() bool {
	return len(t.messages) > 0
}

// HasFirstResponse checks if ticket has received first response
func (t *Ticket) HasFirstResponse() bool {
	return t.firstResponseAt != nil
}

// GetMessageCount returns the number of messages
func (t *Ticket) GetMessageCount() int {
	return len(t.messages)
}

// GetPublicMessages returns non-internal messages
func (t *Ticket) GetPublicMessages() []TicketMessage {
	publicMessages := make([]TicketMessage, 0)
	for _, message := range t.messages {
		if !message.IsInternal() {
			publicMessages = append(publicMessages, message)
		}
	}
	return publicMessages
}

// markAsUpdated updates the timestamp and version
func (t *Ticket) markAsUpdated() {
	t.updatedAt = time.Now()
	t.version++
}

// raiseEvent adds a domain event
func (t *Ticket) raiseEvent(domainEvent event.DomainEvent) {
	t.domainEvents = append(t.domainEvents, domainEvent)
}

// generateTicketNumber generates a unique ticket number
func generateTicketNumber() valueobject.TicketNumber {
	// Generate format: TK-YYYYMMDD-XXXXXX (e.g., TK-20240718-ABC123)
	now := time.Now()
	dateStr := now.Format("20060102")
	
	// Generate 6-character random string
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	
	ticketNumber := fmt.Sprintf("TK-%s-%s", dateStr, string(b))
	return valueobject.MustTicketNumber(ticketNumber)
}