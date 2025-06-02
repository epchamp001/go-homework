package models

import "time"

// OrderStatus определяет текущее состояние заказа.
type OrderStatus string

const (
	// StatusAccepted означает, что заказ принят в ПВЗ.
	StatusAccepted OrderStatus = "ACCEPTED"

	// StatusIssued означает, что заказ выдан клиенту.
	StatusIssued OrderStatus = "ISSUED"

	// StatusReturned означает, что заказ возвращён курьеру или клиентом.
	StatusReturned OrderStatus = "RETURNED"

	// StatusExpired означает, что срок хранения заказа истёк.
	StatusExpired OrderStatus = "EXPIRED"
)

// Order представляет заказ в системе ПВЗ.
type Order struct {
	ID         string      `json:"order_id"`
	UserID     string      `json:"user_id"`
	Status     OrderStatus `json:"status"`
	ExpiresAt  time.Time   `json:"expires_at"`
	IssuedAt   *time.Time  `json:"issued_at,omitempty"`
	ReturnedAt *time.Time  `json:"returned_at,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	Package    string      `json:"package"`
	Weight     float64     `json:"weight"`
	Price      int64       `json:"price"`
	TotalPrice int64       `json:"total_price"`
}
