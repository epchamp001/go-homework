package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOrderEvent_Validate(t *testing.T) {
	t.Parallel()

	utcNow := time.Now().UTC()

	// будем копировать и портить для негативных сценариев
	baseEvt := OrderEvent{
		EventID:   uuid.New(),
		EventType: OrderAccepted,
		Timestamp: utcNow,
		Actor: Actor{
			Type: ActorCourier,
			ID:   42,
		},
		Order: OrderSummary{
			ID:     "order-uuid",
			UserID: "user-uuid",
			Status: "accepted",
		},
		Source: "pvz-api",
	}

	tests := []struct {
		name     string
		prepare  func(e *OrderEvent)
		wantErr  assert.ErrorAssertionFunc
		expError error
	}{
		{
			name:    "ValidEvent",
			wantErr: assert.NoError,
		},
		{
			name: "NilUUID",
			prepare: func(e *OrderEvent) {
				e.EventID = uuid.Nil
			},
			wantErr:  assert.Error,
			expError: ErrInvalidUUID,
		},
		{
			name: "NonUTC",
			prepare: func(e *OrderEvent) {
				e.Timestamp = time.Now()
			},
			wantErr:  assert.Error,
			expError: ErrInvalidTimestamp,
		},
		{
			name: "InvalidActorType",
			prepare: func(e *OrderEvent) {
				e.Actor.Type = "admin"
			},
			wantErr:  assert.Error,
			expError: ErrInvalidActorType,
		},
		{
			name: "InvalidEventType",
			prepare: func(e *OrderEvent) {
				e.EventType = "wrong_type"
			},
			wantErr:  assert.Error,
			expError: ErrInvalidEventType,
		},
		{
			name: "EmptyIDs",
			prepare: func(e *OrderEvent) {
				e.Actor.ID = 0
			},
			wantErr:  assert.Error,
			expError: ErrEmptyIDs,
		},
		{
			name: "StatusMismatch",
			prepare: func(e *OrderEvent) {
				e.Order.Status = "issued"
			},
			wantErr:  assert.Error,
			expError: ErrEventTypeMismatch,
		},
		{
			name: "WrongSource",
			prepare: func(e *OrderEvent) {
				e.Source = "mobile-app"
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evt := baseEvt
			if tt.prepare != nil {
				tt.prepare(&evt)
			}

			err := evt.Validate()
			tt.wantErr(t, err)
			if tt.expError != nil {
				assert.ErrorIs(t, err, tt.expError)
			}
		})
	}
}

func Test_matchStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		eventType OrderEventType
		status    string
		want      bool
	}{
		{
			name:      "OrderAccepted_Match",
			eventType: OrderAccepted,
			status:    "accepted",
			want:      true,
		},
		{
			name:      "OrderAccepted_NoMatch",
			eventType: OrderAccepted,
			status:    "wrong_status",
			want:      false,
		},
		{
			name:      "OrderReturnedToCourier_Match",
			eventType: OrderReturnedToCourier,
			status:    "returned_to_courier",
			want:      true,
		},
		{
			name:      "OrderReturnedToCourier_NoMatch",
			eventType: OrderReturnedToCourier,
			status:    "returned",
			want:      false,
		},
		{
			name:      "OrderIssued_Match",
			eventType: OrderIssued,
			status:    "issued",
			want:      true,
		},
		{
			name:      "OrderIssued_NoMatch",
			eventType: OrderIssued,
			status:    "accepted",
			want:      false,
		},
		{
			name:      "OrderReturnedByClient_Match",
			eventType: OrderReturnedByClient,
			status:    "returned_by_client",
			want:      true,
		},
		{
			name:      "OrderReturnedByClient_NoMatch",
			eventType: OrderReturnedByClient,
			status:    "returned_to_courier",
			want:      false,
		},
		{
			name:      "UnknownEventType",
			eventType: OrderEventType("unknown_event"),
			status:    "whatever",
			want:      false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := matchStatus(tt.eventType, tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}
