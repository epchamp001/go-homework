package models

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

type Actor struct {
	Type ActorType `json:"type"`
	ID   int64     `json:"id"`
}

type OrderSummary struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Status string `json:"status"`
}

type OrderEvent struct {
	EventID   uuid.UUID      `json:"event_id"`
	EventType OrderEventType `json:"event_type"`
	Timestamp time.Time      `json:"timestamp"` // всегда в UTC
	Actor     Actor          `json:"actor"`
	Order     OrderSummary   `json:"order"`
	Source    string         `json:"source"` // всегда "pvz-api"
}

func NewOrderEvent(eventType OrderEventType, orderID, orderStatus, userID string, actor Actor) (*OrderEvent, error) {
	evt := &OrderEvent{
		EventID:   uuid.New(),
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Actor:     actor,
		Order: OrderSummary{
			ID:     orderID,
			UserID: userID,
			Status: orderStatus,
		},
		Source: "pvz-api",
	}
	return evt, evt.Validate()
}

var (
	ErrInvalidUUID       = errors.New("invalid uuid")
	ErrInvalidTimestamp  = errors.New("timestamp must be UTC")
	ErrInvalidActorType  = errors.New("actor.type must be client|courier")
	ErrInvalidEventType  = errors.New("unknown event_type")
	ErrEventTypeMismatch = errors.New("event_type does not match order.status")
	ErrEmptyIDs          = errors.New("actor.id, order.id, order.user_id must be filled")
)

func (e *OrderEvent) Validate() error {
	switch {
	case e.EventID == uuid.Nil:
		return ErrInvalidUUID
	case e.Timestamp.Location() != time.UTC:
		return ErrInvalidTimestamp
	case !e.Actor.Type.valid():
		return ErrInvalidActorType
	case !e.EventType.valid():
		return ErrInvalidEventType
	case e.Actor.ID == 0 || e.Order.ID == "" || e.Order.UserID == "":
		return ErrEmptyIDs
	case !matchStatus(e.EventType, e.Order.Status):
		return ErrEventTypeMismatch
	case e.Source != "pvz-api":
		return errors.New("source must be pvz-api")
	default:
		return nil
	}
}

func matchStatus(t OrderEventType, s string) bool {
	switch t {
	case OrderAccepted:
		return s == "accepted"
	case OrderReturnedToCourier:
		return s == "returned_to_courier"
	case OrderIssued:
		return s == "issued"
	case OrderReturnedByClient:
		return s == "returned_by_client"
	default:
		return false
	}
}
