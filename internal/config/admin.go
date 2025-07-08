package config

type AdminConfig struct {
	User string `mapstructure:"user"`
	Pass string `mapstructure:"pass"`
}
