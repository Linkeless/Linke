package errors

import "fmt"

// Ticket domain specific errors

// TicketNotFoundError represents a ticket not found error
type TicketNotFoundError struct {
	TicketID uint
}

func (e *TicketNotFoundError) Error() string {
	return fmt.Sprintf("ticket with ID %d not found", e.TicketID)
}

func (e *TicketNotFoundError) Is(target error) bool {
	_, ok := target.(*TicketNotFoundError)
	if ok {
		return true
	}
	return IsNotFound(target)
}

// TicketForbiddenError represents a forbidden ticket operation error
type TicketForbiddenError struct {
	TicketID uint
	UserID   uint
	Action   string
}

func (e *TicketForbiddenError) Error() string {
	return fmt.Sprintf("user %d is not allowed to %s ticket %d", e.UserID, e.Action, e.TicketID)
}

func (e *TicketForbiddenError) Is(target error) bool {
	_, ok := target.(*TicketForbiddenError)
	if ok {
		return true
	}
	return IsForbidden(target)
}

// TicketMessageNotFoundError represents a ticket message not found error
type TicketMessageNotFoundError struct {
	MessageID uint
}

func (e *TicketMessageNotFoundError) Error() string {
	return fmt.Sprintf("ticket message with ID %d not found", e.MessageID)
}

func (e *TicketMessageNotFoundError) Is(target error) bool {
	_, ok := target.(*TicketMessageNotFoundError)
	if ok {
		return true
	}
	return IsNotFound(target)
}

// UserNotFoundError represents a user not found error
type UserNotFoundError struct {
	UserID uint
}

func (e *UserNotFoundError) Error() string {
	return fmt.Sprintf("user with ID %d not found", e.UserID)
}

func (e *UserNotFoundError) Is(target error) bool {
	_, ok := target.(*UserNotFoundError)
	if ok {
		return true
	}
	return IsNotFound(target)
}

// InvalidTicketStateError represents an invalid ticket state transition error
type InvalidTicketStateError struct {
	TicketID    uint
	CurrentState string
	Action      string
	Reason      string
}

func (e *InvalidTicketStateError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("cannot %s ticket %d in state '%s': %s", e.Action, e.TicketID, e.CurrentState, e.Reason)
	}
	return fmt.Sprintf("cannot %s ticket %d in state '%s'", e.Action, e.TicketID, e.CurrentState)
}

func (e *InvalidTicketStateError) Is(target error) bool {
	_, ok := target.(*InvalidTicketStateError)
	if ok {
		return true
	}
	return IsInvalidInput(target)
}

// Factory functions for creating domain-specific errors

// NewTicketNotFound creates a new TicketNotFoundError
func NewTicketNotFound(ticketID uint) error {
	return &TicketNotFoundError{TicketID: ticketID}
}

// NewTicketForbidden creates a new TicketForbiddenError
func NewTicketForbidden(ticketID, userID uint, action string) error {
	return &TicketForbiddenError{
		TicketID: ticketID,
		UserID:   userID,
		Action:   action,
	}
}

// NewTicketMessageNotFound creates a new TicketMessageNotFoundError
func NewTicketMessageNotFound(messageID uint) error {
	return &TicketMessageNotFoundError{MessageID: messageID}
}

// NewUserNotFound creates a new UserNotFoundError
func NewUserNotFound(userID uint) error {
	return &UserNotFoundError{UserID: userID}
}

// NewInvalidTicketState creates a new InvalidTicketStateError
func NewInvalidTicketState(ticketID uint, currentState, action, reason string) error {
	return &InvalidTicketStateError{
		TicketID:     ticketID,
		CurrentState: currentState,
		Action:       action,
		Reason:       reason,
	}
}

// Domain-specific error checking functions

// IsTicketNotFound checks if an error is a ticket not found error
func IsTicketNotFound(err error) bool {
	var tErr *TicketNotFoundError
	return As(err, &tErr) || IsNotFound(err)
}

// IsTicketForbidden checks if an error is a ticket forbidden error
func IsTicketForbidden(err error) bool {
	var tErr *TicketForbiddenError
	return As(err, &tErr) || IsForbidden(err)
}

// IsTicketMessageNotFound checks if an error is a ticket message not found error
func IsTicketMessageNotFound(err error) bool {
	var tErr *TicketMessageNotFoundError
	return As(err, &tErr) || IsNotFound(err)
}

// IsUserNotFound checks if an error is a user not found error
func IsUserNotFound(err error) bool {
	var uErr *UserNotFoundError
	return As(err, &uErr) || IsNotFound(err)
}

// IsInvalidTicketState checks if an error is an invalid ticket state error
func IsInvalidTicketState(err error) bool {
	var tErr *InvalidTicketStateError
	return As(err, &tErr) || IsInvalidInput(err)
}

// Helper function for errors.As to avoid import confusion
func As(err error, target interface{}) bool {
	// Using a local implementation to avoid import cycle
	switch e := err.(type) {
	case *TicketNotFoundError:
		if t, ok := target.(**TicketNotFoundError); ok {
			*t = e
			return true
		}
	case *TicketForbiddenError:
		if t, ok := target.(**TicketForbiddenError); ok {
			*t = e
			return true
		}
	case *TicketMessageNotFoundError:
		if t, ok := target.(**TicketMessageNotFoundError); ok {
			*t = e
			return true
		}
	case *UserNotFoundError:
		if t, ok := target.(**UserNotFoundError); ok {
			*t = e
			return true
		}
	case *InvalidTicketStateError:
		if t, ok := target.(**InvalidTicketStateError); ok {
			*t = e
			return true
		}
	}
	return false
}