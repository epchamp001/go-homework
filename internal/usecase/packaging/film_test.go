package packaging

import (
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilmStrategy_Validate(t *testing.T) {
	t.Parallel()

	type args struct {
		weight     float64
		wantErr    assert.ErrorAssertionFunc
		wantErrVal error
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "NegativeWeight",
			args: args{weight: -1, wantErr: assert.Error, wantErrVal: codes.ErrValidationFailed},
		},
		{
			name: "ZeroWeight",
			args: args{weight: 0, wantErr: assert.Error, wantErrVal: codes.ErrValidationFailed},
		},
		{
			name: "PositiveWeight",
			args: args{weight: 0.1, wantErr: assert.NoError, wantErrVal: nil},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			strat := NewFilmStrategy()
			err := strat.Validate(tt.args.weight)

			tt.args.wantErr(t, err)
			if tt.args.wantErrVal != nil {
				assert.Equal(t, tt.args.wantErrVal, err)
			}
		})
	}
}

func TestFilmStrategy_Surcharge(t *testing.T) {
	t.Parallel()

	strat := NewFilmStrategy()
	got := strat.Surcharge()
	assert.Equal(t, models.SurchargeFilm, got)
}
