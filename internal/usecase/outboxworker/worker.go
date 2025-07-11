package outboxworker

import (
	"context"
	"github.com/google/uuid"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/infrastructure/kafka/codec"
	"pvz-cli/internal/infrastructure/kafka/producer"
	"pvz-cli/internal/usecase"
	"pvz-cli/pkg/logger"
	"pvz-cli/pkg/txmanager"
	"time"
)

type Worker struct {
	tx        txmanager.TxManager
	repo      usecase.OutboxRepository
	producer  *producer.Producer
	batchSize int
	interval  time.Duration
	log       logger.Logger
}

// NewWorker создаёт новый outbox-воркер.
// batchSize максимальное число сообщений за итерацию,
// interval задержка между итерациями.
func NewWorker(
	tx txmanager.TxManager,
	repo usecase.OutboxRepository,
	producer *producer.Producer,
	batchSize int,
	interval time.Duration,
	log logger.Logger,
) *Worker {
	return &Worker{tx: tx, repo: repo, producer: producer, batchSize: batchSize, interval: interval, log: log}
}

// Run запускает бесконечный цикл обработки outbox.
func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			w.processOnce(ctx)
		}
	}
}

// processOnce делает одну итерацию: резервирует сообщения и отправляет их.
func (w *Worker) processOnce(ctx context.Context) {
	var recs []models.OutboxRecord

	// выбрать готовые (attempts<3 и прошёл backoff) и пометить их PROCESSING
	if err := w.tx.WithTx(ctx, txmanager.IsolationLevelReadCommitted, txmanager.AccessModeReadWrite,
		func(txCtx context.Context) error {
			var err error
			recs, err = w.repo.PickReadyTx(txCtx, w.batchSize)
			if err != nil {
				w.log.Errorw("outbox: failed to pick ready records", "error", err)
				return err
			}
			if len(recs) == 0 {
				return nil
			}
			ids := make([]uuid.UUID, len(recs))
			for i, r := range recs {
				ids[i] = r.ID
			}
			if err := w.repo.MarkProcessing(txCtx, ids); err != nil {
				w.log.Errorw("outbox: failed to mark processing", "error", err)
				return err
			}
			return nil
		}); err != nil {
		return
	}

	// для каждого записанного рекорда отправить в Kafka и финализировать
	for _, r := range recs {
		var evt models.OrderEvent
		if err := codec.JSON.Unmarshal(r.Payload, &evt); err != nil {
			w.log.Errorw("outbox: invalid payload", "id", r.ID, "error", err)
			_ = w.repo.MarkFinalFailed(ctx, r.ID) // payload неисправен - сразу проваливаем
			continue
		}

		if err := w.producer.Send(context.Background(), []byte(evt.EventID.String()), r.Payload); err != nil {
			w.log.Errorw("outbox: send to kafka failed", "id", r.ID, "error", err)
			// если попыток больше нет — финально проваливаем, иначе отмечаем retry
			if r.Attempts+1 >= 3 {
				_ = w.repo.MarkFinalFailed(ctx, r.ID)
			} else {
				_ = w.repo.MarkRetry(ctx, r.ID, err.Error())
			}
			continue
		}

		// успешно отправлено
		sentAt := time.Now().UTC()
		if err := w.repo.MarkCompleted(ctx, r.ID, sentAt); err != nil {
			w.log.Errorw("outbox: failed to mark completed", "id", r.ID, "error", err)
		}
	}
}
