package cache

// Flush удаляет все элементы из кэша.
func (c *Cache[K, V]) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// колбек на выселение
	if c.cfg.OnEvict != nil {
		for k, e := range c.items {
			c.cfg.OnEvict(k, any(e.val))
		}
	}

	// очищаем элементы
	for k := range c.items {
		delete(c.items, k)
	}

	// сбрасываем стратегию
	if c.cfg.Strategy != nil {
		c.cfg.Strategy.Reset()
	}
}
