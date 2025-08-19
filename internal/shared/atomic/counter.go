package atomic

import (
	"encoding/json"
	"sync/atomic"
	"time"
)

// Counter represents an atomic counter with additional functionality
type Counter struct {
	value     int64
	createdAt time.Time
	name      string
}

// NewCounter creates a new atomic counter
func NewCounter(name string) *Counter {
	return &Counter{
		value:     0,
		createdAt: time.Now(),
		name:      name,
	}
}

// NewCounterWithValue creates a new atomic counter with an initial value
func NewCounterWithValue(name string, initialValue int64) *Counter {
	return &Counter{
		value:     initialValue,
		createdAt: time.Now(),
		name:      name,
	}
}

// Increment atomically increments the counter by 1 and returns the new value
func (c *Counter) Increment() int64 {
	return atomic.AddInt64(&c.value, 1)
}

// IncrementBy atomically increments the counter by delta and returns the new value
func (c *Counter) IncrementBy(delta int64) int64 {
	return atomic.AddInt64(&c.value, delta)
}

// Decrement atomically decrements the counter by 1 and returns the new value
func (c *Counter) Decrement() int64 {
	return atomic.AddInt64(&c.value, -1)
}

// DecrementBy atomically decrements the counter by delta and returns the new value
func (c *Counter) DecrementBy(delta int64) int64 {
	return atomic.AddInt64(&c.value, -delta)
}

// Load atomically loads and returns the current value
func (c *Counter) Load() int64 {
	return atomic.LoadInt64(&c.value)
}

// Store atomically stores a new value
func (c *Counter) Store(value int64) {
	atomic.StoreInt64(&c.value, value)
}

// Swap atomically stores new value and returns the old value
func (c *Counter) Swap(new int64) int64 {
	return atomic.SwapInt64(&c.value, new)
}

// CompareAndSwap atomically compares and swaps the value
func (c *Counter) CompareAndSwap(old, new int64) bool {
	return atomic.CompareAndSwapInt64(&c.value, old, new)
}

// Reset atomically resets the counter to 0 and returns the old value
func (c *Counter) Reset() int64 {
	return atomic.SwapInt64(&c.value, 0)
}

// GetName returns the counter name
func (c *Counter) GetName() string {
	return c.name
}

// GetCreatedAt returns when the counter was created
func (c *Counter) GetCreatedAt() time.Time {
	return c.createdAt
}

// GetAge returns how long ago the counter was created
func (c *Counter) GetAge() time.Duration {
	return time.Since(c.createdAt)
}

// Stats returns counter statistics
func (c *Counter) Stats() CounterStats {
	return CounterStats{
		Name:      c.name,
		Value:     c.Load(),
		CreatedAt: c.createdAt,
		Age:       c.GetAge(),
	}
}

// CounterStats represents counter statistics
type CounterStats struct {
	Name      string        `json:"name"`
	Value     int64         `json:"value"`
	CreatedAt time.Time     `json:"created_at"`
	Age       time.Duration `json:"age"`
}

// Int32Counter represents a 32-bit atomic counter
type Int32Counter struct {
	value     int32
	createdAt time.Time
	name      string
}

// NewInt32Counter creates a new 32-bit atomic counter
func NewInt32Counter(name string) *Int32Counter {
	return &Int32Counter{
		value:     0,
		createdAt: time.Now(),
		name:      name,
	}
}

// NewInt32CounterWithValue creates a new 32-bit atomic counter with an initial value
func NewInt32CounterWithValue(name string, initialValue int32) *Int32Counter {
	return &Int32Counter{
		value:     initialValue,
		createdAt: time.Now(),
		name:      name,
	}
}

// Increment atomically increments the counter by 1 and returns the new value
func (c *Int32Counter) Increment() int32 {
	return atomic.AddInt32(&c.value, 1)
}

// IncrementBy atomically increments the counter by delta and returns the new value
func (c *Int32Counter) IncrementBy(delta int32) int32 {
	return atomic.AddInt32(&c.value, delta)
}

// Decrement atomically decrements the counter by 1 and returns the new value
func (c *Int32Counter) Decrement() int32 {
	return atomic.AddInt32(&c.value, -1)
}

// DecrementBy atomically decrements the counter by delta and returns the new value
func (c *Int32Counter) DecrementBy(delta int32) int32 {
	return atomic.AddInt32(&c.value, -delta)
}

// Load atomically loads and returns the current value
func (c *Int32Counter) Load() int32 {
	return atomic.LoadInt32(&c.value)
}

// Store atomically stores a new value
func (c *Int32Counter) Store(value int32) {
	atomic.StoreInt32(&c.value, value)
}

// Swap atomically stores new value and returns the old value
func (c *Int32Counter) Swap(new int32) int32 {
	return atomic.SwapInt32(&c.value, new)
}

