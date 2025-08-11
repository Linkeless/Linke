package interfaces

import (
	"context"
	"golang.org/x/oauth2"
)

// OAuthService defines the interface for OAuth authentication operations
type OAuthService interface {
	// OAuth flow management
	GenerateState() string
	GetAuthURL(provider, state string) (string, error)
	ExchangeCodeForToken(ctx context.Context, provider, code string) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, provider string, token *oauth2.Token) (*UserInfo, error)

	// Telegram specific
	VerifyTelegramAuth(data map[string]string) (*UserInfo, error)
	GetTelegramLoginURL() string
}

// 注意：UserInfo 已统一到 auth_service.go 中定义
// 请使用 interfaces.UserInfo

// AuthorizeURLRequest represents request for authorization URL
type AuthorizeURLRequest struct {
	Provider    string   `json:"provider" binding:"required"`
	RedirectURI string   `json:"redirect_uri,omitempty"`
	State       string   `json:"state,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
}

// AuthorizeURLResponse represents response with authorization URL
type AuthorizeURLResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

// TelegramUser represents Telegram user data
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}
