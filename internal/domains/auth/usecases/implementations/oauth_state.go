package implementations

import (
	"crypto/subtle"
	"fmt"
	"sync"
	"time"

	"linke/internal/domains/auth/dto"
)

// OAuthStateStore manages OAuth state parameters to prevent CSRF attacks
type OAuthStateStore struct {
	states map[string]*dto.OAuthStateInfo
	mutex  sync.RWMutex
}

// NewOAuthStateStore creates a new OAuth state store
func NewOAuthStateStore() *OAuthStateStore {
	store := &OAuthStateStore{
		states: make(map[string]*dto.OAuthStateInfo),
	}

	// Start cleanup goroutine
	go store.cleanupExpiredStates()

	return store
}

// StoreState stores OAuth state information
func (s *OAuthStateStore) StoreState(state string, info *dto.OAuthStateInfo) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Set expiration time if not set (default 10 minutes)
	if info.ExpiresAt.IsZero() {
		info.ExpiresAt = time.Now().Add(10 * time.Minute)
	}

	if info.CreatedAt.IsZero() {
		info.CreatedAt = time.Now()
	}

	s.states[state] = info
}

// GetState retrieves and removes OAuth state information using constant-time comparison
func (s *OAuthStateStore) GetState(state string) (*dto.OAuthStateInfo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Use constant-time comparison to prevent timing attacks
	var foundInfo *dto.OAuthStateInfo
	var found bool
	
	for storedState, info := range s.states {
		if subtle.ConstantTimeCompare([]byte(state), []byte(storedState)) == 1 {
			foundInfo = info
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("invalid or expired state parameter")
	}

	// Check if expired
	if time.Now().After(foundInfo.ExpiresAt) {
		delete(s.states, state)
		return nil, fmt.Errorf("state parameter has expired")
	}

	// Remove state after use (single use)
	delete(s.states, state)

	return foundInfo, nil
}

// ValidateState checks if a state parameter is valid without removing it using constant-time comparison
func (s *OAuthStateStore) ValidateState(state string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Use constant-time comparison to prevent timing attacks
	for storedState, info := range s.states {
		if subtle.ConstantTimeCompare([]byte(state), []byte(storedState)) == 1 {
			return time.Now().Before(info.ExpiresAt)
		}
	}

	return false
}

// cleanupExpiredStates removes expired states every minute
func (s *OAuthStateStore) cleanupExpiredStates() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		for state, info := range s.states {
			if now.After(info.ExpiresAt) {
				delete(s.states, state)
			}
		}
		s.mutex.Unlock()
	}
}

// GetStatsString returns statistics about stored states
func (s *OAuthStateStore) GetStatsString() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return fmt.Sprintf("Active states: %d", len(s.states))
}
