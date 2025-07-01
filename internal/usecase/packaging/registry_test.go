package packaging

import (
	"github.com/stretchr/testify/assert"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"testing"
)

func TestNewDefaultProvider(t *testing.T) {
	t.Parallel()

	type args struct {
		typ models.PackageType
	}

	tests := []struct {
		name    string
		args    args
		wantOk  bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "None",
			args:    args{typ: models.PackageNone},
			wantOk:  true,
			wantErr: assert.NoError,
		},
		{
			name:    "Bag",
			args:    args{typ: models.PackageBag},
			wantOk:  true,
			wantErr: assert.NoError,
		},
		{
			name:    "Box",
			args:    args{typ: models.PackageBox},
			wantOk:  true,
			wantErr: assert.NoError,
		},
		{
			name:    "Film",
			args:    args{typ: models.PackageFilm},
			wantOk:  true,
			wantErr: assert.NoError,
		},
		{
			name:    "BagFilm",
			args:    args{typ: models.PackageBagFilm},
			wantOk:  true,
			wantErr: assert.NoError,
		},
		{
			name:    "BoxFilm",
			args:    args{typ: models.PackageBoxFilm},
			wantOk:  true,
			wantErr: assert.NoError,
		},
		{
			name:    "Invalid",
			args:    args{typ: models.PackageType("invalid")},
			wantOk:  false,
			wantErr: assert.Error,
		},
	}

	provider := NewDefaultProvider()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			strat, err := provider.Strategy(tt.args.typ)
			tt.wantErr(t, err)
			if tt.wantOk {
				assert.NotNil(t, strat)
			} else {
				assert.Nil(t, strat)
				assert.Equal(t, codes.ErrInvalidPackage, err)
			}
		})
	}
}
func TestDefaultProvider_Strategy(t *testing.T) {
	t.Parallel()

	type args struct {
		t models.PackageType
	}

	p := NewDefaultProvider()

	tests := []struct {
		name    string
		args    args
		wantOk  bool
		wantErr assert.ErrorAssertionFunc
	}{
		{name: "None", args: args{models.PackageNone}, wantOk: true, wantErr: assert.NoError},
		{name: "Bag", args: args{models.PackageBag}, wantOk: true, wantErr: assert.NoError},
		{name: "Box", args: args{models.PackageBox}, wantOk: true, wantErr: assert.NoError},
		{name: "Film", args: args{models.PackageFilm}, wantOk: true, wantErr: assert.NoError},
		{name: "BagFilm", args: args{models.PackageBagFilm}, wantOk: true, wantErr: assert.NoError},
		{name: "BoxFilm", args: args{models.PackageBoxFilm}, wantOk: true, wantErr: assert.NoError},
		{name: "Invalid", args: args{models.PackageType("invalid")}, wantOk: false, wantErr: assert.Error},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			strat, err := p.Strategy(tt.args.t)
			tt.wantErr(t, err)
			if tt.wantOk {
				assert.NotNil(t, strat)
			} else {
				assert.Nil(t, strat)
				assert.Equal(t, codes.ErrInvalidPackage, err)
			}
		})
	}
}
