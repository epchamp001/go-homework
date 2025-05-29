// Package models содержит основные сущности, используемые в приложении.
package models

import "time"

// HistoryEvent описывает событие смены статуса заказа.
type HistoryEvent struct {
	OrderID string      `json:"order_id"`
	Status  OrderStatus `json:"status"`
	Time    time.Time   `json:"timestamp"`
}
