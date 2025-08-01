package valueobject

import (
	"errors"
	"strings"
)

// Provider represents an authentication provider
type Provider struct {
	value string
}

// Provider constants
const (
	ProviderLocal    = "local"
	ProviderGoogle   = "google"
	ProviderGitHub   = "github"
	ProviderTelegram = "telegram"
)

// NewProvider creates a new Provider with validation
func NewProvider(provider string) (Provider, error) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	
	switch provider {
	case ProviderLocal, ProviderGoogle, ProviderGitHub, ProviderTelegram:
		return Provider{value: provider}, nil
	default:
		return Provider{}, errors.New("invalid provider: must be local, google, github, or telegram")
	}
}

// String returns the string representation of the provider
func (p Provider) String() string {
	return p.value
}

// Value returns the provider value
func (p Provider) Value() string {
	return p.value
}

// IsLocal checks if the provider is local
func (p Provider) IsLocal() bool {
	return p.value == ProviderLocal
}

// IsOAuth checks if the provider is OAuth-based
func (p Provider) IsOAuth() bool {
	return p.value != ProviderLocal
}

// IsGoogle checks if the provider is Google
func (p Provider) IsGoogle() bool {
	return p.value == ProviderGoogle
}

// IsGitHub checks if the provider is GitHub
func (p Provider) IsGitHub() bool {
	return p.value == ProviderGitHub
}

// IsTelegram checks if the provider is Telegram
func (p Provider) IsTelegram() bool {
	return p.value == ProviderTelegram
}

// Equals checks if two providers are equal
func (p Provider) Equals(other Provider) bool {
	return p.value == other.value
}

// Local returns a local provider
func Local() Provider {
	return Provider{value: ProviderLocal}
}

// Google returns a Google provider
func Google() Provider {
	return Provider{value: ProviderGoogle}
}

// GitHub returns a GitHub provider
func GitHub() Provider {
	return Provider{value: ProviderGitHub}
}

// Telegram returns a Telegram provider
func Telegram() Provider {
	return Provider{value: ProviderTelegram}
}

// ProviderID represents the ID from the OAuth provider
type ProviderID struct {
	value string
}

// NewProviderID creates a new ProviderID
func NewProviderID(id string) (ProviderID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return ProviderID{}, errors.New("provider ID cannot be empty")
	}

	return ProviderID{value: id}, nil
}

// String returns the string representation of the provider ID
func (p ProviderID) String() string {
	return p.value
}

// Value returns the provider ID value
func (p ProviderID) Value() string {
	return p.value
}

// IsEmpty checks if the provider ID is empty
func (p ProviderID) IsEmpty() bool {
	return strings.TrimSpace(p.value) == ""
}

// Equals checks if two provider IDs are equal
func (p ProviderID) Equals(other ProviderID) bool {
	return p.value == other.value
}