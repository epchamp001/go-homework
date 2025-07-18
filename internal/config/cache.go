package config

import "time"

type CacheConfig struct {
	Capacity int           `mapstructure:"capacity"`
	TTL      time.Duration `mapstructure:"ttl"`
}
