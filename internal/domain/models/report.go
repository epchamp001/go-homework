package models

type ClientReport struct {
	UserID           string
	TotalOrders      int
	ReturnedOrders   int
	TotalPurchaseSum PriceKopecks
}
