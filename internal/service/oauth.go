package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"linke/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type OAuthService struct {
	cfg *config.Config
}

type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Provider string `json:"provider"`
}

type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

func NewOAuthService(cfg *config.Config) *OAuthService {
	return &OAuthService{
		cfg: cfg,
	}
}

func (o *OAuthService) GetAuthURL(provider, state string) (string, error) {
	config := o.getOAuth2Config(provider)
	if config == nil {
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}

	return config.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

func (o *OAuthService) ExchangeCodeForToken(ctx context.Context, provider, code string) (*oauth2.Token, error) {
	config := o.getOAuth2Config(provider)
	if config == nil {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	return config.Exchange(ctx, code)
}

func (o *OAuthService) GetUserInfo(ctx context.Context, provider string, token *oauth2.Token) (*UserInfo, error) {
	switch provider {
	case "google":
		return o.getGoogleUserInfo(ctx, token)
	case "github":
		return o.getGitHubUserInfo(ctx, token)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (o *OAuthService) VerifyTelegramAuth(data map[string]string) (*UserInfo, error) {
	hash, exists := data["hash"]
	if !exists {
		return nil, fmt.Errorf("hash not found in auth data")
	}

	authDate, err := strconv.ParseInt(data["auth_date"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid auth_date: %w", err)
	}

	if time.Now().Unix()-authDate > 86400 {
		return nil, fmt.Errorf("auth data is too old")
	}

	if !o.verifyTelegramHash(data, hash) {
		return nil, fmt.Errorf("invalid hash")
	}

	id := data["id"]
	firstName := data["first_name"]
	lastName := data["last_name"]
	username := data["username"]
	photoURL := data["photo_url"]

	name := firstName
	if lastName != "" {
		name += " " + lastName
	}

	return &UserInfo{
		ID:       id,
		Name:     name,
		Username: username,
		Avatar:   photoURL,
		Provider: "telegram",
	}, nil
}

func (o *OAuthService) GetTelegramLoginURL() string {
	botToken := o.cfg.OAuth2.TelegramBotToken
	if botToken == "" {
		return ""
	}

	return fmt.Sprintf("https://oauth.telegram.org/auth?bot_id=%s&origin=%s&return_to=%s",
		strings.TrimPrefix(botToken, "bot"),
		"http://localhost:8080",
		url.QueryEscape(o.cfg.OAuth2.TelegramRedirectURL))
}

func (o *OAuthService) getOAuth2Config(provider string) *oauth2.Config {
	switch provider {
	case "google":
		return &oauth2.Config{
			ClientID:     o.cfg.OAuth2.GoogleClientID,
			ClientSecret: o.cfg.OAuth2.GoogleClientSecret,
			RedirectURL:  o.cfg.OAuth2.GoogleRedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	case "github":
		return &oauth2.Config{
			ClientID:     o.cfg.OAuth2.GitHubClientID,
			ClientSecret: o.cfg.OAuth2.GitHubClientSecret,
			RedirectURL:  o.cfg.OAuth2.GitHubRedirectURL,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}
	default:
		return nil
	}
}

func (o *OAuthService) getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	config := o.getOAuth2Config("google")
	client := config.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &UserInfo{
		ID:       googleUser.ID,
		Email:    googleUser.Email,
		Name:     googleUser.Name,
		Avatar:   googleUser.Picture,
		Provider: "google",
	}, nil
}

func (o *OAuthService) getGitHubUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	config := o.getOAuth2Config("github")
	client := config.Client(ctx, token)

	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	userInfo := &UserInfo{
		ID:       strconv.Itoa(githubUser.ID),
		Email:    githubUser.Email,
		Name:     githubUser.Name,
		Username: githubUser.Login,
		Avatar:   githubUser.AvatarURL,
		Provider: "github",
	}

	if userInfo.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil && emailResp.StatusCode == http.StatusOK {
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if err := json.NewDecoder(emailResp.Body).Decode(&emails); err == nil {
				for _, email := range emails {
					if email.Primary {
						userInfo.Email = email.Email
						break
					}
				}
			}
			emailResp.Body.Close()
		}
	}

	return userInfo, nil
}

func (o *OAuthService) verifyTelegramHash(data map[string]string, hash string) bool {
	var keys []string
	for key := range data {
		if key != "hash" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	var dataCheckString strings.Builder
	for i, key := range keys {
		if i > 0 {
			dataCheckString.WriteString("\n")
		}
		dataCheckString.WriteString(fmt.Sprintf("%s=%s", key, data[key]))
	}

	secretKey := sha256.Sum256([]byte(o.cfg.OAuth2.TelegramBotToken))
	h := hmac.New(sha256.New, secretKey[:])
	h.Write([]byte(dataCheckString.String()))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	return expectedHash == hash
}