package interfaces

import "time"

// OAuthStateService defines the interface for OAuth state management operations
type OAuthStateService interface {
	// State management
	StoreState(state string, info *OAuthStateInfo)
	GetState(state string) (*OAuthStateInfo, error)
	ValidateState(state string) bool

	// Statistics
	GetStatsString() string
}

// OAuthStateInfo contains OAuth state information
type OAuthStateInfo struct {
	Provider    string    `json:"provider"`
	RedirectURI string    `json:"redirect_uri,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}
