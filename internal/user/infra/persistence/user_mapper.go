package persistence

import (
	"strconv"
	"time"

	"linke/internal/user/domain/aggregate"
	"linke/internal/user/domain/entity"
	"linke/internal/user/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// UserMapper handles mapping between User domain model and UserPO persistence object
type UserMapper struct{}

// NewUserMapper creates a new UserMapper
func NewUserMapper() *UserMapper {
	return &UserMapper{}
}

// ToDomain converts UserPO to User domain model
func (m *UserMapper) ToDomain(po *UserPO) (*aggregate.User, error) {
	if po == nil {
		return nil, nil
	}

	// Create value objects - create shared UserID for aggregate
	sharedUserID, err := sharedvo.NewUserIDFromUint(po.ID)
	if err != nil {
		return nil, err
	}

	email, err := valueobject.NewEmail(po.Email)
	if err != nil {
		return nil, err
	}

	var password valueobject.Password
	if po.Password != "" {
		password, err = valueobject.NewPasswordFromHash(po.Password)
		if err != nil {
			return nil, err
		}
	}

	provider, err := valueobject.NewProvider(po.Provider)
	if err != nil {
		return nil, err
	}

	status, err := valueobject.NewUserStatus(po.Status)
	if err != nil {
		return nil, err
	}

	role, err := valueobject.NewUserRole(po.Role)
	if err != nil {
		return nil, err
	}

	// Create OAuth accounts map
	oauthAccounts := make(map[string]*entity.OAuthAccount)
	if po.GoogleID != nil {
		googleID, err := valueobject.NewProviderID(*po.GoogleID)
		if err == nil {
			provider, _ := valueobject.NewProvider("google")
			oauthAccounts["google"] = entity.NewOAuthAccount(provider, googleID, nil)
		}
	}
	if po.GitHubID != nil {
		githubID, err := valueobject.NewProviderID(*po.GitHubID)
		if err == nil {
			provider, _ := valueobject.NewProvider("github")
			oauthAccounts["github"] = entity.NewOAuthAccount(provider, githubID, nil)
		}
	}
	if po.TelegramID != nil {
		telegramID, err := valueobject.NewProviderID(*po.TelegramID)
		if err == nil {
			provider, _ := valueobject.NewProvider("telegram")
			oauthAccounts["telegram"] = entity.NewOAuthAccount(provider, telegramID, nil)
		}
	}

	// Create value objects for user fields
	username, err := valueobject.NewUsername(po.Username)
	if err != nil {
		return nil, err
	}

	name, err := valueobject.NewDisplayName(po.Name)
	if err != nil {
		return nil, err
	}

	avatar, err := valueobject.NewAvatarURL(po.Avatar)
	if err != nil {
		return nil, err
	}

	// Create invite code value object
	inviteCode, err := m.createInviteCodeFromPO(po)
	if err != nil {
		return nil, err
	}

	// Reconstruct user using the factory method with all data from persistence
	user := aggregate.ReconstructUser(
		sharedUserID,
		email,
		username,
		name,
		avatar,
		password,
		provider,
		status,
		role,
		oauthAccounts,
		po.ProviderData,
		inviteCode,
		po.CreatedAt,
		po.UpdatedAt,
		m.getDeletedAt(po),
	)

	return user, nil
}

// ToPersistence converts User domain model to UserPO persistence object
func (m *UserMapper) ToPersistence(user *aggregate.User) *UserPO {
	if user == nil {
		return nil
	}

	po := &UserPO{
		ID:             user.ID().ToUint(),
		Email:          user.Email().String(),
		Username:       user.Username().String(),
		Name:           user.Name().String(),
		Avatar:         user.Avatar().String(),
		Provider:       user.Provider().String(),
		Status:         user.Status().String(),
		Role:           user.Role().String(),
		ProviderData:   user.ProviderData(),
		InviteCodeID:   m.getInviteCodeIDFromUser(user),
		InviteCodeUsed: m.getInviteCodeUsedFromUser(user),
		CreatedAt:      user.CreatedAt(),
		UpdatedAt:      user.UpdatedAt(),
	}

	// Set password if it's a local account
	if user.Provider().IsLocal() {
		po.Password = user.Password().Hash()
	}

	// Set OAuth provider IDs
	oauthAccounts := user.OAuthAccounts()
	if googleAccount, exists := oauthAccounts["google"]; exists {
		googleIDStr := googleAccount.ProviderID().String()
		po.GoogleID = &googleIDStr
	}
	if githubAccount, exists := oauthAccounts["github"]; exists {
		githubIDStr := githubAccount.ProviderID().String()
		po.GitHubID = &githubIDStr
	}
	if telegramAccount, exists := oauthAccounts["telegram"]; exists {
		telegramIDStr := telegramAccount.ProviderID().String()
		po.TelegramID = &telegramIDStr
	}

	// Set deleted at if user is deleted
	if user.IsDeleted() {
		deletedAt := user.DeletedAt()
		if deletedAt != nil {
			po.DeletedAt.Time = *deletedAt
			po.DeletedAt.Valid = true
		}
	}

	return po
}

// ToDomainList converts slice of UserPO to slice of User domain models
func (m *UserMapper) ToDomainList(pos []*UserPO) ([]*aggregate.User, error) {
	if len(pos) == 0 {
		return []*aggregate.User{}, nil
	}

	users := make([]*aggregate.User, 0, len(pos))
	for _, po := range pos {
		user, err := m.ToDomain(po)
		if err != nil {
			return nil, err
		}
		if user != nil {
			users = append(users, user)
		}
	}

	return users, nil
}

// ToPersistenceList converts slice of User domain models to slice of UserPO
func (m *UserMapper) ToPersistenceList(users []*aggregate.User) []*UserPO {
	if len(users) == 0 {
		return []*UserPO{}
	}

	pos := make([]*UserPO, 0, len(users))
	for _, user := range users {
		po := m.ToPersistence(user)
		if po != nil {
			pos = append(pos, po)
		}
	}

	return pos
}

// Helper methods

// getDeletedAt safely gets the deleted at time
func (m *UserMapper) getDeletedAt(po *UserPO) *time.Time {
	if po.DeletedAt.Valid {
		return &po.DeletedAt.Time
	}
	return nil
}

// CreateUserIDFromString creates a UserID from string (for queries) - returns domain UserID
func (m *UserMapper) CreateUserIDFromString(id string) (valueobject.UserID, error) {
	// Try to parse as uint first (for legacy compatibility)
	if uintID, err := strconv.ParseUint(id, 10, 32); err == nil {
		return valueobject.NewUserIDFromUint(uint(uintID)), nil
	}
	
	// Otherwise, treat as string ID
	return valueobject.NewUserIDFromString(id)
}

// CreateUserIDsFromStrings creates UserIDs from string slice - returns domain UserIDs
func (m *UserMapper) CreateUserIDsFromStrings(ids []string) ([]valueobject.UserID, error) {
	userIDs := make([]valueobject.UserID, 0, len(ids))
	for _, id := range ids {
		userID, err := m.CreateUserIDFromString(id)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

// Helper functions for invite code access
func (m *UserMapper) getInviteCodeIDFromUser(user *aggregate.User) *uint {
	inviteCode := user.InviteCode()
	if inviteCode.IsEmpty() {
		return nil
	}
	id := inviteCode.ID()
	return &id
}

func (m *UserMapper) getInviteCodeUsedFromUser(user *aggregate.User) *string {
	inviteCode := user.InviteCode()
	if inviteCode.IsEmpty() {
		return nil
	}
	code := inviteCode.String()
	return &code
}

// createInviteCodeFromPO creates invite code value object from persistence data
func (m *UserMapper) createInviteCodeFromPO(po *UserPO) (valueobject.InviteCode, error) {
	if po.InviteCodeID == nil || po.InviteCodeUsed == nil {
		return valueobject.NewEmptyInviteCode(), nil
	}
	return valueobject.NewInviteCode(*po.InviteCodeID, *po.InviteCodeUsed)
}