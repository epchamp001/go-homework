package cache

import "time"

// reaper удаляет просроченные элементы раз в TTL.
func (c *Cache[K, V]) reaper() {
	if c.cfg.TTL == 0 {
		return
	}
	ticker := time.NewTicker(c.cfg.TTL)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			var toDelete []K
			c.mu.RLock()
			for k, e := range c.items {
				if now.After(e.exp) {
					toDelete = append(toDelete, k)
				}
			}
			c.mu.RUnlock()
			for _, k := range toDelete {
				c.Delete(k)
			}
		case <-c.closing:
			return
		}
	}
}
