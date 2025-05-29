// Package codes содержит предопределённые ошибки и кодовые константы,
// используемые в бизнес-логике приложения.
package codes

import "pvz-cli/pkg/errs"

var (
	// ErrOrderNotFound возникает, если заказ с указанным ID не найден.
	ErrOrderNotFound = errs.New("ORDER_NOT_FOUND", "order not found")

	// ErrOrderAlreadyExists возникает при попытке создать заказ с уже существующим ID.
	ErrOrderAlreadyExists = errs.New("ORDER_ALREADY_EXISTS", "order already exists")

	// ErrStorageExpired сигнализирует, что срок хранения заказа истёк.
	ErrStorageExpired = errs.New("STORAGE_EXPIRED", "storage period expired")

	// ErrValidationFailed указывает на некорректные входные данные или параметры.
	ErrValidationFailed = errs.New("VALIDATION_FAILED", "validation failed")

	// ErrWeightTooHeavy возникает, если вес заказа превышает допустимый предел для выбранной упаковки.
	ErrWeightTooHeavy = errs.New("WEIGHT_TO_HEAVY", "weight too heavy")

	// ErrInvalidPackage возникает при передаче недопустимого типа упаковки.
	ErrInvalidPackage = errs.New("INVALID_PACKAGE", "invalid package")
)

const (
	// CodeOrderAccepted — заказ успешно принят.
	CodeOrderAccepted = "ORDER_ACCEPTED"

	// CodeOrderReturned — заказ успешно возвращён.
	CodeOrderReturned = "ORDER_RETURNED"

	// CodeProcessed — команда успешно обработана.
	CodeProcessed = "PROCESSED"

	// CodeImported — данные успешно импортированы.
	CodeImported = "IMPORTED"

	// CodeOrder — префикс строки с данными о заказе.
	CodeOrder = "ORDER"

	// CodeTotal — префикс строки с общим числом заказов.
	CodeTotal = "TOTAL"

	// CodeReturn — префикс строки возврата заказа.
	CodeReturn = "RETURN"

	// CodePage — информация о текущей странице (пагинация).
	CodePage = "PAGE"

	// CodeLimit — лимит записей на страницу (пагинация).
	CodeLimit = "LIMIT"

	// CodeHistory — префикс для истории статусов заказа.
	CodeHistory = "HISTORY"

	// CodeNext — указывает на ID последнего заказа в текущей порции (scroll).
	CodeNext = "NEXT"
)
