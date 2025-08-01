package model

import (
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/ticket/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// MessageType represents the type of a ticket message
type MessageType string

const (
	MessageTypeUser   MessageType = "user"
	MessageTypeAdmin  MessageType = "admin"
	MessageTypeSystem MessageType = "system"
)

// Attachment represents a file attachment
type Attachment struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size,omitempty"`
	Type string `json:"type,omitempty"`
}

// TicketMessage represents a message within a ticket
type TicketMessage struct {
	id          uint
	ticketID    valueobject.TicketID
	userID      sharedvo.UserID
	content     string
	messageType MessageType
	attachments []Attachment
	isInternal  bool
	metadata    map[string]interface{}
	createdAt   time.Time
	updatedAt   time.Time
}

// NewTicketMessage creates a new ticket message
func NewTicketMessage(
	id uint,
	ticketID valueobject.TicketID,
	userID sharedvo.UserID,
	content string,
	messageType MessageType,
	isInternal bool,
) (*TicketMessage, error) {
	if content == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}
	
	if ticketID.IsZero() {
		return nil, fmt.Errorf("ticket ID cannot be zero")
	}
	
	if userID.IsZero() {
		return nil, fmt.Errorf("user ID cannot be zero")
	}
	
	// Default to user message type if not specified
	if messageType == "" {
		messageType = MessageTypeUser
	}
	
	now := time.Now()
	
	return &TicketMessage{
		id:          id,
		ticketID:    ticketID,
		userID:      userID,
		content:     content,
		messageType: messageType,
		attachments: make([]Attachment, 0),
		isInternal:  isInternal,
		metadata:    make(map[string]interface{}),
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// ID returns the message ID
func (tm *TicketMessage) ID() uint {
	return tm.id
}

// TicketID returns the ticket ID
func (tm *TicketMessage) TicketID() valueobject.TicketID {
	return tm.ticketID
}

// UserID returns the user ID
func (tm *TicketMessage) UserID() sharedvo.UserID {
	return tm.userID
}

// Content returns the message content
func (tm *TicketMessage) Content() string {
	return tm.content
}

// MessageType returns the message type
func (tm *TicketMessage) MessageType() MessageType {
	return tm.messageType
}

// Attachments returns the attachments
func (tm *TicketMessage) Attachments() []Attachment {
	return tm.attachments
}

// IsInternal returns whether the message is internal
func (tm *TicketMessage) IsInternal() bool {
	return tm.isInternal
}

// Metadata returns the metadata
func (tm *TicketMessage) Metadata() map[string]interface{} {
	return tm.metadata
}

// CreatedAt returns the creation time
func (tm *TicketMessage) CreatedAt() time.Time {
	return tm.createdAt
}

// UpdatedAt returns the last update time
func (tm *TicketMessage) UpdatedAt() time.Time {
	return tm.updatedAt
}

// UpdateContent updates the message content
func (tm *TicketMessage) UpdateContent(content string) error {
	if content == "" {
		return fmt.Errorf("message content cannot be empty")
	}
	
	tm.content = content
	tm.updatedAt = time.Now()
	return nil
}

// AddAttachment adds an attachment to the message
func (tm *TicketMessage) AddAttachment(attachment Attachment) error {
	if attachment.Name == "" {
		return fmt.Errorf("attachment name cannot be empty")
	}
	
	if attachment.URL == "" {
		return fmt.Errorf("attachment URL cannot be empty")
	}
	
	tm.attachments = append(tm.attachments, attachment)
	tm.updatedAt = time.Now()
	return nil
}

// RemoveAttachment removes an attachment by name
func (tm *TicketMessage) RemoveAttachment(name string) {
	for i, attachment := range tm.attachments {
		if attachment.Name == name {
			tm.attachments = append(tm.attachments[:i], tm.attachments[i+1:]...)
			tm.updatedAt = time.Now()
			break
		}
	}
}

// SetAttachments sets all attachments
func (tm *TicketMessage) SetAttachments(attachments []Attachment) {
	tm.attachments = make([]Attachment, len(attachments))
	copy(tm.attachments, attachments)
	tm.updatedAt = time.Now()
}

// ToggleInternal toggles the internal flag
func (tm *TicketMessage) ToggleInternal() {
	tm.isInternal = !tm.isInternal
	tm.updatedAt = time.Now()
}

// SetInternal sets the internal flag
func (tm *TicketMessage) SetInternal(isInternal bool) {
	tm.isInternal = isInternal
	tm.updatedAt = time.Now()
}

// SetMetadata sets a metadata value
func (tm *TicketMessage) SetMetadata(key string, value interface{}) {
	tm.metadata[key] = value
	tm.updatedAt = time.Now()
}

// GetMetadata gets a metadata value
func (tm *TicketMessage) GetMetadata(key string) (interface{}, bool) {
	value, exists := tm.metadata[key]
	return value, exists
}

// IsFromUser checks if message is from a user
func (tm *TicketMessage) IsFromUser() bool {
	return tm.messageType == MessageTypeUser
}

// IsFromAdmin checks if message is from an admin
func (tm *TicketMessage) IsFromAdmin() bool {
	return tm.messageType == MessageTypeAdmin
}

// IsFromSystem checks if message is from system
func (tm *TicketMessage) IsFromSystem() bool {
	return tm.messageType == MessageTypeSystem
}

// HasAttachments checks if message has attachments
func (tm *TicketMessage) HasAttachments() bool {
	return len(tm.attachments) > 0
}

// AttachmentsToJSON converts attachments to JSON string
func (tm *TicketMessage) AttachmentsToJSON() string {
	if len(tm.attachments) == 0 {
		return ""
	}
	
	data, _ := json.Marshal(tm.attachments)
	return string(data)
}

// MetadataToJSON converts metadata to JSON string
func (tm *TicketMessage) MetadataToJSON() string {
	if len(tm.metadata) == 0 {
		return ""
	}
	
	data, _ := json.Marshal(tm.metadata)
	return string(data)
}