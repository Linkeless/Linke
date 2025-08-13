package implementations

import (
	"context"
	"encoding/json"
	"fmt"

	"linke/internal/domains/auth/dto"
	"linke/internal/domains/auth/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/cache"
	"linke/internal/shared/framework"
	"linke/internal/shared/logger"
)

// CachedJWTService wraps JWTService with caching for frequently accessed operations
type CachedJWTService struct {
	baseService  *JWTService
	cacheManager cache.CacheManager
	cacheKeys    *cache.AllCacheKeys
	logger       framework.Logger
}

// NewCachedJWTService creates a new cached JWT service
func NewCachedJWTService(
	baseService *JWTService,
	cacheManager cache.CacheManager,
	cacheKeys *cache.AllCacheKeys,
	logger framework.Logger,
) interfaces.JWTService {
	return &CachedJWTService{
		baseService:  baseService,
		cacheManager: cacheManager,
		cacheKeys:    cacheKeys,
		logger:       logger,
	}
}

// GenerateToken generates JWT token (no caching needed as each token is unique)
func (s *CachedJWTService) GenerateToken(user *userEntities.User) (*dto.TokenResponse, error) {
	return s.baseService.GenerateToken(user)
}

// ValidateToken validates JWT token with caching to reduce computation overhead
func (s *CachedJWTService) ValidateToken(tokenString string) (*dto.Claims, error) {
	// Create cache key based on token hash to avoid storing full token
	tokenHash := s.hashToken(tokenString)
	cacheKey := fmt.Sprintf("%svalidation:%s", cache.CachePrefixAuth, tokenHash)

	// Try to get cached validation result
	cached, err := s.cacheManager.GetCache().Get(context.Background(), cacheKey)
	if err == nil && cached != nil {
		var cachedResult struct {
			Claims *dto.Claims `json:"claims"`
			Valid  bool               `json:"valid"`
		}
		if err := json.Unmarshal(cached, &cachedResult); err == nil && cachedResult.Valid {
			s.logger.Debug("JWT validation cache hit", logger.String("token_hash", tokenHash))
			return cachedResult.Claims, nil
		}
	}

	// Cache miss - validate token
	claims, err := s.baseService.ValidateToken(tokenString)
	
	// Cache the result only if validation is successful
	// We use short TTL to balance performance with security (token might be revoked)
	if err == nil && claims != nil {
		result := struct {
			Claims *dto.Claims `json:"claims"`
			Valid  bool               `json:"valid"`
		}{
			Claims: claims,
			Valid:  true,
		}
		
		if data, marshalErr := json.Marshal(result); marshalErr == nil {
			// Use short cache TTL for security reasons
			_ = s.cacheManager.GetCache().Set(context.Background(), cacheKey, data, cache.ShortCacheTTL)
		}
		
		s.logger.Debug("JWT validation result cached", 
			logger.String("token_hash", tokenHash),
			logger.Uint("user_id", claims.UserID))
	}

	return claims, err
}

// RefreshToken generates new token (no caching needed)
func (s *CachedJWTService) RefreshToken(tokenString string) (*dto.TokenResponse, error) {
	// Invalidate validation cache for old token
	tokenHash := s.hashToken(tokenString)
	cacheKey := fmt.Sprintf("%svalidation:%s", cache.CachePrefixAuth, tokenHash)
	_ = s.cacheManager.GetCache().Delete(context.Background(), cacheKey)
	
	return s.baseService.RefreshToken(tokenString)
}

// RevokeToken revokes token and invalidates related caches
func (s *CachedJWTService) RevokeToken(tokenString string, userID *uint, reason string) error {
	err := s.baseService.RevokeToken(tokenString, userID, reason)
	if err != nil {
		return err
	}

	// Invalidate validation cache for this token
	tokenHash := s.hashToken(tokenString)
	cacheKey := fmt.Sprintf("%svalidation:%s", cache.CachePrefixAuth, tokenHash)
	if err := s.cacheManager.GetCache().Delete(context.Background(), cacheKey); err != nil {
		s.logger.Error("Failed to invalidate token validation cache",
			logger.String("token_hash", tokenHash),
			logger.ErrorField(err))
	}

	s.logger.Info("Token revoked and cache invalidated",
		logger.String("token_hash", tokenHash),
		logger.String("reason", reason))

	return nil
}

// RevokeAllUserTokens revokes all user tokens and invalidates related caches
func (s *CachedJWTService) RevokeAllUserTokens(userID uint, reason string) error {
	err := s.baseService.RevokeAllUserTokens(userID, reason)
	if err != nil {
		return err
	}

	// Invalidate all validation caches (we can't target specific tokens)
	// This is a security measure when all user tokens are revoked
	pattern := fmt.Sprintf("%svalidation:*", cache.CachePrefixAuth)
	if err := s.cacheManager.GetCache().DeleteByPattern(context.Background(), pattern); err != nil {
		s.logger.Error("Failed to invalidate token validation caches after user token revocation",
			logger.Uint("user_id", userID),
			logger.ErrorField(err))
	}

	s.logger.Warn("All user tokens revoked and validation caches cleared",
		logger.Uint("user_id", userID),
		logger.String("reason", reason))

	return nil
}

// hashToken creates a consistent hash of the token for caching
// This avoids storing the actual token in cache for security
func (s *CachedJWTService) hashToken(token string) string {
	if len(token) < 20 {
		return fmt.Sprintf("%x", token) // Fallback for very short tokens
	}
	
	// Use first 10 and last 10 characters as a simple hash
	// This provides reasonable uniqueness without storing the full token
	return fmt.Sprintf("%s...%s", token[:10], token[len(token)-10:])
}