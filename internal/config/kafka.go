package config

import "time"

// KafkaConfig — настройки подключения к Kafka.
type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
}

// OutboxConfig — параметры работы transactional-outbox воркера.
type OutboxConfig struct {
	BatchSize int           `mapstructure:"batch_size"`
	Interval  time.Duration `mapstructure:"interval"`
}
