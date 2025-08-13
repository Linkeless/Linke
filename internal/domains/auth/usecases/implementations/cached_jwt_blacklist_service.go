package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/domains/auth/usecases/interfaces"
	"linke/internal/shared/cache"
	"linke/internal/shared/framework"
	"linke/internal/shared/logger"
)

// CachedJWTBlacklistService wraps JWTBlacklistService with caching for frequently accessed blacklist checks
type CachedJWTBlacklistService struct {
	baseService  *JWTBlacklistService
	cacheManager cache.CacheManager
	cacheKeys    *cache.AllCacheKeys
	logger       framework.Logger
}

// NewCachedJWTBlacklistService creates a new cached JWT blacklist service
func NewCachedJWTBlacklistService(
	baseService *JWTBlacklistService,
	cacheManager cache.CacheManager,
	cacheKeys *cache.AllCacheKeys,
	logger framework.Logger,
) interfaces.JWTBlacklistService {
	return &CachedJWTBlacklistService{
		baseService:  baseService,
		cacheManager: cacheManager,
		cacheKeys:    cacheKeys,
		logger:       logger,
	}
}

// BlacklistToken adds token to blacklist and invalidates related caches
func (s *CachedJWTBlacklistService) BlacklistToken(ctx context.Context, token string, userID *uint, reason string, expiresAt time.Time) error {
	err := s.baseService.BlacklistToken(ctx, token, userID, reason, expiresAt)
	if err != nil {
		return err
	}

	// Invalidate token cache immediately
	tokenHash := s.hashToken(token)
	s.invalidateTokenCache(ctx, tokenHash)

	// If userID is provided, also invalidate user blacklist cache
	if userID != nil {
		s.invalidateUserBlacklistCache(ctx, *userID)
	}

	s.logger.Info("Token blacklisted and cache invalidated",
		logger.String("token_hash", tokenHash),
		logger.String("reason", reason))

	return nil
}

// IsTokenBlacklisted checks if token is blacklisted with caching
func (s *CachedJWTBlacklistService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	tokenHash := s.hashToken(token)
	cacheKey := fmt.Sprintf("%stoken_blacklist:%s", cache.CachePrefixAuth, tokenHash)

	// Try cache first
	cached, err := s.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var result struct {
			IsBlacklisted bool      `json:"is_blacklisted"`
			CachedAt      time.Time `json:"cached_at"`
		}
		if err := json.Unmarshal(cached, &result); err == nil {
			s.logger.Debug("Token blacklist check cache hit",
				logger.String("token_hash", tokenHash),
				logger.Bool("is_blacklisted", result.IsBlacklisted))
			return result.IsBlacklisted, nil
		}
	}

	// Cache miss - check database
	isBlacklisted, err := s.baseService.IsTokenBlacklisted(ctx, token)
	if err != nil {
		return false, err
	}

	// Cache the result
	result := struct {
		IsBlacklisted bool      `json:"is_blacklisted"`
		CachedAt      time.Time `json:"cached_at"`
	}{
		IsBlacklisted: isBlacklisted,
		CachedAt:      time.Now(),
	}

	if data, marshalErr := json.Marshal(result); marshalErr == nil {
		// Use short TTL for blacklist checks to ensure security
		_ = s.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.ShortCacheTTL)
	}

	s.logger.Debug("Token blacklist check result cached",
		logger.String("token_hash", tokenHash),
		logger.Bool("is_blacklisted", isBlacklisted))

	return isBlacklisted, nil
}

// BlacklistAllUserTokens blacklists all user tokens and invalidates caches
func (s *CachedJWTBlacklistService) BlacklistAllUserTokens(ctx context.Context, userID uint, reason string, tokenExpiresAt time.Time) error {
	err := s.baseService.BlacklistAllUserTokens(ctx, userID, reason, tokenExpiresAt)
	if err != nil {
		return err
	}

	// Invalidate user blacklist cache
	s.invalidateUserBlacklistCache(ctx, userID)

	// Also clear all token blacklist caches since we can't target specific user tokens
	pattern := fmt.Sprintf("%stoken_blacklist:*", cache.CachePrefixAuth)
	if err := s.cacheManager.GetCache().DeleteByPattern(ctx, pattern); err != nil {
		s.logger.Error("Failed to invalidate token blacklist caches after user blacklist",
			logger.Uint("user_id", userID),
			logger.ErrorField(err))
	}

	s.logger.Warn("All user tokens blacklisted and caches cleared",
		logger.Uint("user_id", userID),
		logger.String("reason", reason))

	return nil
}

