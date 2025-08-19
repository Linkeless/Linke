package implementations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync/atomic"

	"linke/internal/domains/auth/dto"
	"linke/internal/domains/auth/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/cache"
	"linke/internal/shared/framework"
	"linke/internal/shared/logger"
)

// CachedJWTService wraps JWTService with caching for frequently accessed operations
type CachedJWTService struct {
	baseService   *JWTService
	cacheManager  cache.CacheManager
	cacheKeys     *cache.AllCacheKeys
	logger        framework.Logger
	cacheStrategy *CacheStrategyManager

	// Atomic counters for performance metrics
	cacheHits     int64
	cacheMisses   int64
	validationOps int64
	errorCount    int64
}

// NewCachedJWTService creates a new cached JWT service
func NewCachedJWTService(
	baseService *JWTService,
	cacheManager cache.CacheManager,
	cacheKeys *cache.AllCacheKeys,
	logger framework.Logger,
) interfaces.JWTService {
	return &CachedJWTService{
		baseService:   baseService,
		cacheManager:  cacheManager,
		cacheKeys:     cacheKeys,
		logger:        logger,
		cacheStrategy: NewCacheStrategyManager(),
	}
}

// GenerateToken generates JWT token (no caching needed as each token is unique)
func (s *CachedJWTService) GenerateToken(user *userEntities.User) (*dto.TokenResponse, error) {
	return s.baseService.GenerateToken(user)
}

// ValidateToken validates JWT token with caching to reduce computation overhead
func (s *CachedJWTService) ValidateToken(tokenString string) (*dto.Claims, error) {
	// Increment validation operation counter
	atomic.AddInt64(&s.validationOps, 1)

	// Generate optimized cache key using strategy manager
	cacheKey := s.cacheStrategy.GenerateKey(TokenValidationKey, tokenString)
	cacheTTL := s.cacheStrategy.GetTTL(TokenValidationKey)

	// Try to get cached validation result
	cached, err := s.cacheManager.GetCache().Get(context.Background(), cacheKey)
	if err == nil && cached != nil {
		var cachedResult struct {
			Claims *dto.Claims `json:"claims"`
			Valid  bool        `json:"valid"`
		}
		if err := json.Unmarshal(cached, &cachedResult); err == nil && cachedResult.Valid {
			atomic.AddInt64(&s.cacheHits, 1)
			tokenHash := s.hashToken(tokenString)
			s.logger.Debug("JWT validation cache hit", logger.String("token_hash", tokenHash))
			return cachedResult.Claims, nil
		}
	}

	// Cache miss - increment counter
	atomic.AddInt64(&s.cacheMisses, 1)

	// Validate token
	claims, err := s.baseService.ValidateToken(tokenString)

	// Cache the result only if validation is successful
	if err == nil && claims != nil {
		result := struct {
			Claims *dto.Claims `json:"claims"`
			Valid  bool        `json:"valid"`
		}{
			Claims: claims,
			Valid:  true,
		}

		if data, marshalErr := json.Marshal(result); marshalErr == nil {
			// Use strategy-defined TTL for security-appropriate caching
			_ = s.cacheManager.GetCache().Set(context.Background(), cacheKey, data, cacheTTL)
		}

		tokenHash := s.hashToken(tokenString)
		s.logger.Debug("JWT validation result cached",
			logger.String("token_hash", tokenHash),
			logger.Uint("user_id", claims.UserID),
			logger.Duration("ttl", cacheTTL))
	} else if err != nil {
		atomic.AddInt64(&s.errorCount, 1)
	}

	return claims, err
}

// RefreshToken generates new token (no caching needed)
func (s *CachedJWTService) RefreshToken(tokenString string) (*dto.TokenResponse, error) {
	// Invalidate validation cache for old token using strategy pattern
	oldTokenKey := s.cacheStrategy.GenerateKey(TokenValidationKey, tokenString)
	_ = s.cacheManager.GetCache().Delete(context.Background(), oldTokenKey)

	return s.baseService.RefreshToken(tokenString)
}

// RevokeToken revokes token and invalidates related caches
func (s *CachedJWTService) RevokeToken(tokenString string, userID *uint, reason string) error {
	err := s.baseService.RevokeToken(tokenString, userID, reason)
	if err != nil {
		return err
	}

	// Invalidate validation cache for this token using strategy pattern
	tokenKey := s.cacheStrategy.GenerateKey(TokenValidationKey, tokenString)
	if err := s.cacheManager.GetCache().Delete(context.Background(), tokenKey); err != nil {
		tokenHash := s.hashToken(tokenString)
		s.logger.Error("Failed to invalidate token validation cache",
			logger.String("token_hash", tokenHash),
			logger.ErrorField(err))
	}

	tokenHash := s.hashToken(tokenString)
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

	// Invalidate all validation caches using strategy pattern
	pattern := s.cacheStrategy.GetKeyPattern(TokenValidationKey)
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

// RotateSecret rotates the JWT signing secret (delegates to base service)
func (s *CachedJWTService) RotateSecret(newSecret string) {
	s.baseService.RotateSecret(newSecret)

	// Clear all validation caches when secret is rotated for security using strategy pattern
	pattern := s.cacheStrategy.GetKeyPattern(TokenValidationKey)
	if err := s.cacheManager.GetCache().DeleteByPattern(context.Background(), pattern); err != nil {
		s.logger.Error("Failed to clear validation caches after secret rotation",
			logger.ErrorField(err))
	}

	s.logger.Warn("JWT secret rotated and validation caches cleared")
}

// GetMetrics returns current performance metrics using atomic loads
func (s *CachedJWTService) GetMetrics() map[string]int64 {
	return map[string]int64{
		"cache_hits":     atomic.LoadInt64(&s.cacheHits),
		"cache_misses":   atomic.LoadInt64(&s.cacheMisses),
		"validation_ops": atomic.LoadInt64(&s.validationOps),
		"error_count":    atomic.LoadInt64(&s.errorCount),
	}
}

// GetCacheHitRatio returns the cache hit ratio as a percentage
func (s *CachedJWTService) GetCacheHitRatio() float64 {
	hits := atomic.LoadInt64(&s.cacheHits)
	total := hits + atomic.LoadInt64(&s.cacheMisses)
	if total == 0 {
		return 0.0
	}
	return float64(hits) / float64(total) * 100.0
}

// hashToken creates a secure SHA256 hash of the token for caching
// This provides better security than simple string truncation
func (s *CachedJWTService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
