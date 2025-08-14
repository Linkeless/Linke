package implementations

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"linke/internal/shared/cache"
)

// CacheKeyType represents different types of cache keys
type CacheKeyType int

const (
	TokenValidationKey CacheKeyType = iota
	UserDataKey
	BlacklistKey
	ConfigKey
)

// CacheStrategy defines TTL and key generation strategy for different cache types
type CacheStrategy struct {
	KeyType    CacheKeyType
	TTL        time.Duration
	Prefix     string
	UseShortID bool // Use short hash for performance
}

// CacheStrategyManager manages cache key generation and TTL strategies
type CacheStrategyManager struct {
	strategies map[CacheKeyType]*CacheStrategy
}

// NewCacheStrategyManager creates a new cache strategy manager with optimized TTL layers
func NewCacheStrategyManager() *CacheStrategyManager {
	strategies := make(map[CacheKeyType]*CacheStrategy)
	
	// Define layered TTL strategies based on data sensitivity and usage patterns
	strategies[TokenValidationKey] = &CacheStrategy{
		KeyType:    TokenValidationKey,
		TTL:        cache.ShortCacheTTL, // 1 minute - security sensitive
		Prefix:     cache.CachePrefixAuth + "val",
		UseShortID: true, // Use shorter hash for performance
	}
	
	strategies[UserDataKey] = &CacheStrategy{
		KeyType:    UserDataKey,
		TTL:        cache.MediumCacheTTL, // 15 minutes - moderately dynamic
		Prefix:     cache.CachePrefixAuth + "user",
		UseShortID: false, // Full hash for uniqueness
	}
	
	strategies[BlacklistKey] = &CacheStrategy{
		KeyType:    BlacklistKey,
		TTL:        cache.LongCacheTTL, // 1 hour - relatively static
		Prefix:     cache.CachePrefixAuth + "bl",
		UseShortID: true, // Performance over uniqueness
	}
	
	strategies[ConfigKey] = &CacheStrategy{
		KeyType:    ConfigKey,
		TTL:        time.Hour * 6, // 6 hours - very static
		Prefix:     cache.CachePrefixAuth + "cfg",
		UseShortID: false,
	}
	
	return &CacheStrategyManager{
		strategies: strategies,
	}
}

// GenerateKey generates an optimized cache key based on the strategy
func (csm *CacheStrategyManager) GenerateKey(keyType CacheKeyType, data string) string {
	strategy, exists := csm.strategies[keyType]
	if !exists {
		// Fallback to default strategy
		return fmt.Sprintf("%s:default:%s", cache.CachePrefixAuth, data)
	}
	
	// Use different hash strategies based on performance vs uniqueness needs
	var hashedData string
	if strategy.UseShortID {
		// Use first 16 chars of SHA256 for performance
		hash := sha256.Sum256([]byte(data))
		hashedData = hex.EncodeToString(hash[:8]) // 16 chars
	} else {
		// Use full SHA256 for maximum uniqueness
		hash := sha256.Sum256([]byte(data))
		hashedData = hex.EncodeToString(hash[:])
	}
	
	return fmt.Sprintf("%s:%s", strategy.Prefix, hashedData)
}

// GenerateCompositeKey generates a cache key from multiple data elements
func (csm *CacheStrategyManager) GenerateCompositeKey(keyType CacheKeyType, elements ...string) string {
	data := strings.Join(elements, "|")
	return csm.GenerateKey(keyType, data)
}

// GetTTL returns the TTL for a specific cache key type
func (csm *CacheStrategyManager) GetTTL(keyType CacheKeyType) time.Duration {
	strategy, exists := csm.strategies[keyType]
	if !exists {
		return cache.ShortCacheTTL // Default fallback
	}
	return strategy.TTL
}

// GetStrategy returns the complete strategy for a cache key type
func (csm *CacheStrategyManager) GetStrategy(keyType CacheKeyType) *CacheStrategy {
	return csm.strategies[keyType]
}

// UpdateTTL allows dynamic TTL adjustment for specific cache types
func (csm *CacheStrategyManager) UpdateTTL(keyType CacheKeyType, newTTL time.Duration) {
	if strategy, exists := csm.strategies[keyType]; exists {
		strategy.TTL = newTTL
	}
}

// GetKeyPattern returns a pattern for bulk cache operations
func (csm *CacheStrategyManager) GetKeyPattern(keyType CacheKeyType) string {
	strategy, exists := csm.strategies[keyType]
	if !exists {
		return cache.CachePrefixAuth + "*"
	}
	return strategy.Prefix + ":*"
}