package implementations

import (
	"fmt"
	"sync"
	"time"

	"linke/internal/domains/auth/usecases/interfaces"
)

// OAuthStateStore manages OAuth state parameters to prevent CSRF attacks
type OAuthStateStore struct {
	states map[string]*interfaces.OAuthStateInfo
	mutex  sync.RWMutex
}


// NewOAuthStateStore creates a new OAuth state store
func NewOAuthStateStore() *OAuthStateStore {
	store := &OAuthStateStore{
		states: make(map[string]*interfaces.OAuthStateInfo),
	}
	
	// Start cleanup goroutine
	go store.cleanupExpiredStates()
	
	return store
}

// StoreState stores OAuth state information
func (s *OAuthStateStore) StoreState(state string, info *interfaces.OAuthStateInfo) {
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

// GetState retrieves and removes OAuth state information
func (s *OAuthStateStore) GetState(state string) (*interfaces.OAuthStateInfo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	info, exists := s.states[state]
	if !exists {
		return nil, fmt.Errorf("invalid or expired state parameter")
	}
	
	// Check if expired
	if time.Now().After(info.ExpiresAt) {
		delete(s.states, state)
		return nil, fmt.Errorf("state parameter has expired")
	}
	
	// Remove state after use (single use)
	delete(s.states, state)
	
	return info, nil
}

// ValidateState checks if a state parameter is valid without removing it
func (s *OAuthStateStore) ValidateState(state string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	info, exists := s.states[state]
	if !exists {
		return false
	}
	
	return time.Now().Before(info.ExpiresAt)
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