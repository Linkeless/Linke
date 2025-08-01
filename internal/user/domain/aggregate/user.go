package aggregate

import (
	"errors"
	"time"

	"linke/internal/shared/domain"
	sharedvo "linke/internal/shared/valueobject"
	"linke/internal/user/domain/entity"
	"linke/internal/user/domain/event"
	"linke/internal/user/domain/valueobject"
)

// User represents the user aggregate root
type User struct {
	// Core identity
	id    sharedvo.UserID
	email valueobject.Email
	
	// Profile information (managed as entity)
	profile *entity.UserProfile
	
	// Authentication
	password valueobject.Password
	provider valueobject.Provider
	status   valueobject.UserStatus
	role     valueobject.UserRole
	
	// OAuth accounts (managed as entities)
	oauthAccounts map[string]*entity.OAuthAccount
	
	// Provider metadata for the primary provider
	providerData *string
	
	// Invite code information
	inviteCode valueobject.InviteCode
	
	// Timestamps
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
	
	// Domain events
	domainEvents []domain.DomainEvent
}

// NewUser creates a new User with email and password (local account)
func NewUser(email, password string) (*User, error) {
	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, err
	}
	
	passwordVO, err := valueobject.NewPassword(password)
	if err != nil {
		return nil, err
	}
	
	userID := sharedvo.NewUserID()
	status := valueobject.ActiveStatus()
	role := valueobject.User()
	provider := valueobject.Local()
	
	// Generate profile information from email
	localPart := emailVO.LocalPart()
	username := valueobject.GenerateUsernameFromEmail(localPart)
	name := valueobject.GenerateDisplayNameFromEmail(localPart)
	avatar := valueobject.NewEmptyAvatarURL()
	
	profile := entity.NewUserProfile(username, name, avatar)
	inviteCode := valueobject.NewEmptyInviteCode()
	
	user := &User{
		id:            userID,
		email:         emailVO,
		profile:       profile,
		password:      passwordVO,
		provider:      provider,
		status:        status,
		role:          role,
		oauthAccounts: make(map[string]*entity.OAuthAccount),
		inviteCode:    inviteCode,
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
		domainEvents:  make([]domain.DomainEvent, 0),
	}
	
	// Raise domain event
	domainUserID, _ := valueobject.ConvertFromSharedUserID(userID)
	user.raiseEvent(event.NewUserCreated(domainUserID, emailVO, provider))
	
	return user, nil
}

// NewUserFromOAuth creates a new User from OAuth provider
func NewUserFromOAuth(
	email, name, username, avatar string,
	provider string,
	providerID string,
	providerData *string,
) (*User, error) {
	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, err
	}
	
	providerVO, err := valueobject.NewProvider(provider)
	if err != nil {
		return nil, err
	}
	
	providerIDVO, err := valueobject.NewProviderID(providerID)
	if err != nil {
		return nil, err
	}
	
	userID := sharedvo.NewUserID()
	status := valueobject.ActiveStatus()
	role := valueobject.User()
	
	// Create profile with provided or generated values
	var usernameVO valueobject.Username
	if username != "" {
		usernameVO, err = valueobject.NewUsername(username)
		if err != nil {
			usernameVO = valueobject.GenerateUsernameFromEmail(emailVO.LocalPart())
		}
	} else {
		usernameVO = valueobject.GenerateUsernameFromEmail(emailVO.LocalPart())
	}
	
	var nameVO valueobject.DisplayName
	if name != "" {
		nameVO, err = valueobject.NewDisplayName(name)
		if err != nil {
			nameVO = valueobject.GenerateDisplayNameFromEmail(emailVO.LocalPart())
		}
	} else {
		nameVO = valueobject.GenerateDisplayNameFromEmail(emailVO.LocalPart())
	}
	
	var avatarVO valueobject.AvatarURL
	if avatar != "" {
		avatarVO, err = valueobject.NewAvatarURL(avatar)
		if err != nil {
			avatarVO = valueobject.NewEmptyAvatarURL()
		}
	} else {
		avatarVO = valueobject.NewEmptyAvatarURL()
	}
	
	profile := entity.NewUserProfile(usernameVO, nameVO, avatarVO)
	inviteCode := valueobject.NewEmptyInviteCode()
	
	// Create OAuth account
	oauthAccount := entity.NewOAuthAccount(providerVO, providerIDVO, providerData)
	oauthAccounts := map[string]*entity.OAuthAccount{
		provider: oauthAccount,
	}
	
	user := &User{
		id:            userID,
		email:         emailVO,
		profile:       profile,
		password:      valueobject.Password{}, // Empty for OAuth accounts
		provider:      providerVO,
		status:        status,
		role:          role,
		oauthAccounts: oauthAccounts,
		providerData:  providerData,
		inviteCode:    inviteCode,
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
		domainEvents:  make([]domain.DomainEvent, 0),
	}
	
	// Raise domain event
	domainUserID, _ := valueobject.ConvertFromSharedUserID(userID)
	user.raiseEvent(event.NewUserCreated(domainUserID, emailVO, providerVO))
	
	return user, nil
}

