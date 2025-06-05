package mappers

import (
	"pvz-cli/internal/domain/vo"
	pvzpb "pvz-cli/pkg/pvz"
)

func ProtoToDomainPagination(p *pvzpb.Pagination) vo.Pagination {
	if p == nil {
		return vo.Pagination{}
	}
	return vo.Pagination{
		Page:  int(p.Page),
		Limit: int(p.CountOnPage),
	}
}
