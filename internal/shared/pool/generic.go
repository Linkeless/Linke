package pool

import (
	"sync"
	"sync/atomic"
)

// Pool is a generic object pool with statistics tracking
type Pool[T any] struct {
	pool      sync.Pool
	newFunc   func() *T
	resetFunc func(*T)
	stats     *PoolStats
}

// PoolStats tracks pool usage statistics
type PoolStats struct {
	Gets   atomic.Int64
	Puts   atomic.Int64
	Misses atomic.Int64
	Hits   atomic.Int64
}

// NewPool creates a new generic object pool
func NewPool[T any](newFunc func() *T, resetFunc func(*T)) *Pool[T] {
	p := &Pool[T]{
		newFunc:   newFunc,
		resetFunc: resetFunc,
		stats:     &PoolStats{},
	}

	p.pool = sync.Pool{
		New: func() any {
			p.stats.Misses.Add(1)
			return newFunc()
		},
	}

	return p
}

// Get retrieves an object from the pool
func (p *Pool[T]) Get() *T {
	p.stats.Gets.Add(1)
	obj := p.pool.Get().(*T)

	// Track if this was a hit (object was reused from pool)
	// If the object is not newly created, it's a hit
	if obj != nil {
		p.stats.Hits.Add(1)
	}

	return obj
}

// Put returns an object to the pool after resetting it
func (p *Pool[T]) Put(obj *T) {
	if obj == nil {
		return
	}

	p.stats.Puts.Add(1)

	// Reset the object if reset function is provided
	if p.resetFunc != nil {
		p.resetFunc(obj)
	}

	p.pool.Put(obj)
}

// Stats returns the current pool statistics
func (p *Pool[T]) Stats() PoolStatsSnapshot {
	return PoolStatsSnapshot{
		Gets:   p.stats.Gets.Load(),
		Puts:   p.stats.Puts.Load(),
		Misses: p.stats.Misses.Load(),
		Hits:   p.stats.Hits.Load(),
	}
}

// PoolStatsSnapshot is a snapshot of pool statistics
type PoolStatsSnapshot struct {
	Gets   int64
	Puts   int64
	Misses int64
	Hits   int64
}

// HitRate calculates the hit rate percentage
func (s PoolStatsSnapshot) HitRate() float64 {
	if s.Gets == 0 {
		return 0
	}
	return float64(s.Hits) / float64(s.Gets) * 100
}

// ReuseRate calculates the reuse rate percentage
func (s PoolStatsSnapshot) ReuseRate() float64 {
	total := s.Gets
	if total == 0 {
		return 0
	}
	return float64(total-s.Misses) / float64(total) * 100
}
