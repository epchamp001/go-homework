package mappers

import (
	"pvz-cli/internal/domain/models"
	"pvz-cli/pkg/errs"
	pvzpb "pvz-cli/pkg/pvz"
	"strconv"
	"time"
)

func ProtoToDomainOrderForImport(req *pvzpb.AcceptOrderRequest) (*models.Order, error) {
	if req == nil {
		return nil, errs.New(errs.CodeMissingParameter, "AcceptOrderRequest is nil")
	}

	if req.OrderId == 0 {
		return nil, errs.New(errs.CodeMissingParameter, "order_id обязательный параметр")
	}
	if req.UserId == 0 {
		return nil, errs.New(errs.CodeMissingParameter, "user_id обязательный параметр")
	}
	if req.ExpiresAt == nil {
		return nil, errs.New(errs.CodeMissingParameter, "expires_at обязательный параметр")
	}
	if req.Weight <= 0 {
		return nil, errs.New(errs.CodeInvalidParameter, "weight должно быть больше 0")
	}
	if req.Price <= 0 {
		return nil, errs.New(errs.CodeInvalidParameter, "price должно быть больше 0")
	}

	idStr := strconv.FormatUint(req.GetOrderId(), 10)
	userStr := strconv.FormatUint(req.GetUserId(), 10)

	status := models.StatusAccepted

	createdAt := time.Now()

	weight := float64(req.GetWeight())
	priceKopecks := models.PriceKopecks(int64(req.GetPrice() * 100))

	var pkg models.PackageType
	if req.Package != nil {
		pkg = ProtoToDomainPackage(*req.Package)
	}

	expiresAt := req.ExpiresAt.AsTime()

	return &models.Order{
		ID:        idStr,
		UserID:    userStr,
		Status:    status,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
		Weight:    weight,
		Price:     priceKopecks,
		Package:   pkg,
	}, nil
}
