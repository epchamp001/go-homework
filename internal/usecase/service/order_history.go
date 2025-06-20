package service

import (
	"context"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/pkg/errs"
)

func (s *ServiceImpl) OrderHistory(
	ctx context.Context,
	pg vo.Pagination,
) ([]*models.HistoryEvent, int, error) {

	roCtx := s.tx.WithReadOnly(ctx)

	events, err := s.hrRepo.History(roCtx, pg)
	if err != nil {
		return nil, 0, errs.Wrap(err,
			errs.CodeDatabaseError, "list history failed")
	}
	return events, len(events), nil
}
