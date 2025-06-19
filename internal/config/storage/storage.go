package storage

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"pvz-cli/pkg/logger"
)

type StorageConfig struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
}

// ConnectMaster возвращает пул к мастеру
func (s *StorageConfig) ConnectMaster(log logger.Logger) (*pgxpool.Pool, error) {
	return s.Postgres.connect(s.Postgres.Master, log)
}

// ConnectReplica(idx) – пул к replica[idx] (0-based)
func (s *StorageConfig) ConnectReplica(idx int, log logger.Logger) (*pgxpool.Pool, error) {
	if idx < 0 || idx >= len(s.Postgres.Replicas) {
		return nil, fmt.Errorf("replica index %d out of range", idx)
	}
	return s.Postgres.connect(s.Postgres.Replicas[idx], log)
}

// GetDSN(0) → master; GetDSN(1) → первая реплика; 2 → вторая, и т.д.
func (s *StorageConfig) GetDSN(id int) (string, error) {
	switch id {
	case 0:
		return s.Postgres.buildDSN(s.Postgres.Master), nil
	default:
		idx := id - 1
		if idx < 0 || idx >= len(s.Postgres.Replicas) {
			return "", fmt.Errorf("replica id %d not found", id)
		}
		return s.Postgres.buildDSN(s.Postgres.Replicas[idx]), nil
	}
}
