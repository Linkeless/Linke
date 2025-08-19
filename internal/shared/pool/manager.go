package pool

import (
	"fmt"
	"sync"
)

// PoolManager manages multiple object pools with type safety
type PoolManager struct {
	pools map[string]interface{}
	mutex sync.RWMutex
}

// NewPoolManager creates a new pool manager
func NewPoolManager() *PoolManager {
	return &PoolManager{
		pools: make(map[string]interface{}),
	}
}

// RegisterPool registers a typed pool with the manager
func (pm *PoolManager) RegisterPool(name string, pool interface{}) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, exists := pm.pools[name]; exists {
		return fmt.Errorf("pool %s already registered", name)
	}

	pm.pools[name] = pool
	return nil
}

// GetPool retrieves a registered pool by name
func (pm *PoolManager) GetPool(name string) (interface{}, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	pool, exists := pm.pools[name]
	if !exists {
		return nil, fmt.Errorf("pool %s not found", name)
	}

	return pool, nil
}

// GetTypedPool retrieves a typed pool by name with type assertion
func GetTypedPool[T any](pm *PoolManager, name string) (*Pool[T], error) {
	pool, err := pm.GetPool(name)
	if err != nil {
		return nil, err
	}

	typedPool, ok := pool.(*Pool[T])
	if !ok {
		return nil, fmt.Errorf("pool %s is not of the expected type", name)
	}

	return typedPool, nil
}

// ListPools returns all registered pool names
func (pm *PoolManager) ListPools() []string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	names := make([]string, 0, len(pm.pools))
	for name := range pm.pools {
		names = append(names, name)
	}

	return names
}

// UnregisterPool removes a pool from the manager
func (pm *PoolManager) UnregisterPool(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, exists := pm.pools[name]; !exists {
		return fmt.Errorf("pool %s not found", name)
	}

	delete(pm.pools, name)
	return nil
}

// PoolCount returns the number of registered pools
func (pm *PoolManager) PoolCount() int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return len(pm.pools)
}

// Global pool manager instance
var defaultManager = NewPoolManager()

// GetDefaultManager returns the default global pool manager
func GetDefaultManager() *PoolManager {
	return defaultManager
}

// RegisterPoolGlobal registers a pool in the default global manager
func RegisterPoolGlobal(name string, pool interface{}) error {
	return defaultManager.RegisterPool(name, pool)
}

// GetPoolGlobal retrieves a pool from the default global manager
func GetPoolGlobal(name string) (interface{}, error) {
	return defaultManager.GetPool(name)
}

// GetTypedPoolGlobal retrieves a typed pool from the default global manager
func GetTypedPoolGlobal[T any](name string) (*Pool[T], error) {
	return GetTypedPool[T](defaultManager, name)
}
