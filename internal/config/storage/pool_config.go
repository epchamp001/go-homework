package storage

type PoolConfig struct {
	MaxConnections    int `mapstructure:"max_connections"`
	MinConnections    int `mapstructure:"min_connections"`
	MaxLifeTime       int `mapstructure:"max_lifetime"`
	MaxIdleTime       int `mapstructure:"max_idle_time"`
	HealthCheckPeriod int `mapstructure:"health_check_period"`
}
