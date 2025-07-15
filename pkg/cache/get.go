package cache

import "time"

// Get возвращает значение и true, если ключ найден и не просрочен.
func (c *Cache[K, V]) Get(k K) (V, bool) {
	c.mu.RLock()
	e, ok := c.items[k]
	c.mu.RUnlock()

	if !ok {
		var zero V
		return zero, false
	}
	if c.cfg.TTL > 0 && time.Now().After(e.exp) {
		c.Delete(k)
		var zero V
		return zero, false
	}
	if c.cfg.Strategy != nil {
		c.cfg.Strategy.RecordAccess(k)
	}
	return e.val, true
}
