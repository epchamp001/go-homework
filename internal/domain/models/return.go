package models

import "time"

// ReturnRecord представляет информацию о возврате заказа клиентом.
type ReturnRecord struct {
	OrderID    string    `json:"order_id"`
	UserID     string    `json:"user_id"`
	ReturnedAt time.Time `json:"returned_at"`
}
