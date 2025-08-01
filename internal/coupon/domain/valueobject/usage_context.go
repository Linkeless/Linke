package valueobject

import (
	"fmt"
	"time"
	
	sharedvo "linke/internal/shared/valueobject"
)

// UsageContext represents the context in which a coupon is used
type UsageContext struct {
	userID    sharedvo.UserID
	orderID   OrderID
	sessionID SessionID
	channel   UsageChannel
	timestamp time.Time
	metadata  map[string]interface{}
}

// OrderID represents an order identifier
type OrderID struct {
	value uint64
}

// NewOrderID creates a new order ID
func NewOrderID(value uint64) (OrderID, error) {
	if value == 0 {
		return OrderID{}, fmt.Errorf("order ID cannot be zero")
	}
	return OrderID{value: value}, nil
}

// Value returns the underlying value
func (oid OrderID) Value() uint64 {
	return oid.value
}

// String returns string representation
func (oid OrderID) String() string {
	return fmt.Sprintf("%d", oid.value)
}

// IsZero checks if the ID is zero
func (oid OrderID) IsZero() bool {
	return oid.value == 0
}

// SessionID represents a user session identifier
type SessionID struct {
	value string
}

// NewSessionID creates a new session ID
func NewSessionID(value string) (SessionID, error) {
	if value == "" {
		return SessionID{}, fmt.Errorf("session ID cannot be empty")
	}
	return SessionID{value: value}, nil
}

// Value returns the underlying value
func (sid SessionID) Value() string {
	return sid.value
}

// String returns string representation
func (sid SessionID) String() string {
	return sid.value
}

// IsEmpty checks if the ID is empty
func (sid SessionID) IsEmpty() bool {
	return sid.value == ""
}

// UsageChannel represents the channel through which a coupon is used
type UsageChannel string

const (
	UsageChannelWeb     UsageChannel = "web"
	UsageChannelMobile  UsageChannel = "mobile"
	UsageChannelAPI     UsageChannel = "api"
	UsageChannelAdmin   UsageChannel = "admin"
)

// NewUsageChannel creates a new usage channel
func NewUsageChannel(value string) (UsageChannel, error) {
	channel := UsageChannel(value)
	switch channel {
	case UsageChannelWeb, UsageChannelMobile, UsageChannelAPI, UsageChannelAdmin:
		return channel, nil
	default:
		return "", fmt.Errorf("invalid usage channel: %s", value)
	}
}

// String returns string representation
func (uc UsageChannel) String() string {
	return string(uc)
}

// IsValid checks if the channel is valid
func (uc UsageChannel) IsValid() bool {
	switch uc {
	case UsageChannelWeb, UsageChannelMobile, UsageChannelAPI, UsageChannelAdmin:
		return true
	default:
		return false
	}
}

// NewUsageContext creates a new usage context
func NewUsageContext(
	userID sharedvo.UserID,
	orderID OrderID,
	sessionID SessionID,
	channel UsageChannel,
) (UsageContext, error) {
	if userID.IsZero() {
		return UsageContext{}, fmt.Errorf("user ID cannot be zero")
	}
	
	if orderID.IsZero() {
		return UsageContext{}, fmt.Errorf("order ID cannot be zero")
	}
	
	if sessionID.IsEmpty() {
		return UsageContext{}, fmt.Errorf("session ID cannot be empty")
	}
	
	if !channel.IsValid() {
		return UsageContext{}, fmt.Errorf("invalid usage channel: %s", channel)
	}
	
	return UsageContext{
		userID:    userID,
		orderID:   orderID,
		sessionID: sessionID,
		channel:   channel,
		timestamp: time.Now(),
		metadata:  make(map[string]interface{}),
	}, nil
}

// UserID returns the user ID
func (uc UsageContext) UserID() sharedvo.UserID {
	return uc.userID
}

// OrderID returns the order ID
func (uc UsageContext) OrderID() OrderID {
	return uc.orderID
}

// SessionID returns the session ID
func (uc UsageContext) SessionID() SessionID {
	return uc.sessionID
}

// Channel returns the usage channel
func (uc UsageContext) Channel() UsageChannel {
	return uc.channel
}

// Timestamp returns when the context was created
func (uc UsageContext) Timestamp() time.Time {
	return uc.timestamp
}

// AddMetadata adds metadata to the context
func (uc *UsageContext) AddMetadata(key string, value interface{}) {
	if uc.metadata == nil {
		uc.metadata = make(map[string]interface{})
	}
	uc.metadata[key] = value
}

// GetMetadata retrieves metadata from the context
func (uc UsageContext) GetMetadata(key string) (interface{}, bool) {
	if uc.metadata == nil {
		return nil, false
	}
	value, exists := uc.metadata[key]
	return value, exists
}

// String returns string representation
func (uc UsageContext) String() string {
	return fmt.Sprintf("UsageContext{UserID: %s, OrderID: %s, Channel: %s, Timestamp: %s}",
		uc.userID.String(), uc.orderID.String(), uc.channel.String(), uc.timestamp.Format(time.RFC3339))
}