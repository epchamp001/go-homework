package codes

import "pvz-cli/pkg/errs"

var (
	ErrOrderNotFound      = errs.New("ORDER_NOT_FOUND", "order not found")
	ErrOrderAlreadyExists = errs.New("ORDER_ALREADY_EXISTS", "order already exists")
	ErrStorageExpired     = errs.New("STORAGE_EXPIRED", "storage period expired")
	ErrValidationFailed   = errs.New("VALIDATION_FAILED", "validation failed")
)

const (
	CodeOrderAccepted = "ORDER_ACCEPTED"
	CodeOrderReturned = "ORDER_RETURNED"
	CodeProcessed     = "PROCESSED"
	CodeImported      = "IMPORTED"
	CodeOrder         = "ORDER"
	CodeTotal         = "TOTAL"
	CodeReturn        = "RETURN"
	CodePage          = "PAGE"
	CodeLimit         = "LIMIT"
	CodeHistory       = "HISTORY"
	CodeNext          = "NEXT"
)
