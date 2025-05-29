package config

// AppConfig содержит параметры приложения (имя и shutdown timeout).
type AppConfig struct {
	AppName         string `mapstructure:"app_name"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}