// ReconstructUser reconstructs a User from persistence data
func ReconstructUser(
	id sharedvo.UserID,
	email valueobject.Email,
	username valueobject.Username,
	name valueobject.DisplayName,
	avatar valueobject.AvatarURL,
	password valueobject.Password,
	provider valueobject.Provider,
	status valueobject.UserStatus,
	role valueobject.UserRole,
	oauthAccounts map[string]*entity.OAuthAccount,
	providerData *string,
	inviteCode valueobject.InviteCode,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) *User {
	profile := entity.ReconstructUserProfile(username, name, avatar, createdAt, updatedAt)
	
	return &User{
		id:            id,
		email:         email,
		profile:       profile,
		password:      password,
		provider:      provider,
		status:        status,
		role:          role,
		oauthAccounts: oauthAccounts,
		providerData:  providerData,
		inviteCode:    inviteCode,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		deletedAt:     deletedAt,
		domainEvents:  make([]domain.DomainEvent, 0),
	}
}

// Authenticate performs user authentication for local accounts
func (u *User) Authenticate(password, ipAddress, userAgent string) error {
	// OAuth accounts cannot authenticate with password
	if !u.provider.IsLocal() {
		u.raiseEvent(event.NewUserLoginFailed(u.email, "authentication not supported for OAuth accounts", ipAddress, userAgent))
		return errors.New("authentication not supported for OAuth accounts")
	}
	
	// Suspended users cannot authenticate
	if u.status.IsSuspended() {
		u.raiseEvent(event.NewUserLoginFailed(u.email, "user account is suspended", ipAddress, userAgent))
		return errors.New("user account is suspended")
	}
	
	// Inactive users cannot authenticate
	if !u.status.IsActive() {
		u.raiseEvent(event.NewUserLoginFailed(u.email, "user account is not active", ipAddress, userAgent))
		return errors.New("user account is not active")
	}
	
	// Verify password
	if !u.password.Verify(password) {
		u.raiseEvent(event.NewUserLoginFailed(u.email, "invalid credentials", ipAddress, userAgent))
		return errors.New("invalid credentials")
	}
	
	// Raise successful login event
	domainUserID, _ := valueobject.ConvertFromSharedUserID(u.id)
	u.raiseEvent(event.NewUserLoggedIn(domainUserID, u.provider, ipAddress, userAgent))
	
	return nil
}

// ChangePassword changes the user's password (for local accounts only)
func (u *User) ChangePassword(oldPassword, newPassword string) error {
	if !u.provider.IsLocal() {
		return errors.New("password change not supported for OAuth accounts")
	}
	
	// Verify current password
	if !u.password.Verify(oldPassword) {
		return errors.New("current password is incorrect")
	}
	
	// Create new password value object (this will validate the new password)
	newPasswordVO, err := valueobject.NewPassword(newPassword)
	if err != nil {
		return err
	}
	
	u.password = newPasswordVO
	u.updatedAt = time.Now()
	
	// Raise domain event
	domainUserID, _ := valueobject.ConvertFromSharedUserID(u.id)
	u.raiseEvent(event.NewUserPasswordChanged(domainUserID))
	
	return nil
}

