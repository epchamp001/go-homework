package producer

import (
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Config struct {
	Brokers       []string `yaml:"brokers"`
	Topic         string   `yaml:"topic"`
	Idempotent    bool     `yaml:"idempotent"`
	RequiredAcks  string   `yaml:"required_acks"` // "", leader, all, none
	Partitioner   string   `yaml:"partitioner"`   // "", sticky, roundrobin, manual
	Compression   string   `yaml:"compression"`   // "", none, gzip, …
	BatchMaxBytes int32    `yaml:"batch_max_bytes"`
	LingerMs      int      `yaml:"linger_ms"`
	MaxAttempts   int      `yaml:"max_attempts"`
}

// FranzOpts превращает конфиг в набор kgo.Opt.
func (c *Config) FranzOpts() ([]kgo.Opt, error) {
	if len(c.Brokers) == 0 {
		return nil, fmt.Errorf("brokers list empty")
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(c.Brokers...),
		kgo.ClientID("pvz-api"),
		// базовые таймауты чтобы не зависал при сетевых проблемах
		kgo.DialTimeout(10 * time.Second),
		kgo.ProduceRequestTimeout(30 * time.Second),
		kgo.SessionTimeout(45 * time.Second),
	}

	// topic по умолчанию
	if c.Topic != "" {
		opts = append(opts, kgo.DefaultProduceTopic(c.Topic))
	}

	// ACKs & идемпотентность
	switch c.RequiredAcks {
	case "", "leader":
		opts = append(opts, kgo.RequiredAcks(kgo.LeaderAck()))
	case "all":
		opts = append(opts, kgo.RequiredAcks(kgo.AllISRAcks()))
	case "none":
		opts = append(opts, kgo.RequiredAcks(kgo.NoAck()))
		// при acks=none ВСЕ идемпотентность и ретраи запрещаем
		if c.Idempotent {
			return nil, fmt.Errorf("idempotent=true несовместимо с acks=none")
		}
		if c.MaxAttempts > 0 {
			return nil, fmt.Errorf("retries>0 несовместимы с acks=none")
		}
	default:
		return nil, fmt.Errorf("unknown required_acks: %s", c.RequiredAcks)
	}

	if !c.Idempotent {
		opts = append(opts, kgo.DisableIdempotentWrite())
	}

	// партиционер
	switch c.Partitioner {
	case "", "sticky":
		opts = append(opts, kgo.RecordPartitioner(kgo.StickyKeyPartitioner(nil)))
	case "roundrobin":
		opts = append(opts, kgo.RecordPartitioner(kgo.RoundRobinPartitioner()))
	case "manual":
		opts = append(opts, kgo.RecordPartitioner(kgo.ManualPartitioner()))
	default:
		return nil, fmt.Errorf("unknown partitioner: %s", c.Partitioner)
	}

	// сжатие
	switch c.Compression {
	case "", "none":
	case "gzip":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.GzipCompression()))
	case "snappy":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.SnappyCompression()))
	case "lz4":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.Lz4Compression()))
	case "zstd":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.ZstdCompression()))
	default:
		return nil, fmt.Errorf("unknown compression: %s", c.Compression)
	}

	// батчинг
	if c.BatchMaxBytes > 0 {
		opts = append(opts, kgo.ProducerBatchMaxBytes(c.BatchMaxBytes))
	}
	if c.LingerMs > 0 {
		opts = append(opts, kgo.ProducerLinger(time.Duration(c.LingerMs)*time.Millisecond))
	}

	// ретраи
	if c.MaxAttempts > 0 {
		opts = append(opts, kgo.RecordRetries(c.MaxAttempts))
	}

	return opts, nil
}
