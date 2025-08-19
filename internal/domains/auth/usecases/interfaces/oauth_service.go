package interfaces

import (
	"context"

	"golang.org/x/oauth2"
	"linke/internal/domains/auth/dto"
)

// OAuthService defines the interface for OAuth authentication operations
type OAuthService interface {
	// OAuth flow management
	GenerateState() string
	GetAuthURL(provider, state string) (string, error)
	ExchangeCodeForToken(ctx context.Context, provider, code string) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, provider string, token *oauth2.Token) (*dto.UserInfo, error)

	// Telegram specific
	VerifyTelegramAuth(data map[string]string) (*dto.UserInfo, error)
	GetTelegramLoginURL() string
}

// 注意：UserInfo 已统一到 dto 包中定义
// 请使用 dto.UserInfo
