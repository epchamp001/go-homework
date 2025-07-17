package producer

import (
	"context"
	"pvz-cli/internal/infrastructure/kafka"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct{ client *kgo.Client }

// NewProducer создаёт продюсер.
func NewProducer(cfg Config, extra ...kafka.ClientOption) (*Producer, error) {
	opts, err := cfg.FranzOpts()
	if err != nil {
		return nil, err
	}
	for _, fn := range extra {
		opts = fn(opts)
	}
	cli, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return &Producer{client: cli}, nil
}

// Send работает, только если задан DefaultProduceTopic.
func (p *Producer) Send(ctx context.Context, key, val []byte, ro ...kafka.RecordOption) error {
	return p.SendTo(ctx, "", key, val, ro...)
}

// SendTo явно указываем topic (пустая строка ⇒ берётся дефолтный).
func (p *Producer) SendTo(ctx context.Context, topic string,
	key, val []byte, ro ...kafka.RecordOption) error {

	rec := &kgo.Record{Topic: topic, Key: key, Value: val}
	for _, fn := range ro {
		fn(rec)
	}

	return p.client.ProduceSync(ctx, rec).FirstErr()
}

func (p *Producer) Flush(ctx context.Context) error {
	return p.client.Flush(ctx)
}

func (p *Producer) Close() {
	p.client.Close()
}
