package config

import (
	"github.com/spf13/viper"
	"path/filepath"
	"pvz-cli/pkg/errs"
)

type Config struct {
	Logging LoggingConfig `mapstructure:"logging"`
}

func LoadConfig(configPath string) (*Config, error) {

	viper := viper.New()

	// Если путь содержит расширение – явно указываем полный файл:
	if ext := filepath.Ext(configPath); ext == ".yaml" || ext == ".yml" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath(configPath)
		viper.SetConfigName("config")
	}

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
