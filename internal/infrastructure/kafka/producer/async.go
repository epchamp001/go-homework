package producer

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
	"pvz-cli/internal/infrastructure/kafka"
)

type AsyncProducer struct {
	client *kgo.Client
	ErrCh  chan error
}

func NewAsyncProducer(cfg Config, extra ...kafka.ClientOption) (*AsyncProducer, error) {
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
	return &AsyncProducer{client: cli, ErrCh: make(chan error, 1024)}, nil
}

func (p *AsyncProducer) Publish(key, val []byte, ro ...kafka.RecordOption) {
	rec := &kgo.Record{Key: key, Value: val}
	for _, fn := range ro {
		fn(rec)
	}
	p.client.Produce(context.Background(), rec, func(_ *kgo.Record, err error) {
		if err != nil && p.ErrCh != nil {
			p.ErrCh <- err
		}
	})
}

func (p *AsyncProducer) Flush(ctx context.Context) error { return p.client.Flush(ctx) }

// Close закрывает клиент и канал.
func (p *AsyncProducer) Close(ctx context.Context) error {
	if err := p.client.Flush(ctx); err != nil {
		return err
	}
	p.client.Close()
	if p.ErrCh != nil {
		close(p.ErrCh)
	}
	return nil
}
