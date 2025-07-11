package integration

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/infrastructure/kafka/consumer"
	"pvz-cli/internal/infrastructure/kafka/producer"
	"pvz-cli/internal/repository/storage/postgres"
	"pvz-cli/internal/usecase/outboxworker"
	"time"
)

func (s *TestSuite) TestOutboxWorkerPublishesToKafka_WithWrappers() {
	t := s.T()

	ctx := context.Background()
	topic := s.createUniqueTopic(ctx)

	_, err := s.svc.AcceptOrder(
		ctx,
		"42", "123",
		time.Now().Add(24*time.Hour),
		1.5,
		models.PriceKopecks(10_000),
		models.PackageBox,
	)
	require.NoError(t, err)

	var outboxID string
	_ = s.masterPool.QueryRow(ctx,
		`SELECT id::text FROM outbox ORDER BY created_at DESC LIMIT 1`,
	).Scan(&outboxID)

	prodCfg := s.baseProd
	prodCfg.Topic = topic

	prod, err := producer.NewProducer(prodCfg)
	require.NoError(t, err)
	defer prod.Close()

	workerCtx, stopWorker := context.WithCancel(ctx)
	defer stopWorker()
	go outboxworker.NewWorker(
		s.tx,
		postgres.NewOutboxPostgresRepo(s.tx),
		prod,
		1, time.Second,
		s.log,
	).Run(workerCtx)

	recCh := make(chan *kgo.Record, 1)
	cons, err := consumer.NewConsumer(consumer.Config{
		Brokers:     []string{s.kafkaAddr},
		Topic:       topic,
		ResetOffset: "earliest",
		AutoCommit:  false,
	}, func(_ context.Context, r *kgo.Record) error {
		recCh <- r
		return nil
	})
	require.NoError(t, err)
	defer cons.Close()

	consCtx, cancelCons := context.WithCancel(ctx)
	go func() {
		if err := cons.Run(consCtx); err != nil &&
			!errors.Is(err, context.Canceled) &&
			!errors.Is(err, kgo.ErrClientClosed) {
			t.Fatalf("consumer error: %v", err)
		}
	}()

	var rec *kgo.Record
	select {
	case rec = <-recCh:
		cancelCons()
	case <-time.After(5 * time.Second):
		cancelCons()
		t.Fatal("timeout waiting for kafka message")
	}

	require.Equal(t, []byte(outboxID), rec.Key)

	var event struct {
		EventType string `json:"event_type"`
		Order     struct {
			ID     string `json:"id"`
			UserID string `json:"user_id"`
		} `json:"order"`
	}
	require.NoError(t, json.Unmarshal(rec.Value, &event))
	require.Equal(t, "order_accepted", event.EventType)
	require.Equal(t, "42", event.Order.ID)
	require.Equal(t, "123", event.Order.UserID)
}
