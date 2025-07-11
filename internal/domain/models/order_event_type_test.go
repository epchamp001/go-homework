package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOrderEventType_valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		e    OrderEventType
		want bool
	}{
		{name: "Accepted", e: OrderAccepted, want: true},
		{name: "ReturnedToCourier", e: OrderReturnedToCourier, want: true},
		{name: "Issued", e: OrderIssued, want: true},
		{name: "ReturnedByClient", e: OrderReturnedByClient, want: true},
		{name: "Unknown", e: OrderEventType("unknown_event"), want: false},
		{name: "Empty", e: OrderEventType(""), want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.e.valid()
			assert.Equal(t, tt.want, got)
		})
	}
}
