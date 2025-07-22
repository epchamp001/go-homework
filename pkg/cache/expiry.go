package cache

import "time"

// expiry вычисляет время истечения, если TTL>0.
func (c *Cache[K, V]) expiry() time.Time {
	if c.cfg.TTL == 0 {
		return time.Time{}
	}
	return time.Now().Add(c.cfg.TTL)
}
