package entity

import (
	"time"

	"linke/internal/user/domain/valueobject"
)

// OAuthAccount represents an OAuth account linked to a user
type OAuthAccount struct {
	// Identity
	provider   valueobject.Provider
	providerID valueobject.ProviderID
	
	// OAuth-specific data
	providerData *string // JSON metadata from provider
	
	// Timestamps
	linkedAt  time.Time
	updatedAt time.Time
}

// NewOAuthAccount creates a new OAuth account
func NewOAuthAccount(provider valueobject.Provider, providerID valueobject.ProviderID, providerData *string) *OAuthAccount {
	now := time.Now()
	return &OAuthAccount{
		provider:     provider,
		providerID:   providerID,
		providerData: providerData,
		linkedAt:     now,
		updatedAt:    now,
	}
}

// ReconstructOAuthAccount reconstructs an OAuth account from persistence
func ReconstructOAuthAccount(
	provider valueobject.Provider,
	providerID valueobject.ProviderID,
	providerData *string,
	linkedAt, updatedAt time.Time,
) *OAuthAccount {
	return &OAuthAccount{
		provider:     provider,
		providerID:   providerID,
		providerData: providerData,
		linkedAt:     linkedAt,
		updatedAt:    updatedAt,
	}
}

// UpdateProviderData updates the provider-specific metadata
func (o *OAuthAccount) UpdateProviderData(providerData *string) {
	o.providerData = providerData
	o.updatedAt = time.Now()
}

// Provider returns the OAuth provider
func (o *OAuthAccount) Provider() valueobject.Provider {
	return o.provider
}

// ProviderID returns the provider-specific ID
func (o *OAuthAccount) ProviderID() valueobject.ProviderID {
	return o.providerID
}

// ProviderData returns the provider metadata
func (o *OAuthAccount) ProviderData() *string {
	return o.providerData
}

// LinkedAt returns when the account was linked
func (o *OAuthAccount) LinkedAt() time.Time {
	return o.linkedAt
}

// UpdatedAt returns when the account was last updated
func (o *OAuthAccount) UpdatedAt() time.Time {
	return o.updatedAt
}

// IsProvider checks if this account is for the given provider
func (o *OAuthAccount) IsProvider(provider valueobject.Provider) bool {
	return o.provider.Equals(provider)
}

// HasProviderID checks if this account has the given provider ID
func (o *OAuthAccount) HasProviderID(providerID valueobject.ProviderID) bool {
	return o.providerID.Equals(providerID)
}

// Key returns a unique key for this OAuth account (provider name)
func (o *OAuthAccount) Key() string {
	return o.provider.String()
}