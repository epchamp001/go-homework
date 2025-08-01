package config

type GRPCServerConfig struct {
	Enable          bool   `mapstructure:"enabled"`
	Endpoint        string `mapstructure:"endpoint"`
	Port            int    `mapstructure:"port"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

type GatewayConfig struct {
	Port     int    `mapstructure:"port"`
	Endpoint string `mapstructure:"endpoint"`
}

type MetricsConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Port     int    `mapstructure:"port"`
}
