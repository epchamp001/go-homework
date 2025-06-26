package packaging

import (
	"github.com/stretchr/testify/assert"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"testing"
)

func TestBagStrategy_Validate(t *testing.T) {
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
				weight:     10,
				wantErr:    assert.Error,
				wantErrVal: codes.ErrWeightTooHeavy,
			},
		},
		{
			name: "TooHeavyAboveLimit",
			args: args{
				weight:     15,
				wantErr:    assert.Error,
				wantErrVal: codes.ErrWeightTooHeavy,
			},
		},
		{
			name: "ValidWeight",
			args: args{
				weight:     5.0,
				wantErr:    assert.NoError,
				wantErrVal: nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			strat := NewBagStrategy()
			err := strat.Validate(tt.args.weight)

			tt.args.wantErr(t, err)
			if tt.args.wantErrVal != nil {
				assert.Equal(t, tt.args.wantErrVal, err)
			}
		})
	}
}

func TestBagStrategy_Surcharge(t *testing.T) {
	t.Parallel()

	strat := NewBagStrategy()
	got := strat.Surcharge()
	assert.Equal(t, models.SurchargeBag, got)
}
