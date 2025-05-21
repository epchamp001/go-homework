package config

type AppConfig struct {
	AppName         string `mapstructure:"app_name"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}
