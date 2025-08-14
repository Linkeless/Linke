package handlers

import (
	"context"
	"sync"

	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	sharedDTO "linke/internal/shared/dto"
	"linke/internal/shared/logger"
)

// BatchUserLoader provides efficient batch loading of user data to solve N+1 query problems
type BatchUserLoader struct {
	userService userInterfaces.UserService
	cache       map[uint]*sharedDTO.UserBasicDTO
	mu          sync.RWMutex
	// Performance optimizations
	maxCacheSize int
	preloadIDs   []uint // IDs to preload on next batch operation
}

// NewBatchUserLoader creates a new batch user loader
func NewBatchUserLoader(userService userInterfaces.UserService) *BatchUserLoader {
	return &BatchUserLoader{
		userService:  userService,
		cache:        make(map[uint]*sharedDTO.UserBasicDTO),
		maxCacheSize: 1000, // Default max cache size
		preloadIDs:   make([]uint, 0),
	}
}

// LoadUsers loads multiple users by their IDs in a single batch operation
func (loader *BatchUserLoader) LoadUsers(ctx context.Context, userIDs []uint) error {
	if len(userIDs) == 0 {
		return nil
	}

	// Filter out already cached users
	var uncachedIDs []uint
	loader.mu.RLock()
	for _, id := range userIDs {
		if _, exists := loader.cache[id]; !exists {
			uncachedIDs = append(uncachedIDs, id)
		}
	}
	loader.mu.RUnlock()

	if len(uncachedIDs) == 0 {
		// Process preload IDs if no new IDs to load
		if len(loader.preloadIDs) > 0 {
			return loader.preloadUsers(ctx)
		}
		return nil
	}

	// Load uncached users
	users, err := loader.userService.GetUsersByIDs(ctx, uncachedIDs)
	if err != nil {
		logger.Error("Failed to batch load users",
			logger.Any("user_ids", uncachedIDs),
			logger.ErrorField(err))
		return err
	}

	// Convert to DTOs
	userDTOs := make([]*sharedDTO.UserBasicDTO, 0, len(users))
	for _, user := range users {
		userDTO := &sharedDTO.UserBasicDTO{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Provider: user.Provider,
			Status:   user.Status,
			Role:     user.Role,
		}
		userDTOs = append(userDTOs, userDTO)
	}

	// Cache the loaded users
	loader.mu.Lock()
	for _, userDTO := range userDTOs {
		loader.cache[userDTO.ID] = userDTO
	}
	loader.mu.Unlock()

	// Check if cache cleanup is needed
	if len(loader.cache) > loader.maxCacheSize {
		loader.EvictLRU()
	}

	logger.Debug("Batch loaded users",
		logger.Int("requested_count", len(userIDs)),
		logger.Int("loaded_count", len(userDTOs)),
		logger.Int("cached_count", len(userIDs)-len(uncachedIDs)))

	return nil
}

// GetUser retrieves a user from the cache. Must call LoadUsers first.
func (loader *BatchUserLoader) GetUser(userID uint) *sharedDTO.UserBasicDTO {
	loader.mu.RLock()
	defer loader.mu.RUnlock()
	return loader.cache[userID]
}

// GetUsers retrieves multiple users from the cache
func (loader *BatchUserLoader) GetUsers(userIDs []uint) map[uint]*sharedDTO.UserBasicDTO {
	loader.mu.RLock()
	defer loader.mu.RUnlock()

	result := make(map[uint]*sharedDTO.UserBasicDTO, len(userIDs))
	for _, userID := range userIDs {
		if user, exists := loader.cache[userID]; exists {
			result[userID] = user
		}
	}
	return result
}

// HasUser checks if a user is loaded in the cache
func (loader *BatchUserLoader) HasUser(userID uint) bool {
	loader.mu.RLock()
	defer loader.mu.RUnlock()
	_, exists := loader.cache[userID]
	return exists
}

// Clear clears the cache
func (loader *BatchUserLoader) Clear() {
	loader.mu.Lock()
	defer loader.mu.Unlock()
	loader.cache = make(map[uint]*sharedDTO.UserBasicDTO)
}

// CacheSize returns the number of cached users
func (loader *BatchUserLoader) CacheSize() int {
	loader.mu.RLock()
	defer loader.mu.RUnlock()
	return len(loader.cache)
}

