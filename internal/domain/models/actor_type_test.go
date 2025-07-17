package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActorType_valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    ActorType
		want bool
	}{
		{name: "Client", a: ActorClient, want: true},
		{name: "Courier", a: ActorCourier, want: true},
		{name: "Empty", a: ActorType(""), want: false},
		{name: "Unknown", a: ActorType("admin"), want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.a.valid()
			assert.Equal(t, tt.want, got)
		})
	}
}