// CompareAndSwap atomically compares and swaps the value
func (c *Int32Counter) CompareAndSwap(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&c.value, old, new)
}

// Reset atomically resets the counter to 0 and returns the old value
func (c *Int32Counter) Reset() int32 {
	return atomic.SwapInt32(&c.value, 0)
}

// GetName returns the counter name
func (c *Int32Counter) GetName() string {
	return c.name
}

// GetCreatedAt returns when the counter was created
func (c *Int32Counter) GetCreatedAt() time.Time {
	return c.createdAt
}

// GetAge returns how long ago the counter was created
func (c *Int32Counter) GetAge() time.Duration {
	return time.Since(c.createdAt)
}

// Stats returns counter statistics
func (c *Int32Counter) Stats() Int32CounterStats {
	return Int32CounterStats{
		Name:      c.name,
		Value:     c.Load(),
		CreatedAt: c.createdAt,
		Age:       c.GetAge(),
	}
}

// Int32CounterStats represents 32-bit counter statistics
type Int32CounterStats struct {
	Name      string        `json:"name"`
	Value     int32         `json:"value"`
	CreatedAt time.Time     `json:"created_at"`
	Age       time.Duration `json:"age"`
}

// AtomicBool represents an atomic boolean
type AtomicBool struct {
	value     int32
	createdAt time.Time
	name      string
}

// NewAtomicBool creates a new atomic boolean
func NewAtomicBool(name string) *AtomicBool {
	return &AtomicBool{
		value:     0,
		createdAt: time.Now(),
		name:      name,
	}
}

// NewAtomicBoolWithValue creates a new atomic boolean with an initial value
func NewAtomicBoolWithValue(name string, initialValue bool) *AtomicBool {
	var value int32
	if initialValue {
		value = 1
	}
	return &AtomicBool{
		value:     value,
		createdAt: time.Now(),
		name:      name,
	}
}

// Load atomically loads and returns the current boolean value
func (ab *AtomicBool) Load() bool {
	return atomic.LoadInt32(&ab.value) == 1
}

// Store atomically stores a new boolean value
func (ab *AtomicBool) Store(value bool) {
	var intValue int32
	if value {
		intValue = 1
	}
	atomic.StoreInt32(&ab.value, intValue)
}

// Swap atomically stores new boolean value and returns the old value
func (ab *AtomicBool) Swap(new bool) bool {
	var newValue int32
	if new {
		newValue = 1
	}
	oldValue := atomic.SwapInt32(&ab.value, newValue)
	return oldValue == 1
}

// CompareAndSwap atomically compares and swaps the boolean value
func (ab *AtomicBool) CompareAndSwap(old, new bool) bool {
	var oldValue, newValue int32
	if old {
		oldValue = 1
	}
	if new {
		newValue = 1
	}
	return atomic.CompareAndSwapInt32(&ab.value, oldValue, newValue)
}

// Toggle atomically toggles the boolean value and returns the new value
func (ab *AtomicBool) Toggle() bool {
	for {
		current := atomic.LoadInt32(&ab.value)
		newValue := 1 - current
		if atomic.CompareAndSwapInt32(&ab.value, current, newValue) {
			return newValue == 1
		}
	}
}

// SetTrue atomically sets the value to true and returns whether it was changed
func (ab *AtomicBool) SetTrue() bool {
	return atomic.SwapInt32(&ab.value, 1) == 0
}

// SetFalse atomically sets the value to false and returns whether it was changed
func (ab *AtomicBool) SetFalse() bool {
	return atomic.SwapInt32(&ab.value, 0) == 1
}

// GetName returns the atomic boolean name
func (ab *AtomicBool) GetName() string {
	return ab.name
}

// GetCreatedAt returns when the atomic boolean was created
func (ab *AtomicBool) GetCreatedAt() time.Time {
	return ab.createdAt
}

// Stats returns atomic boolean statistics
func (ab *AtomicBool) Stats() AtomicBoolStats {
	return AtomicBoolStats{
		Name:      ab.name,
		Value:     ab.Load(),
		CreatedAt: ab.createdAt,
		Age:       time.Since(ab.createdAt),
	}
}

// AtomicBoolStats represents atomic boolean statistics
type AtomicBoolStats struct {
	Name      string        `json:"name"`
	Value     bool          `json:"value"`
	CreatedAt time.Time     `json:"created_at"`
	Age       time.Duration `json:"age"`
}

// CounterGroup manages multiple related counters
type CounterGroup struct {
	counters   map[string]*Counter
	int32s     map[string]*Int32Counter
	bools      map[string]*AtomicBool
	createdAt  time.Time
	groupName  string
}

