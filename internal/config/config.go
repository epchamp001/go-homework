// Package config предоставляет функциональность для чтения и хранения
// конфигурационных параметров приложения.
package config

import (
	"github.com/spf13/viper"

	"path/filepath"
	"pvz-cli/pkg/errs"
)

// Config объединяет все конфигурации в одну структуру.
type Config struct {
	Logging    LoggingConfig    `mapstructure:"logging"`
	GRPCServer GRPCServerConfig `mapstructure:"grpc_server"`
	Gateway    GatewayConfig    `mapstructure:"gateway"`
}

// LoadConfig загружает и распаковывает конфигурацию по указанному пути.
//
// Если путь содержит расширение (.yaml/.yml), используется полный путь к файлу.
// Иначе ожидается config.{yaml,yml,json,...} внутри директории.
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
