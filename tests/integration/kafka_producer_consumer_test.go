//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"errors"
	"pvz-cli/internal/infrastructure/kafka/codec"
	"pvz-cli/internal/infrastructure/kafka/consumer"
	"pvz-cli/internal/infrastructure/kafka/producer"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
)

func (s *TestSuite) TestKafkaProducerConsumer_WithWrappers() {
	ctx := context.Background()
	t := s.T()

	topic := s.createUniqueTopic(ctx)

	prodCfg := s.baseProd
	prodCfg.Topic = topic

	prod, err := producer.NewProducer(prodCfg)
	require.NoError(t, err)
	defer prod.Close()

	type Order struct{ ID, UserID int }
	order := Order{ID: 1, UserID: 123}
	payload, _ := codec.JSON.Marshal(order)

	err = prod.Send(ctx, []byte("order-1"), payload)
	require.NoError(t, err, "producer.Send failed")

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

	require.Equal(t, []byte("order-1"), rec.Key)
	require.Equal(t, payload, rec.Value)

	var got Order
	require.NoError(t, json.Unmarshal(rec.Value, &got))
	require.Equal(t, order, got)
}
