package consumer

import (
	"fmt"
	"github.com/twmb/franz-go/pkg/kgo"
	"time"
)

// Config описывает минимальный набор параметров для консьюмера (как одиночного, так и в составе consumer-group).
// Если GroupID пустой  → создаётся stand-alone consumer.
// Если GroupID задан  → создаётся consumer-group.
type Config struct {
	Brokers     []string `yaml:"brokers"`      // список брокеров
	Topic       string   `yaml:"topic"`        // обязательный
	GroupID     string   `yaml:"group_id"`     // optional
	ResetOffset string   `yaml:"reset_offset"` // latest|earliest

	AutoCommit         bool          `yaml:"auto_commit"` // default=true
	AutoCommitInterval time.Duration `yaml:"auto_commit_interval"`
}

// FranzOpts превращает конфиг в []kgo.Opt.
func (c *Config) FranzOpts() ([]kgo.Opt, error) {
	if len(c.Brokers) == 0 {
		return nil, fmt.Errorf("brokers list empty")
	}
	if c.Topic == "" {
		return nil, fmt.Errorf("topic required")
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(c.Brokers...),
		kgo.ConsumeTopics(c.Topic),
		kgo.ClientID("pvz-api-consumer"),

		kgo.DialTimeout(10 * time.Second),
		kgo.SessionTimeout(45 * time.Second),
		kgo.FetchMaxWait(500 * time.Millisecond),
	}

	if c.GroupID != "" {
		opts = append(opts, kgo.ConsumerGroup(c.GroupID))
	}

	// offset reset
	var off kgo.Offset
	switch c.ResetOffset {
	case "", "latest":
		off = kgo.NewOffset().AtEnd()
	case "earliest":
		off = kgo.NewOffset().AtStart()
	default:
		return nil, fmt.Errorf("unknown reset_offset: %s", c.ResetOffset)
	}
	opts = append(opts,
		kgo.ConsumeStartOffset(off),
		kgo.ConsumeResetOffset(off),
	)

	if c.GroupID != "" {
		if c.AutoCommit {
			opts = append(opts, kgo.AutoCommitMarks())
			if c.AutoCommitInterval > 0 {
				opts = append(opts, kgo.AutoCommitInterval(c.AutoCommitInterval))
			}
		} else {
			opts = append(opts, kgo.DisableAutoCommit())
		}
	}
	return opts, nil
}
