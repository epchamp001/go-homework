package service

import (
	"context"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/pkg/errs"
)

func (s *ServiceImpl) ListReturns(
	ctx context.Context,
	pg vo.Pagination,
) ([]*models.ReturnRecord, error) {

	roCtx := s.tx.WithReadOnly(ctx)

	records, err := s.hrRepo.ListReturns(roCtx, pg)
	if err != nil {
		return nil, errs.Wrap(err,
			errs.CodeDatabaseError, "list returns failed")
	}
	return records, nil
}
