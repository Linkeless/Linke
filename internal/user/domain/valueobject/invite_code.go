package valueobject

import (
	"errors"
	"strings"
)

// InviteCode represents an invite code and its metadata
type InviteCode struct {
	id   uint
	code string
}

// NewInviteCode creates a new InviteCode
func NewInviteCode(id uint, code string) (InviteCode, error) {
	code = strings.TrimSpace(code)
	
	if id == 0 {
		return InviteCode{}, errors.New("invite code ID cannot be zero")
	}
	
	if code == "" {
		return InviteCode{}, errors.New("invite code cannot be empty")
	}
	
	return InviteCode{
		id:   id,
		code: code,
	}, nil
}

// NewEmptyInviteCode creates an empty invite code (no invite used)
func NewEmptyInviteCode() InviteCode {
	return InviteCode{}
}

// ID returns the invite code ID
func (i InviteCode) ID() uint {
	return i.id
}

// Code returns the invite code string
func (i InviteCode) Code() string {
	return i.code
}

// IsEmpty checks if the invite code is empty (not set)
func (i InviteCode) IsEmpty() bool {
	return i.id == 0 && i.code == ""
}

// HasInvite checks if the user used an invite code
func (i InviteCode) HasInvite() bool {
	return !i.IsEmpty()
}

// String returns the string representation of the invite code
func (i InviteCode) String() string {
	if i.IsEmpty() {
		return ""
	}
	return i.code
}

// Equals checks if two invite codes are equal
func (i InviteCode) Equals(other InviteCode) bool {
	return i.id == other.id && i.code == other.code
}