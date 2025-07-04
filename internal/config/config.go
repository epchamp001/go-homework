// Package config предоставляет функциональность для чтения и хранения
// конфигурационных параметров приложения.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"

	"path/filepath"
	strgCfg "pvz-cli/internal/config/storage"
	"pvz-cli/pkg/errs"

	"github.com/joho/godotenv"
)

// Config объединяет все конфигурации в одну структуру.
type Config struct {
	Logging    LoggingConfig    `mapstructure:"logging"`
	GRPCServer GRPCServerConfig `mapstructure:"grpc_server"`
	Gateway    GatewayConfig    `mapstructure:"gateway"`
	Storage    strgCfg.StorageConfig
	Workers    WorkersConfig `mapstructure:"workers"`
	Admin      AdminConfig   `mapstructure:"admin"`
}

// LoadConfig загружает и распаковывает конфигурацию по указанному пути.
//
// Если путь содержит расширение (.yaml/.yml), используется полный путь к файлу.
// Иначе ожидается config.{yaml,yml,json,...} внутри директории.
func LoadConfig(configPath, envPath string) (*Config, error) {
	if err := godotenv.Load(envPath); err != nil {
		fmt.Printf("WARNING: error loading .env from %s: %v\n", envPath, err)
	}

	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Хелпер для биндинга
	bind := func(key, env string) {
		if err := v.BindEnv(key, env); err != nil {
			fmt.Printf("WARNING: BindEnv %s -> %s failed: %v\n", key, env, err)
		}
	}

	bind("storage.postgres.master.host", "PG_HOST")
	bind("storage.postgres.master.port", "PG_MASTER_PORT")
	bind("storage.postgres.database", "PG_DATABASE")
	bind("storage.postgres.username", "PG_SUPER_USER")
	bind("storage.postgres.password", "PG_SUPER_PASSWORD")

	bind("storage.postgres.replicas[0].username", "PG_REPL_USER")
	bind("storage.postgres.replicas[0].password", "PG_REPL_PASSWORD")
	bind("storage.postgres.replicas[1].username", "PG_REPL_USER")
	bind("storage.postgres.replicas[1].password", "PG_REPL_PASSWORD")

	bind("admin.user",  "ADMIN_USER")
	bind("admin.pass",  "ADMIN_PASS")

	if ext := filepath.Ext(configPath); ext == ".yaml" || ext == ".yml" {
		v.SetConfigFile(configPath)
	} else {
		v.AddConfigPath(configPath)
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}
	if err := v.ReadInConfig(); err != nil {
		return nil, errs.Wrap(
			err,
			errs.CodeConfigError,
			"cannot read config file",
			"path", configPath,
		)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, errs.Wrap(
			err,
			errs.CodeInvalidConfiguration,
			"cannot unmarshal config into struct",
		)
	}

	if h := os.Getenv("PG_HOST"); h != "" {
		for i := range cfg.Storage.Postgres.Replicas {
			cfg.Storage.Postgres.Replicas[i].Host = h
		}
	}

	return &cfg, nil
}
