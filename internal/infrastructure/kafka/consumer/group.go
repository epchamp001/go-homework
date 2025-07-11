package consumer

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"pvz-cli/internal/infrastructure/kafka"
)

// GroupHandler задаёт колбэки пользовательской обработки.
type GroupHandler struct {
	OnMessage func(ctx context.Context, rec *kgo.Record) error
	OnAssign  func(ctx context.Context, parts map[string][]int32)
	OnRevoke  func(ctx context.Context, parts map[string][]int32)
	OnLost    func(ctx context.Context, parts map[string][]int32)
}

type Group struct {
	client *kgo.Client
	h      GroupHandler
}

// NewConsumerGroup создаёт консьюмер-группу.
func NewConsumerGroup(cfg Config, h GroupHandler,
	extra ...kafka.ClientOption) (*Group, error) {

	if cfg.GroupID == "" {
		return nil, fmt.Errorf("group_id is required for consumer group")
	}

	opts, err := cfg.FranzOpts()
	if err != nil {
		return nil, err
	}

	// колбэки ребаланса
	if h.OnAssign != nil {
		opts = append(opts, kgo.OnPartitionsAssigned(
			func(ctx context.Context, _ *kgo.Client, p map[string][]int32) {
				h.OnAssign(ctx, p)
			}))
	}
	if h.OnRevoke != nil {
		opts = append(opts, kgo.OnPartitionsRevoked(
			func(ctx context.Context, _ *kgo.Client, p map[string][]int32) {
				h.OnRevoke(ctx, p)
			}))
	}
	if h.OnLost != nil {
		opts = append(opts, kgo.OnPartitionsLost(
			func(ctx context.Context, _ *kgo.Client, p map[string][]int32) {
				h.OnLost(ctx, p)
			}))
	}

	for _, fn := range extra {
		opts = fn(opts)
	}

	cli, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return &Group{client: cli, h: h}, nil
}

// Run - основной цикл обработки.
func (g *Group) Run(ctx context.Context) error {
	for {
		fetches := g.client.PollFetches(ctx)

		if err := firstFetchErr(fetches.Errors()); err != nil {
			return err
		}

		fetches.EachRecord(func(rec *kgo.Record) {
			if g.h.OnMessage != nil && g.h.OnMessage(ctx, rec) == nil {
				g.client.MarkCommitRecords(rec)
			}
		})
	}
}

func (g *Group) Close() { g.client.Close() }
