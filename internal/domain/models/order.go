package models

import "time"

type OrderStatus string

const (
	StatusAccepted OrderStatus = "ACCEPTED" // принят в ПВЗ
	StatusIssued   OrderStatus = "ISSUED"   // выдан клиенту
	StatusReturned OrderStatus = "RETURNED" // возвращён курьеру/клиентом
	StatusExpired  OrderStatus = "EXPIRED"  // срок хранения истёк
)

type Order struct {
	ID         string      `json:"order_id"`
	UserID     string      `json:"user_id"`
	Status     OrderStatus `json:"status"`
	ExpiresAt  time.Time   `json:"expires_at"`
	IssuedAt   *time.Time  `json:"issued_at,omitempty"`
	ReturnedAt *time.Time  `json:"returned_at,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
}
