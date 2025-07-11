package consumer

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"

	"pvz-cli/internal/infrastructure/kafka"
)

type Handler func(ctx context.Context, rec *kgo.Record) error

type Consumer struct {
	client  *kgo.Client
	handler Handler
}

// NewConsumer создаёт одиночного консьюмера (GroupID в cfg должен быть пустым).
func NewConsumer(cfg Config, h Handler, extra ...kafka.ClientOption) (*Consumer, error) {
	if cfg.GroupID != "" {
		return nil, fmt.Errorf("group_id must be empty for stand-alone consumer")
	}
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
	return &Consumer{client: cli, handler: h}, nil
}

// Run запускает бесконечный Poll-цикл.
func (c *Consumer) Run(ctx context.Context) error {
	for {
		fetches := c.client.PollFetches(ctx)

		if err := firstFetchErr(fetches.Errors()); err != nil {
			return err
		}

		fetches.EachRecord(func(rec *kgo.Record) {
			if err := c.handler(ctx, rec); err != nil {
				fmt.Printf("handler error: %v\n", err)
				return
			}
			// отмечаем запись как обработанную → offset будет закоммичен
			c.client.MarkCommitRecords(rec)
		})
	}
}

// Close завершает работу клиента.
func (c *Consumer) Close() { c.client.Close() }
