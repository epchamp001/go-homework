package cache

// Close останавливает фоновые горутины (reaper).
func (c *Cache[K, V]) Close() {
	close(c.closing)
}
