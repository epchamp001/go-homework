package cache

// Delete удаляет ключ.
func (c *Cache[K, V]) Delete(k K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deleteLocked(k)
}

func (c *Cache[K, V]) deleteLocked(k K) {
	if e, ok := c.items[k]; ok {
		delete(c.items, k)
		if c.cfg.Strategy != nil {
			c.cfg.Strategy.RecordDeletion(k)
		}
		if c.cfg.OnEvict != nil {
			// передаём value как any, чтобы не зависеть от V в колбеке
			c.cfg.OnEvict(k, any(e.val))
		}
	}
}
