// inspired from https://github.com/patrickmn/go-cache
package cache

import (
	"sync"
	"time"
)

type MemCache struct {
	items sync.Map
	ci    time.Duration
}

// memcache will scan all objects for every clean interval and delete expired key.
func NewMemCache(ci time.Duration) *MemCache {
	c := &MemCache{
		items: sync.Map{},
		ci:    ci,
	}

	go c.runJanitor()
	return c
}

// return true if data is fresh
func (c *MemCache) Load(k string) (*Item, bool) {
	it, exists := c.Get(k)
	if !exists {
		return nil, false
	}
	return it, !it.Outdated()
}

// Get an item from the memcache. Returns the item or nil, and a bool indicating whether the key was found.
func (c *MemCache) Get(k string) (*Item, bool) {
	tmp, found := c.items.Load(k)
	if !found {
		return nil, false
	}
	item := tmp.(*Item)
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}
	return item, true
}

func (c *MemCache) Set(k string, it *Item) {
	c.items.Store(k, it)
}

// Delete an item from the memcache. Does nothing if the key is not in the memcache.
func (c *MemCache) Delete(k string) {
	c.items.Delete(k)
}

// start key scanning to delete expired keys
func (c *MemCache) runJanitor() {
	ticker := time.NewTicker(c.ci)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		}
	}
}

// Delete all expired items from the memcache.
func (c *MemCache) DeleteExpired() {
	c.items.Range(func(key, value interface{}) bool {
		v := value.(*Item)
		k := key.(string)
		// delete outdated for memory cache
		if v.Outdated() {
			c.items.Delete(k)
		}
		return true
	})
}
