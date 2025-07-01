package storage

import (
	"context"
	"fmt"
	"pvz-cli/pkg/logger"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	Master             PostgresEndpoint   `mapstructure:"master"`
	Replicas           []PostgresEndpoint `mapstructure:"replicas"`
	Database           string             `mapstructure:"database"`
	Username           string             `mapstructure:"username"`
	Password           string             `mapstructure:"password"`
	SSLMode            string             `mapstructure:"ssl_mode"`
	ConnectionAttempts int                `mapstructure:"connection_attempts"`
	Pool               PoolConfig         `mapstructure:"pool"`
}

func (pc *PostgresConfig) connect(ep PostgresEndpoint, log logger.Logger) (*pgxpool.Pool, error) {

	dsn := pc.buildDSN(ep)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DSN: %w", err)
	}

	poolCfg.MaxConns = int32(pc.Pool.MaxConnections)
	poolCfg.MinConns = int32(pc.Pool.MinConnections)
	poolCfg.MaxConnLifetime = time.Duration(pc.Pool.MaxLifeTime) * time.Second
	poolCfg.MaxConnIdleTime = time.Duration(pc.Pool.MaxIdleTime) * time.Second
	poolCfg.HealthCheckPeriod = time.Duration(pc.Pool.HealthCheckPeriod) * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	for i := 1; i <= pc.ConnectionAttempts; i++ {
		pingErr := pool.Ping(context.Background())
		if pingErr == nil {
			log.Infow("Successfully connected to PostgreSQL",
				"host", ep.Host,
				"port", ep.Port,
				"attempt", i,
			)
			return pool, nil
		}

		log.Warnw("PostgreSQL ping failed",
			"host", ep.Host,
			"port", ep.Port,
			"attempt", i,
			"error", pingErr,
		)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("could not connect after %d attempts", pc.ConnectionAttempts)
}

func (pc *PostgresConfig) buildDSN(ep PostgresEndpoint) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		ep.Host, ep.Port, pc.Username, pc.Password, pc.Database, pc.SSLMode,
	)
}
