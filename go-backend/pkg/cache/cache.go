package cache

import (
	"sync"
	"time"
)

type item struct {
	value      any
	expiration int64
}

// MemoryCache represents a thread-safe in-memory cache with TTL support
type MemoryCache struct {
	items map[string]item
	mu    sync.RWMutex
}

var (
	defaultCache *MemoryCache
	once         sync.Once
)

func GetCache() *MemoryCache {
	once.Do(func() {
		defaultCache = NewMemoryCache()
	})
	return defaultCache
}

func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{
		items: make(map[string]item),
	}
	go c.startJanitor(5 * time.Minute)
	return c
}

func (c *MemoryCache) Set(key string, value any, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp int64
	if duration > 0 {
		exp = time.Now().Add(duration).UnixNano()
	}

	c.items[key] = item{
		value:      value,
		expiration: exp,
	}
}

func (c *MemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	it, found := c.items[key]
	if !found {
		return nil, false
	}

	if it.expiration > 0 && time.Now().UnixNano() > it.expiration {
		return nil, false
	}

	return it.value, true
}

func (c *MemoryCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *MemoryCache) startJanitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now().UnixNano()
		for k, it := range c.items {
			if it.expiration > 0 && now > it.expiration {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}
