package interfaces

import (
	"linke/internal/domains/auth/dto"
)

// OAuthStateService defines the interface for OAuth state management operations
type OAuthStateService interface {
	// State management
	StoreState(state string, info *dto.OAuthStateInfo)
	GetState(state string) (*dto.OAuthStateInfo, error)
	ValidateState(state string) bool

	// Statistics
	GetStatsString() string
}

