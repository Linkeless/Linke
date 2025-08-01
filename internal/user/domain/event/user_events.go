package event

import (
	"time"

	"linke/internal/user/domain/valueobject"
)


// BaseDomainEvent provides common functionality for domain events
type BaseDomainEvent struct {
	eventID     string
	aggregateID string
	eventType   string
	occurredAt  time.Time
	eventData   interface{}
}

// NewBaseDomainEvent creates a new base domain event
func NewBaseDomainEvent(aggregateID, eventType string, eventData interface{}) BaseDomainEvent {
	return BaseDomainEvent{
		eventID:     valueobject.NewUserID().String(), // Using UserID generator for event IDs
		aggregateID: aggregateID,
		eventType:   eventType,
		occurredAt:  time.Now(),
		eventData:   eventData,
	}
}

// EventID returns the event ID
func (e BaseDomainEvent) EventID() string {
	return e.eventID
}

// AggregateID returns the aggregate ID
func (e BaseDomainEvent) AggregateID() string {
	return e.aggregateID
}

// EventType returns the event type
func (e BaseDomainEvent) EventType() string {
	return e.eventType
}

// OccurredAt returns when the event occurred
func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// EventData returns the event data
func (e BaseDomainEvent) EventData() interface{} {
	return e.eventData
}

// UserCreated event
type UserCreated struct {
	BaseDomainEvent
	UserID   valueobject.UserID
	Email    valueobject.Email
	Provider valueobject.Provider
}

// NewUserCreated creates a new UserCreated event
func NewUserCreated(userID valueobject.UserID, email valueobject.Email, provider valueobject.Provider) UserCreated {
	eventData := map[string]interface{}{
		"user_id":  userID.String(),
		"email":    email.String(),
		"provider": provider.String(),
	}

	return UserCreated{
		BaseDomainEvent: NewBaseDomainEvent(userID.String(), "UserCreated", eventData),
		UserID:          userID,
		Email:           email,
		Provider:        provider,
	}
}

// UserLoggedIn event
type UserLoggedIn struct {
	BaseDomainEvent
	UserID    valueobject.UserID
	Provider  valueobject.Provider
	IPAddress string
	UserAgent string
}

// NewUserLoggedIn creates a new UserLoggedIn event
func NewUserLoggedIn(userID valueobject.UserID, provider valueobject.Provider, ipAddress, userAgent string) UserLoggedIn {
	eventData := map[string]interface{}{
		"user_id":    userID.String(),
		"provider":   provider.String(),
		"ip_address": ipAddress,
		"user_agent": userAgent,
	}

	return UserLoggedIn{
		BaseDomainEvent: NewBaseDomainEvent(userID.String(), "UserLoggedIn", eventData),
		UserID:          userID,
		Provider:        provider,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
	}
}

// UserLoginFailed event
type UserLoginFailed struct {
	BaseDomainEvent
	Email     valueobject.Email
	Reason    string
	IPAddress string
	UserAgent string
}

// NewUserLoginFailed creates a new UserLoginFailed event
func NewUserLoginFailed(email valueobject.Email, reason, ipAddress, userAgent string) UserLoginFailed {
	eventData := map[string]interface{}{
		"email":      email.String(),
		"reason":     reason,
		"ip_address": ipAddress,
		"user_agent": userAgent,
	}

	return UserLoginFailed{
		BaseDomainEvent: NewBaseDomainEvent(email.String(), "UserLoginFailed", eventData),
		Email:           email,
		Reason:          reason,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
	}
}

// UserPasswordChanged event
type UserPasswordChanged struct {
	BaseDomainEvent
	UserID valueobject.UserID
}

// NewUserPasswordChanged creates a new UserPasswordChanged event
func NewUserPasswordChanged(userID valueobject.UserID) UserPasswordChanged {
	eventData := map[string]interface{}{
		"user_id": userID.String(),
	}

	return UserPasswordChanged{
		BaseDomainEvent: NewBaseDomainEvent(userID.String(), "UserPasswordChanged", eventData),
		UserID:          userID,
	}
}

// UserStatusChanged event
type UserStatusChanged struct {
	BaseDomainEvent
	UserID    valueobject.UserID
	OldStatus valueobject.UserStatus
	NewStatus valueobject.UserStatus
}

// NewUserStatusChanged creates a new UserStatusChanged event
func NewUserStatusChanged(userID valueobject.UserID, oldStatus, newStatus valueobject.UserStatus) UserStatusChanged {
	eventData := map[string]interface{}{
		"user_id":    userID.String(),
		"old_status": oldStatus.String(),
		"new_status": newStatus.String(),
	}

	return UserStatusChanged{
		BaseDomainEvent: NewBaseDomainEvent(userID.String(), "UserStatusChanged", eventData),
		UserID:          userID,
		OldStatus:       oldStatus,
		NewStatus:       newStatus,
	}
}

// UserRoleChanged event
type UserRoleChanged struct {
	BaseDomainEvent
	UserID  valueobject.UserID
	OldRole valueobject.UserRole
	NewRole valueobject.UserRole
}

// NewUserRoleChanged creates a new UserRoleChanged event
func NewUserRoleChanged(userID valueobject.UserID, oldRole, newRole valueobject.UserRole) UserRoleChanged {
	eventData := map[string]interface{}{
		"user_id":  userID.String(),
		"old_role": oldRole.String(),
		"new_role": newRole.String(),
	}

	return UserRoleChanged{
		BaseDomainEvent: NewBaseDomainEvent(userID.String(), "UserRoleChanged", eventData),
		UserID:          userID,
		OldRole:         oldRole,
		NewRole:         newRole,
	}
}

// UserProfileUpdated event
type UserProfileUpdated struct {
	BaseDomainEvent
	UserID valueobject.UserID
	Fields []string
}

// NewUserProfileUpdated creates a new UserProfileUpdated event
func NewUserProfileUpdated(userID valueobject.UserID, fields []string) UserProfileUpdated {
	eventData := map[string]interface{}{
		"user_id": userID.String(),
		"fields":  fields,
	}

	return UserProfileUpdated{
		BaseDomainEvent: NewBaseDomainEvent(userID.String(), "UserProfileUpdated", eventData),
		UserID:          userID,
		Fields:          fields,
	}
}