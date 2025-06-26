package packaging

import (
	"github.com/stretchr/testify/assert"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"testing"
)

func TestBoxStrategy_Validate(t *testing.T) {
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
			args: args{
				weight:     -1,
				wantErr:    assert.Error,
				wantErrVal: codes.ErrValidationFailed,
			},
		},
		{
			name: "ZeroWeight",
			args: args{
				weight:     0,
				wantErr:    assert.Error,
				wantErrVal: codes.ErrValidationFailed,
			},
		},
		{
			name: "TooHeavyAtLimit",
			args: args{
				weight:     30,
				wantErr:    assert.Error,
				wantErrVal: codes.ErrWeightTooHeavy,
			},
		},
		{
			name: "TooHeavyAboveLimit",
			args: args{
				weight:     31,
				wantErr:    assert.Error,
				wantErrVal: codes.ErrWeightTooHeavy,
			},
		},
		{
			name: "ValidWeight",
			args: args{
				weight:     1.0,
				wantErr:    assert.NoError,
				wantErrVal: nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			strat := NewBoxStrategy()
			err := strat.Validate(tt.args.weight)

			tt.args.wantErr(t, err)
			if tt.args.wantErrVal != nil {
				assert.Equal(t, tt.args.wantErrVal, err)
			}
		})
	}
}

func TestBoxStrategy_Surcharge(t *testing.T) {
	t.Parallel()

	strat := NewBoxStrategy()
	got := strat.Surcharge()
	assert.Equal(t, models.SurchargeBox, got)
}
