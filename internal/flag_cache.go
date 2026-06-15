package internal

import "sync"

// FlagCache is a thread-safe store for RemoteFlag values.
// The entire cache is replaced atomically on each update.
type FlagCache struct {
	mu    sync.RWMutex
	flags map[string]RemoteFlag
}

// NewFlagCache returns an empty, ready-to-use FlagCache.
func NewFlagCache() *FlagCache {
	return &FlagCache{
		flags: make(map[string]RemoteFlag),
	}
}

// Update replaces the entire contents of the cache with the provided slice.
// The replacement is atomic from the perspective of concurrent readers.
func (c *FlagCache) Update(flags []RemoteFlag) {
	m := make(map[string]RemoteFlag, len(flags))
	for _, f := range flags {
		m[f.Key] = f
	}
	c.mu.Lock()
	c.flags = m
	c.mu.Unlock()
}

// Get returns the flag with the given key and a boolean indicating whether
// the flag was found.
func (c *FlagCache) Get(key string) (RemoteFlag, bool) {
	c.mu.RLock()
	f, ok := c.flags[key]
	c.mu.RUnlock()
	return f, ok
}

// All returns an immutable snapshot of all flags currently in the cache.
func (c *FlagCache) All() map[string]RemoteFlag {
	c.mu.RLock()
	snapshot := make(map[string]RemoteFlag, len(c.flags))
	for k, v := range c.flags {
		snapshot[k] = v
	}
	c.mu.RUnlock()
	return snapshot
}

// IsEmpty reports whether the cache holds no flags.
func (c *FlagCache) IsEmpty() bool {
	c.mu.RLock()
	empty := len(c.flags) == 0
	c.mu.RUnlock()
	return empty
}
