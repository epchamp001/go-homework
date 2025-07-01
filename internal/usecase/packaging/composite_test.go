package packaging

import (
	"github.com/stretchr/testify/assert"
	"pvz-cli/internal/domain/models"
	"testing"
)

func TestCompositeStrategy_Validate(t *testing.T) {
	t.Parallel()

	type args struct {
		parts   []PackagingStrategy
		weight  float64
		wantErr assert.ErrorAssertionFunc
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "NegativeWeight",
			args: args{
				parts:   []PackagingStrategy{NewFilmStrategy(), NewFilmStrategy()},
				weight:  -1,
				wantErr: assert.Error,
			},
		},
		{
			name: "ZeroWeight",
			args: args{
				parts:   []PackagingStrategy{NewFilmStrategy(), NewFilmStrategy()},
				weight:  0,
				wantErr: assert.Error,
			},
		},
		{
			name: "PositiveWeight",
			args: args{
				parts:   []PackagingStrategy{NewFilmStrategy(), NewFilmStrategy()},
				weight:  1.0,
				wantErr: assert.NoError,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			strat := NewCompositeStrategy(tt.args.parts...)
			err := strat.Validate(tt.args.weight)

			tt.args.wantErr(t, err)
		})
	}
}

func TestCompositeStrategy_Surcharge(t *testing.T) {
	t.Parallel()

	type args struct {
		parts []PackagingStrategy
		want  models.PriceKopecks
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "FilmPlusFilm",
			args: args{
				parts: []PackagingStrategy{NewFilmStrategy(), NewFilmStrategy()},
				want:  models.SurchargeFilm * 2,
			},
		},
		{
			name: "FilmPlusBag",
			args: args{
				parts: []PackagingStrategy{NewFilmStrategy(), NewBagStrategy()},
				want:  models.SurchargeFilm + models.SurchargeBag,
			},
		},
		{
			name: "EmptyParts",
			args: args{
				parts: []PackagingStrategy{},
				want:  0,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			strat := NewCompositeStrategy(tt.args.parts...)
			got := strat.Surcharge()
			assert.Equal(t, tt.args.want, got)
		})
	}
}