// ChangeStatus changes the user's status
func (u *User) ChangeStatus(newStatus string) error {
	newStatusVO, err := valueobject.NewUserStatus(newStatus)
	if err != nil {
		return err
	}
	
	if u.status.Equals(newStatusVO) {
		return nil // No change needed
	}
	
	oldStatus := u.status
	u.status = newStatusVO
	u.updatedAt = time.Now()
	
	// Raise domain event
	domainUserID, _ := valueobject.ConvertFromSharedUserID(u.id)
	u.raiseEvent(event.NewUserStatusChanged(domainUserID, oldStatus, newStatusVO))
	
	return nil
}

// ChangeRole changes the user's role
func (u *User) ChangeRole(newRole string) error {
	newRoleVO, err := valueobject.NewUserRole(newRole)
	if err != nil {
		return err
	}
	
	if u.role.Equals(newRoleVO) {
		return nil // No change needed
	}
	
	oldRole := u.role
	u.role = newRoleVO
	u.updatedAt = time.Now()
	
	// Raise domain event
	domainUserID, _ := valueobject.ConvertFromSharedUserID(u.id)
	u.raiseEvent(event.NewUserRoleChanged(domainUserID, oldRole, newRoleVO))
	
	return nil
}

// UpdateProfile updates the user's profile information
func (u *User) UpdateProfile(username, name, avatar string) error {
	var usernameVO valueobject.Username
	var nameVO valueobject.DisplayName
	var avatarVO valueobject.AvatarURL
	var err error
	
	// Parse and validate new values
	if username != "" {
		usernameVO, err = valueobject.NewUsername(username)
		if err != nil {
			return err
		}
	} else {
		usernameVO = u.profile.Username()
	}
	
	if name != "" {
		nameVO, err = valueobject.NewDisplayName(name)
		if err != nil {
			return err
		}
	} else {
		nameVO = u.profile.Name()
	}
	
	if avatar != "" {
		avatarVO, err = valueobject.NewAvatarURL(avatar)
		if err != nil {
			return err
		}
	} else {
		avatarVO = u.profile.Avatar()
	}
	
	// Update profile and get changed fields
	updatedFields := u.profile.UpdateAll(usernameVO, nameVO, avatarVO)
	
	if len(updatedFields) > 0 {
		u.updatedAt = time.Now()
		// Raise domain event
		domainUserID, _ := valueobject.ConvertFromSharedUserID(u.id)
		u.raiseEvent(event.NewUserProfileUpdated(domainUserID, updatedFields))
	}
	
	return nil
}

// SetInviteCode sets the invite code information
func (u *User) SetInviteCode(inviteCodeID uint, inviteCodeStr string) error {
	inviteCode, err := valueobject.NewInviteCode(inviteCodeID, inviteCodeStr)
	if err != nil {
		return err
	}
	
	u.inviteCode = inviteCode
	u.updatedAt = time.Now()
	
	return nil
}

// AddOAuthAccount adds a new OAuth account
func (u *User) AddOAuthAccount(provider, providerID string, providerData *string) error {
	providerVO, err := valueobject.NewProvider(provider)
	if err != nil {
		return err
	}
	
	providerIDVO, err := valueobject.NewProviderID(providerID)
	if err != nil {
		return err
	}
	
	// Check if account already exists
	if _, exists := u.oauthAccounts[provider]; exists {
		return errors.New("OAuth account already exists for this provider")
	}
	
	oauthAccount := entity.NewOAuthAccount(providerVO, providerIDVO, providerData)
	u.oauthAccounts[provider] = oauthAccount
	u.updatedAt = time.Now()
	
	return nil
}

// RemoveOAuthAccount removes an OAuth account
func (u *User) RemoveOAuthAccount(provider string) error {
	if _, exists := u.oauthAccounts[provider]; !exists {
		return errors.New("OAuth account not found for this provider")
	}
	
	delete(u.oauthAccounts, provider)
	u.updatedAt = time.Now()
	
	return nil
}

