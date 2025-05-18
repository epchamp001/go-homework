package config

import (
	"github.com/spf13/viper"
	"pvz-cli/pkg/errs"
)

type Config struct {
	Logging LoggingConfig     `mapstructure:"logging"`
	App     ApplicationConfig `mapstructure:"app"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, errs.Wrap(
			err,
			errs.CodeConfigError,
			"cannot read config filerepo",
			"path", configPath,
		)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, errs.Wrap(
			err,
			errs.CodeInvalidConfiguration,
			"cannot unmarshal config into struct",
		)
	}

	return &config, nil
}
