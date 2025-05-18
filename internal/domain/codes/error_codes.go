package codes

import "pvz-cli/pkg/errs"

var (
	ErrInvalidPackage     = errs.New("INVALID_PACKAGE", "package data is invalid")
	ErrWeightTooHeavy     = errs.New("WEIGHT_TOO_HEAVY", "package exceeds weight limit")
	ErrOrderNotFound      = errs.New("ORDER_NOT_FOUND", "order not found")
	ErrOrderAlreadyExists = errs.New("ORDER_ALREADY_EXISTS", "order already exists")
	ErrStorageExpired     = errs.New("STORAGE_EXPIRED", "storage period expired")
	ErrValidationFailed   = errs.New("VALIDATION_FAILED", "validation failed")
	ErrInternal           = errs.New("INTERNAL_ERROR", "internal error")
)

const (
	CodeOrderAccepted = "ORDER_ACCEPTED"
	CodeOrderReturned = "ORDER_RETURNED"
	CodeProcessed     = "PROCESSED"
)