// SoftDelete performs soft delete on the user
func (u *User) SoftDelete() {
	now := time.Now()
	u.deletedAt = &now
	u.updatedAt = now
}

// Restore restores a soft deleted user
func (u *User) Restore() {
	u.deletedAt = nil
	u.updatedAt = time.Now()
}

// Business logic methods

// IsActive checks if the user is active and not deleted
func (u *User) IsActive() bool {
	return u.status.IsActive() && !u.IsDeleted()
}

// IsAdmin checks if the user is an admin and active
func (u *User) IsAdmin() bool {
	return u.role.IsAdmin() && u.IsActive()
}

// IsDeleted checks if the user is soft deleted
func (u *User) IsDeleted() bool {
	return u.deletedAt != nil
}

// IsLocalAccount checks if this is a local account
func (u *User) IsLocalAccount() bool {
	return u.provider.IsLocal()
}

// IsOAuthAccount checks if this is an OAuth account
func (u *User) IsOAuthAccount() bool {
	return u.provider.IsOAuth()
}

// CanLogin checks if the user can login
func (u *User) CanLogin() bool {
	return u.status.IsActive() && !u.status.IsSuspended() && !u.IsDeleted()
}

// HasOAuthAccount checks if user has an OAuth account for the given provider
func (u *User) HasOAuthAccount(provider string) bool {
	_, exists := u.oauthAccounts[provider]
	return exists
}

// GetOAuthAccount returns the OAuth account for the given provider
func (u *User) GetOAuthAccount(provider string) (*entity.OAuthAccount, bool) {
	account, exists := u.oauthAccounts[provider]
	return account, exists
}

// Getters

// ID returns the user ID (converted to domain type)
func (u *User) ID() valueobject.UserID {
	domainUserID, _ := valueobject.ConvertFromSharedUserID(u.id)
	return domainUserID
}

// Email returns the user email
func (u *User) Email() valueobject.Email {
	return u.email
}

// Profile returns the user profile
func (u *User) Profile() *entity.UserProfile {
	return u.profile
}

// Username returns the username
func (u *User) Username() valueobject.Username {
	return u.profile.Username()
}

// Name returns the user name
func (u *User) Name() valueobject.DisplayName {
	return u.profile.Name()
}

// Avatar returns the avatar URL
func (u *User) Avatar() valueobject.AvatarURL {
	return u.profile.Avatar()
}

// Password returns the password (for internal use)
func (u *User) Password() valueobject.Password {
	return u.password
}

// Provider returns the authentication provider
func (u *User) Provider() valueobject.Provider {
	return u.provider
}

// Status returns the user status
func (u *User) Status() valueobject.UserStatus {
	return u.status
}

// Role returns the user role
func (u *User) Role() valueobject.UserRole {
	return u.role
}

// OAuthAccounts returns all OAuth accounts
func (u *User) OAuthAccounts() map[string]*entity.OAuthAccount {
	// Return a copy to prevent external modification
	accounts := make(map[string]*entity.OAuthAccount)
	for k, v := range u.oauthAccounts {
		accounts[k] = v
	}
	return accounts
}

// ProviderData returns the provider metadata
func (u *User) ProviderData() *string {
	return u.providerData
}

// InviteCode returns the invite code
func (u *User) InviteCode() valueobject.InviteCode {
	return u.inviteCode
}

// CreatedAt returns the creation time
func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

// UpdatedAt returns the last update time
func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// DeletedAt returns the deletion time
func (u *User) DeletedAt() *time.Time {
	return u.deletedAt
}

// Domain events methods

// DomainEvents returns the domain events
func (u *User) DomainEvents() []domain.DomainEvent {
	return u.domainEvents
}

// ClearDomainEvents clears the domain events
func (u *User) ClearDomainEvents() {
	u.domainEvents = make([]domain.DomainEvent, 0)
}

// raiseEvent adds a domain event
func (u *User) raiseEvent(domainEvent domain.DomainEvent) {
	u.domainEvents = append(u.domainEvents, domainEvent)
}