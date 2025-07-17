package models

type OrderEventType string

const (
	OrderAccepted          OrderEventType = "order_accepted"
	OrderReturnedToCourier OrderEventType = "order_returned_to_courier"
	OrderIssued            OrderEventType = "order_issued"
	OrderReturnedByClient  OrderEventType = "order_returned_by_client"
)

func (e OrderEventType) valid() bool {
	switch e {
	case OrderAccepted, OrderReturnedToCourier, OrderIssued, OrderReturnedByClient:
		return true
	default:
		return false
	}
}
