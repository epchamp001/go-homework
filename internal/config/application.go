package config

type ApplicationConfig struct {
	Name            string `mapstructure:"name"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}