// IsUserTokensBlacklisted checks if user tokens are blacklisted with caching
func (s *CachedJWTBlacklistService) IsUserTokensBlacklisted(ctx context.Context, userID uint, tokenIssuedAt time.Time) (bool, error) {
	cacheKey := fmt.Sprintf("%suser_blacklist:%d", cache.CachePrefixAuth, userID)

	// Try cache first
	cached, err := s.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var result struct {
			IsBlacklisted bool      `json:"is_blacklisted"`
			BlacklistedAt time.Time `json:"blacklisted_at"`
			CachedAt      time.Time `json:"cached_at"`
		}
		if err := json.Unmarshal(cached, &result); err == nil {
			// Check if token was issued before blacklist time
			if result.IsBlacklisted && tokenIssuedAt.Before(result.BlacklistedAt) {
				s.logger.Debug("User blacklist check cache hit - blacklisted",
					logger.Uint("user_id", userID))
				return true, nil
			} else if !result.IsBlacklisted {
				s.logger.Debug("User blacklist check cache hit - not blacklisted",
					logger.Uint("user_id", userID))
				return false, nil
			}
		}
	}

	// Cache miss or token issued after cached blacklist time - check database
	isBlacklisted, err := s.baseService.IsUserTokensBlacklisted(ctx, userID, tokenIssuedAt)
	if err != nil {
		return false, err
	}

	// Cache the result with blacklist timestamp if applicable
	result := struct {
		IsBlacklisted bool      `json:"is_blacklisted"`
		BlacklistedAt time.Time `json:"blacklisted_at"`
		CachedAt      time.Time `json:"cached_at"`
	}{
		IsBlacklisted: isBlacklisted,
		BlacklistedAt: time.Now(), // This will be overridden if we need exact timestamp
		CachedAt:      time.Now(),
	}

	if data, marshalErr := json.Marshal(result); marshalErr == nil {
		// Use medium TTL for user blacklist checks
		_ = s.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.MediumCacheTTL)
	}

	s.logger.Debug("User blacklist check result cached",
		logger.Uint("user_id", userID),
		logger.Bool("is_blacklisted", isBlacklisted))

	return isBlacklisted, nil
}

// CleanupExpiredEntries cleans up expired entries and clears related caches
func (s *CachedJWTBlacklistService) CleanupExpiredEntries(ctx context.Context) error {
	err := s.baseService.CleanupExpiredEntries(ctx)
	if err != nil {
		return err
	}

	// Clear all blacklist caches since we don't know which entries were cleaned
	patterns := []string{
		fmt.Sprintf("%stoken_blacklist:*", cache.CachePrefixAuth),
		fmt.Sprintf("%suser_blacklist:*", cache.CachePrefixAuth),
	}

	for _, pattern := range patterns {
		if err := s.cacheManager.GetCache().DeleteByPattern(ctx, pattern); err != nil {
			s.logger.Error("Failed to clear blacklist cache after cleanup",
				logger.String("pattern", pattern),
				logger.ErrorField(err))
		}
	}

	s.logger.Info("Blacklist cleanup completed and caches cleared")
	return nil
}

// GetBlacklistStats gets blacklist statistics with caching
func (s *CachedJWTBlacklistService) GetBlacklistStats(ctx context.Context) (map[string]any, error) {
	cacheKey := fmt.Sprintf("%sstats:blacklist", cache.CachePrefixAuth)

	// Try cache first
	cached, err := s.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var stats map[string]any
		if err := json.Unmarshal(cached, &stats); err == nil {
			s.logger.Debug("Blacklist stats cache hit")
			return stats, nil
		}
	}

	// Cache miss - get stats from database
	stats, err := s.baseService.GetBlacklistStats(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the stats
	if data, marshalErr := json.Marshal(stats); marshalErr == nil {
		// Use medium TTL for stats
		_ = s.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.MediumCacheTTL)
	}

	s.logger.Debug("Blacklist stats cached")
	return stats, nil
}

// Cache invalidation helper methods

func (s *CachedJWTBlacklistService) invalidateTokenCache(ctx context.Context, tokenHash string) {
	cacheKey := fmt.Sprintf("%stoken_blacklist:%s", cache.CachePrefixAuth, tokenHash)
	if err := s.cacheManager.GetCache().Delete(ctx, cacheKey); err != nil {
		s.logger.Error("Failed to invalidate token blacklist cache",
			logger.String("token_hash", tokenHash),
			logger.ErrorField(err))
	}
}

func (s *CachedJWTBlacklistService) invalidateUserBlacklistCache(ctx context.Context, userID uint) {
	cacheKey := fmt.Sprintf("%suser_blacklist:%d", cache.CachePrefixAuth, userID)
	if err := s.cacheManager.GetCache().Delete(ctx, cacheKey); err != nil {
		s.logger.Error("Failed to invalidate user blacklist cache",
			logger.Uint("user_id", userID),
			logger.ErrorField(err))
	}
}

// hashToken creates a consistent hash of the token for caching
func (s *CachedJWTBlacklistService) hashToken(token string) string {
	if len(token) < 20 {
		return fmt.Sprintf("%x", token)
	}
	return fmt.Sprintf("%s...%s", token[:10], token[len(token)-10:])
}
