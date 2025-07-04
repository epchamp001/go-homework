package config

type WorkersConfig struct {
	Start int `mapstructure:"start"`
	Queue int `mapstructure:"queue"`
}
