package postgres

import (
	"context"
	"encoding/json"
	"pvz-cli/internal/domain/models"
	"pvz-cli/pkg/errs"
	"pvz-cli/pkg/txmanager"
	"time"

	"github.com/google/uuid"
)

type OutboxPostgresRepo struct {
	conn txmanager.TxManager
}

func NewOutboxPostgresRepo(conn txmanager.TxManager) *OutboxPostgresRepo {
	return &OutboxPostgresRepo{conn: conn}
}

const (
	insertOutboxSQL = `
INSERT INTO outbox (id, payload, status, created_at)
VALUES ($1,$2,'CREATED',now());`

	// Берём только готовые записи: attempts<3
	pickReadySQL = `
SELECT id, payload, attempts, last_attempt_at, created_at
FROM   outbox
WHERE  status = 'CREATED'
  AND  attempts < 3
  AND  (last_attempt_at IS NULL
        OR last_attempt_at < now() - interval '2 seconds')
ORDER  BY created_at, id
FOR UPDATE SKIP LOCKED
LIMIT  $1;`

	markProcessingSQL = `
UPDATE outbox
SET    status = 'PROCESSING'
WHERE  id = ANY($1);`

	markCompletedSQL = `
UPDATE outbox
SET    status  = 'COMPLETED',
       sent_at = $2
WHERE  id = $1;`

	// ещё будем пробовать позже
	markRetrySQL = `
UPDATE outbox
SET    attempts        = attempts + 1,
       last_attempt_at = now(),
       error           = $2
WHERE  id = $1;`

	// попыток больше нет
	markFinalFailedSQL = `
UPDATE outbox
SET    status          = 'FAILED',
       attempts        = attempts + 1,
       last_attempt_at = now(),
       error           = 'NO_ATTEMPTS_LEFT'
WHERE  id = $1;`
)

func (r *OutboxPostgresRepo) Add(ctx context.Context, evt *models.OrderEvent) error {
	exec := r.conn.GetExecutor(ctx)

	b, err := json.Marshal(evt)
	if err != nil {
		return errs.Wrap(err, errs.CodeDatabaseError,
			"marshal event payload", "event_id", evt.EventID)
	}
	if _, err = exec.Exec(ctx, insertOutboxSQL, evt.EventID, b); err != nil {
		return errs.Wrap(err, errs.CodeDatabaseError,
			"insert outbox event", "event_id", evt.EventID)
	}
	return nil
}

func (r *OutboxPostgresRepo) PickReadyTx(ctx context.Context, limit int) ([]models.OutboxRecord, error) {
	exec := r.conn.GetExecutor(ctx)

	rows, err := exec.Query(ctx, pickReadySQL, limit)
	if err != nil {
		return nil, errs.Wrap(err, errs.CodeDatabaseError, "select ready outbox records")
	}
	defer rows.Close()

	var res []models.OutboxRecord
	for rows.Next() {
		var rec models.OutboxRecord
		if err := rows.Scan(
			&rec.ID,
			&rec.Payload,
			&rec.Attempts,
			&rec.LastAttemptAt,
			&rec.CreatedAt,
		); err != nil {
			return nil, errs.Wrap(err, errs.CodeDatabaseError, "scan outbox record")
		}
		res = append(res, rec)
	}
	return res, nil
}

func (r *OutboxPostgresRepo) MarkProcessing(ctx context.Context, ids []uuid.UUID) error {
	exec := r.conn.GetExecutor(ctx)
	if _, err := exec.Exec(ctx, markProcessingSQL, ids); err != nil {
		return errs.Wrap(err, errs.CodeDatabaseError, "mark processing")
	}
	return nil
}

func (r *OutboxPostgresRepo) MarkCompleted(ctx context.Context, id uuid.UUID, sentAt time.Time) error {
	exec := r.conn.GetExecutor(ctx)
	if _, err := exec.Exec(ctx, markCompletedSQL, id, sentAt); err != nil {
		return errs.Wrap(err, errs.CodeDatabaseError, "mark completed", "id", id)
	}
	return nil
}

func (r *OutboxPostgresRepo) MarkRetry(ctx context.Context, id uuid.UUID, errMsg string) error {
	exec := r.conn.GetExecutor(ctx)
	if _, err := exec.Exec(ctx, markRetrySQL, id, errMsg); err != nil {
		return errs.Wrap(err, errs.CodeDatabaseError, "mark retry", "id", id)
	}
	return nil
}

func (r *OutboxPostgresRepo) MarkFinalFailed(ctx context.Context, id uuid.UUID) error {
	exec := r.conn.GetExecutor(ctx)
	if _, err := exec.Exec(ctx, markFinalFailedSQL, id); err != nil {
		return errs.Wrap(err, errs.CodeDatabaseError, "mark final failed", "id", id)
	}
	return nil
}
