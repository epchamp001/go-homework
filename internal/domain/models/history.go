package models

import "time"

type HistoryEvent struct {
	OrderID string      `json:"order_id"`
	Status  OrderStatus `json:"status"`
	Time    time.Time   `json:"timestamp"`
}
