package implementations

import (
	"testing"

	"linke/internal/domains/user/entities"
	"linke/internal/shared/cache"
	"linke/internal/shared/framework"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Mock logger for testing
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...zap.Field)                  {}
func (m *MockLogger) Info(msg string, fields ...zap.Field)                   {}
func (m *MockLogger) Warn(msg string, fields ...zap.Field)                   {}
func (m *MockLogger) Error(msg string, fields ...zap.Field)                  {}
func (m *MockLogger) Fatal(msg string, fields ...zap.Field)                  {}
func (m *MockLogger) With(fields ...zap.Field) framework.Logger              { return m }
func (m *MockLogger) Sync() error                                            { return nil }

func TestCachedUserService_CacheKeyGeneration(t *testing.T) {
	cacheKeys := cache.NewAllCacheKeys()
	
	// Test key generation
	userID := uint(123)
	email := "user@example.com"
	username := "username123"
	
	keyByID := cacheKeys.User.UserByID(userID)
	keyByEmail := cacheKeys.User.UserByEmail(email)
	keyByUsername := cacheKeys.User.UserByUsername(username)
	profileKey := cacheKeys.User.UserProfile(userID)
	
	// Verify keys are properly formatted
	assert.Contains(t, keyByID, "user:")
	assert.Contains(t, keyByID, "id")
	assert.Contains(t, keyByID, "123")
	
	assert.Contains(t, keyByEmail, "user:")
	assert.Contains(t, keyByEmail, "email")
	assert.Contains(t, keyByEmail, email)
	
	assert.Contains(t, keyByUsername, "user:")
	assert.Contains(t, keyByUsername, "username")
	assert.Contains(t, keyByUsername, username)
	
	assert.Contains(t, profileKey, "user:")
	assert.Contains(t, profileKey, "profile")
	assert.Contains(t, profileKey, "123")
}

func TestCachedUserService_UserEntity(t *testing.T) {
	// Test user entity methods that are used in caching
	testUser := &entities.User{
		ID:       1,
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
		Status:   entities.UserStatusActive,
		Role:     entities.UserRoleUser,
		Provider: entities.ProviderLocal,
	}
	
	// Test user status checks
	assert.True(t, testUser.IsActive())
	assert.False(t, testUser.IsAdmin())
	assert.True(t, testUser.IsLocalAccount())
	assert.False(t, testUser.IsOAuthAccount())
	assert.False(t, testUser.IsDeleted())
	
	// Test user response conversion
	response := testUser.ToResponse()
	assert.NotNil(t, response)
	assert.Equal(t, testUser.ID, response.ID)
	assert.Equal(t, testUser.Email, response.Email)
	assert.Equal(t, testUser.Username, response.Username)
	assert.Equal(t, testUser.Status, response.Status)
}

func TestCachedUserService_KeyConstruction(t *testing.T) {
	cacheKeys := cache.NewAllCacheKeys()
	
	// Test that different users get different cache keys
	user1ID := uint(1)
	user2ID := uint(2)
	
	key1 := cacheKeys.User.UserByID(user1ID)
	key2 := cacheKeys.User.UserByID(user2ID)
	
	assert.NotEqual(t, key1, key2)
	assert.Contains(t, key1, "1")
	assert.Contains(t, key2, "2")
	
	// Test email-based keys
	email1 := "user1@example.com"
	email2 := "user2@example.com"
	
	emailKey1 := cacheKeys.User.UserByEmail(email1)
	emailKey2 := cacheKeys.User.UserByEmail(email2)
	
	assert.NotEqual(t, emailKey1, emailKey2)
	assert.Contains(t, emailKey1, email1)
	assert.Contains(t, emailKey2, email2)
}