// NewCounterGroup creates a new counter group
func NewCounterGroup(groupName string) *CounterGroup {
	return &CounterGroup{
		counters:   make(map[string]*Counter),
		int32s:     make(map[string]*Int32Counter),
		bools:      make(map[string]*AtomicBool),
		createdAt:  time.Now(),
		groupName:  groupName,
	}
}

// AddCounter adds a new counter to the group
func (cg *CounterGroup) AddCounter(name string) *Counter {
	counter := NewCounter(name)
	cg.counters[name] = counter
	return counter
}

// AddInt32Counter adds a new int32 counter to the group
func (cg *CounterGroup) AddInt32Counter(name string) *Int32Counter {
	counter := NewInt32Counter(name)
	cg.int32s[name] = counter
	return counter
}

// AddAtomicBool adds a new atomic boolean to the group
func (cg *CounterGroup) AddAtomicBool(name string) *AtomicBool {
	atomicBool := NewAtomicBool(name)
	cg.bools[name] = atomicBool
	return atomicBool
}

// GetCounter retrieves a counter by name
func (cg *CounterGroup) GetCounter(name string) (*Counter, bool) {
	counter, exists := cg.counters[name]
	return counter, exists
}

// GetInt32Counter retrieves an int32 counter by name
func (cg *CounterGroup) GetInt32Counter(name string) (*Int32Counter, bool) {
	counter, exists := cg.int32s[name]
	return counter, exists
}

// GetAtomicBool retrieves an atomic boolean by name
func (cg *CounterGroup) GetAtomicBool(name string) (*AtomicBool, bool) {
	atomicBool, exists := cg.bools[name]
	return atomicBool, exists
}

// GetAllStats returns statistics for all counters in the group
func (cg *CounterGroup) GetAllStats() CounterGroupStats {
	stats := CounterGroupStats{
		GroupName: cg.groupName,
		CreatedAt: cg.createdAt,
		Age:       time.Since(cg.createdAt),
		Counters:  make(map[string]CounterStats),
		Int32s:    make(map[string]Int32CounterStats),
		Bools:     make(map[string]AtomicBoolStats),
	}

	for name, counter := range cg.counters {
		stats.Counters[name] = counter.Stats()
	}

	for name, counter := range cg.int32s {
		stats.Int32s[name] = counter.Stats()
	}

	for name, atomicBool := range cg.bools {
		stats.Bools[name] = atomicBool.Stats()
	}

	return stats
}

// CounterGroupStats represents statistics for a counter group
type CounterGroupStats struct {
	GroupName string                         `json:"group_name"`
	CreatedAt time.Time                      `json:"created_at"`
	Age       time.Duration                  `json:"age"`
	Counters  map[string]CounterStats        `json:"counters"`
	Int32s    map[string]Int32CounterStats   `json:"int32s"`
	Bools     map[string]AtomicBoolStats     `json:"bools"`
}

// JSON returns the JSON representation of the group stats
func (cgs CounterGroupStats) JSON() ([]byte, error) {
	return json.MarshalIndent(cgs, "", "  ")
}

// ResetAll resets all counters in the group
func (cg *CounterGroup) ResetAll() map[string]interface{} {
	results := make(map[string]interface{})

	for name, counter := range cg.counters {
		results[name] = counter.Reset()
	}

	for name, counter := range cg.int32s {
		results[name] = counter.Reset()
	}

	for name, atomicBool := range cg.bools {
		results[name] = atomicBool.Load()
		atomicBool.Store(false)
	}

	return results
}

// RemoveCounter removes a counter from the group
func (cg *CounterGroup) RemoveCounter(name string) bool {
	_, exists := cg.counters[name]
	if exists {
		delete(cg.counters, name)
	}
	return exists
}

// RemoveInt32Counter removes an int32 counter from the group
func (cg *CounterGroup) RemoveInt32Counter(name string) bool {
	_, exists := cg.int32s[name]
	if exists {
		delete(cg.int32s, name)
	}
	return exists
}

// RemoveAtomicBool removes an atomic boolean from the group
func (cg *CounterGroup) RemoveAtomicBool(name string) bool {
	_, exists := cg.bools[name]
	if exists {
		delete(cg.bools, name)
	}
	return exists
}

// CounterNames returns all counter names
func (cg *CounterGroup) CounterNames() []string {
	names := make([]string, 0, len(cg.counters))
	for name := range cg.counters {
		names = append(names, name)
	}
	return names
}

// Int32CounterNames returns all int32 counter names
func (cg *CounterGroup) Int32CounterNames() []string {
	names := make([]string, 0, len(cg.int32s))
	for name := range cg.int32s {
		names = append(names, name)
	}
	return names
}

// AtomicBoolNames returns all atomic boolean names
func (cg *CounterGroup) AtomicBoolNames() []string {
	names := make([]string, 0, len(cg.bools))
	for name := range cg.bools {
		names = append(names, name)
	}
	return names
}