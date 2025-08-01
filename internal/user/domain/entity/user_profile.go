package entity

import (
	"time"
	
	"linke/internal/user/domain/valueobject"
)

// UserProfile represents the user's profile information
type UserProfile struct {
	// Profile fields
	username valueobject.Username
	name     valueobject.DisplayName
	avatar   valueobject.AvatarURL
	
	// Timestamps
	createdAt time.Time
	updatedAt time.Time
}

// NewUserProfile creates a new user profile
func NewUserProfile(username valueobject.Username, name valueobject.DisplayName, avatar valueobject.AvatarURL) *UserProfile {
	now := time.Now()
	return &UserProfile{
		username:  username,
		name:      name,
		avatar:    avatar,
		createdAt: now,
		updatedAt: now,
	}
}

// ReconstructUserProfile reconstructs a user profile from persistence
func ReconstructUserProfile(
	username valueobject.Username,
	name valueobject.DisplayName,
	avatar valueobject.AvatarURL,
	createdAt, updatedAt time.Time,
) *UserProfile {
	return &UserProfile{
		username:  username,
		name:      name,
		avatar:    avatar,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// UpdateUsername updates the username
func (p *UserProfile) UpdateUsername(username valueobject.Username) {
	if !p.username.Equals(username) {
		p.username = username
		p.updatedAt = time.Now()
	}
}

// UpdateName updates the display name
func (p *UserProfile) UpdateName(name valueobject.DisplayName) {
	if !p.name.Equals(name) {
		p.name = name
		p.updatedAt = time.Now()
	}
}

// UpdateAvatar updates the avatar URL
func (p *UserProfile) UpdateAvatar(avatar valueobject.AvatarURL) {
	if !p.avatar.Equals(avatar) {
		p.avatar = avatar
		p.updatedAt = time.Now()
	}
}

// UpdateAll updates all profile fields at once
func (p *UserProfile) UpdateAll(username valueobject.Username, name valueobject.DisplayName, avatar valueobject.AvatarURL) []string {
	var updatedFields []string
	
	if !p.username.Equals(username) {
		p.username = username
		updatedFields = append(updatedFields, "username")
	}
	
	if !p.name.Equals(name) {
		p.name = name
		updatedFields = append(updatedFields, "name")
	}
	
	if !p.avatar.Equals(avatar) {
		p.avatar = avatar
		updatedFields = append(updatedFields, "avatar")
	}
	
	if len(updatedFields) > 0 {
		p.updatedAt = time.Now()
	}
	
	return updatedFields
}

// Username returns the username
func (p *UserProfile) Username() valueobject.Username {
	return p.username
}

// Name returns the display name
func (p *UserProfile) Name() valueobject.DisplayName {
	return p.name
}

// Avatar returns the avatar URL
func (p *UserProfile) Avatar() valueobject.AvatarURL {
	return p.avatar
}

// CreatedAt returns when the profile was created
func (p *UserProfile) CreatedAt() time.Time {
	return p.createdAt
}

// UpdatedAt returns when the profile was last updated
func (p *UserProfile) UpdatedAt() time.Time {
	return p.updatedAt
}