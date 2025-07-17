package outboxworker

import (
	"context"
	"errors"
	"pvz-cli/internal/domain/models"
	repoMock "pvz-cli/internal/usecase/mock"
	"pvz-cli/pkg/logger"
	txMock "pvz-cli/pkg/txmanager/mock"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestWorker_ProcessOnce(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	type fields struct {
		tx   *txMock.TxManagerMock
		repo *repoMock.OutboxRepositoryMock
	}

	tests := []struct {
		name    string
		prepare func(f *fields)
	}{
		{
			name: "NoReadyRecords",
			prepare: func(f *fields) {
				// Транзакция должна выполниться и просто вернуть nil
				f.tx.WithTxMock.Set(func(
					_ context.Context, _ pgx.TxIsoLevel, _ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(ctx)
				})
				// PickReadyTx отдаёт пустой список — больше не ждём MarkProcessing
				f.repo.PickReadyTxMock.Set(func(_ context.Context, _ int) ([]models.OutboxRecord, error) {
					return nil, nil
				})
			},
		},
		{
			name: "PickError",
			prepare: func(f *fields) {
				f.tx.WithTxMock.Set(func(
					_ context.Context, _ pgx.TxIsoLevel, _ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(ctx)
				})
				// PickReadyTx возвращает ошибку — MarkProcessing не должен вызываться
				f.repo.PickReadyTxMock.Set(func(_ context.Context, _ int) ([]models.OutboxRecord, error) {
					return nil, errors.New("db failure")
				})
			},
		},
		{
			name: "InvalidPayload_FinalFail",
			prepare: func(f *fields) {
				badRec := models.OutboxRecord{
					ID:       uuid.New(),
					Payload:  []byte("not a json"),
					Attempts: 1,
				}
				f.tx.WithTxMock.Set(func(
					_ context.Context, _ pgx.TxIsoLevel, _ pgx.TxAccessMode,
					fn func(context.Context) error,
				) error {
					return fn(ctx)
				})
				f.repo.PickReadyTxMock.Set(func(_ context.Context, _ int) ([]models.OutboxRecord, error) {
					return []models.OutboxRecord{badRec}, nil
				})
				// Теперь точно вызывается MarkProcessing
				f.repo.MarkProcessingMock.Set(func(_ context.Context, ids []uuid.UUID) error {
					assert.Equal(t, badRec.ID, ids[0])
					return nil
				})
				// И MarkFinalFailed при invalid payload
				f.repo.MarkFinalFailedMock.Set(func(_ context.Context, id uuid.UUID) error {
					assert.Equal(t, badRec.ID, id)
					return nil
				})
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mc := minimock.NewController(t)

			f := &fields{
				tx:   txMock.NewTxManagerMock(mc),
				repo: repoMock.NewOutboxRepositoryMock(mc),
			}
			if tt.prepare != nil {
				tt.prepare(f)
			}

			// тихий logger, выводим только ошибки
			log, err := logger.NewLogger(
				logger.WithMode("dev"),
				logger.WithLevel(zapcore.ErrorLevel),
				logger.WithDisableCaller(true),
				logger.WithDisableStacktrace(true),
				logger.WithOutputPaths("stdout"),
				logger.WithErrorOutputPaths("stderr"),
			)
			require.NoError(t, err)

			w := NewWorker(f.tx, f.repo, nil, 10, time.Millisecond, log)
			w.processOnce(ctx)

			mc.Wait(time.Second)
		})
	}
}
