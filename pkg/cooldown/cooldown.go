package cooldown

import (
	"sync"
	"time"
)

type CooldownCache struct {
	mu    sync.Mutex
	items map[string]time.Time
}

func New() *CooldownCache {
	return &CooldownCache{
		items: make(map[string]time.Time),
	}
}
func (c *CooldownCache) ShouldAlert(key string, cooldown time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	last, ok := c.items[key]
	if ok && time.Since(last) < cooldown {
		return false
	}
	c.items[key] = time.Now()
	return true
}
