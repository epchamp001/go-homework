package cache

// Set помещает значение в кэш, при необходимости выселяя старые.
func (c *Cache[K, V]) Set(k K, v V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[k]; exists {
		c.items[k] = entry[V]{val: v, exp: c.expiry()}
		if c.cfg.Strategy != nil {
			c.cfg.Strategy.RecordAccess(k)
		}
		return
	}

	if c.cfg.Capacity > 0 && len(c.items) >= c.cfg.Capacity {
		// спросить стратегию, кого выселить
		if c.cfg.Strategy != nil {
			if evictK, ok := c.cfg.Strategy.RecordInsertion(k); ok {
				c.deleteLocked(evictK)
			}
		}
	}

	c.items[k] = entry[V]{val: v, exp: c.expiry()}
}
