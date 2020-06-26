package server

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/types"
)

func TestGetPruningOptionsFromFlags(t *testing.T) {
	tests := []struct {
		name            string
		initParams      func()
		expectedOptions types.PruningOptions
		wantErr         bool
	}{
		{
			name: FlagPruning,
			initParams: func() {
				viper.Set(FlagPruning, types.PruningOptionNothing)
			},
			expectedOptions: types.PruneNothing,
		},
		{
			name: "custom pruning options",
			initParams: func() {
				viper.Set(FlagPruning, types.PruningOptionCustom)
				viper.Set(FlagPruningKeepRecent, 1234)
				viper.Set(FlagPruningKeepEvery, 4321)
				viper.Set(FlagPruningInterval, 10)
			},
			expectedOptions: types.PruningOptions{
				KeepRecent: 1234,
				KeepEvery:  4321,
				Interval:   10,
			},
		},
		{
			name:            types.PruningOptionDefault,
			initParams:      func() {},
			expectedOptions: types.PruneDefault,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(j *testing.T) {
			viper.Reset()
			viper.SetDefault(FlagPruning, types.PruningOptionDefault)
			tt.initParams()

			opts, err := GetPruningOptionsFromFlags()
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.Equal(t, tt.expectedOptions, opts)
		})
	}
}