// CollectUserIDs is a utility function to collect unique user IDs from various sources
type UserIDCollector struct {
	userIDs map[uint]bool
}

// NewUserIDCollector creates a new user ID collector
func NewUserIDCollector() *UserIDCollector {
	return &UserIDCollector{
		userIDs: make(map[uint]bool),
	}
}

// Add adds a user ID to the collection
func (collector *UserIDCollector) Add(userID uint) {
	if userID != 0 {
		collector.userIDs[userID] = true
	}
}

// AddPtr adds a user ID from a pointer (handles nil pointers)
func (collector *UserIDCollector) AddPtr(userID *uint) {
	if userID != nil && *userID != 0 {
		collector.userIDs[*userID] = true
	}
}

// ToSlice returns collected user IDs as a slice
func (collector *UserIDCollector) ToSlice() []uint {
	result := make([]uint, 0, len(collector.userIDs))
	for userID := range collector.userIDs {
		result = append(result, userID)
	}
	return result
}

// Count returns the number of unique user IDs collected
func (collector *UserIDCollector) Count() int {
	return len(collector.userIDs)
}

// preloadUsers processes the preload queue
func (loader *BatchUserLoader) preloadUsers(ctx context.Context) error {
	loader.mu.Lock()
	preloadIDs := make([]uint, len(loader.preloadIDs))
	copy(preloadIDs, loader.preloadIDs)
	loader.preloadIDs = loader.preloadIDs[:0] // Clear preload queue
	loader.mu.Unlock()

	if len(preloadIDs) == 0 {
		return nil
	}

	// Filter out already cached users
	var uncachedPreloadIDs []uint
	loader.mu.RLock()
	for _, id := range preloadIDs {
		if _, exists := loader.cache[id]; !exists {
			uncachedPreloadIDs = append(uncachedPreloadIDs, id)
		}
	}
	loader.mu.RUnlock()

	if len(uncachedPreloadIDs) == 0 {
		return nil
	}

	// Load preload users
	users, err := loader.userService.GetUsersByIDs(ctx, uncachedPreloadIDs)
	if err != nil {
		logger.Warn("Failed to preload users",
			logger.Any("user_ids", uncachedPreloadIDs),
			logger.ErrorField(err))
		return err
	}

	// Convert to DTOs and cache
	loader.mu.Lock()
	for _, user := range users {
		userDTO := &sharedDTO.UserBasicDTO{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Provider: user.Provider,
			Status:   user.Status,
			Role:     user.Role,
		}
		loader.cache[userDTO.ID] = userDTO
	}
	loader.mu.Unlock()

	logger.Debug("Preloaded users",
		logger.Int("preload_count", len(users)))

	return nil
}

// AddPreloadIDs adds user IDs to the preload queue for the next batch operation
func (loader *BatchUserLoader) AddPreloadIDs(userIDs []uint) {
	loader.mu.Lock()
	defer loader.mu.Unlock()
	
	for _, userID := range userIDs {
		if userID != 0 {
			// Avoid duplicates
			alreadyExists := false
			for _, existingID := range loader.preloadIDs {
				if existingID == userID {
					alreadyExists = true
					break
				}
			}
			if !alreadyExists {
				loader.preloadIDs = append(loader.preloadIDs, userID)
			}
		}
	}
}

// EvictLRU evicts least recently used items if cache exceeds max size
func (loader *BatchUserLoader) EvictLRU() {
	loader.mu.Lock()
	defer loader.mu.Unlock()
	
	if len(loader.cache) <= loader.maxCacheSize {
		return
	}
	
	// Simple LRU: remove random items (in production, you'd want proper LRU)
	count := 0
	targetRemove := len(loader.cache) - loader.maxCacheSize + 100 // Remove extra to avoid frequent eviction
	for userID := range loader.cache {
		if count >= targetRemove {
			break
		}
		delete(loader.cache, userID)
		count++
	}
	
	logger.Debug("Evicted cache entries",
		logger.Int("evicted_count", count),
		logger.Int("remaining_count", len(loader.cache)))
}

// SetMaxCacheSize sets the maximum cache size
func (loader *BatchUserLoader) SetMaxCacheSize(size int) {
	loader.mu.Lock()
	defer loader.mu.Unlock()
	loader.maxCacheSize = size
}