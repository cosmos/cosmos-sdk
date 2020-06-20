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
			name: flagPruning,
			initParams: func() {
				viper.Set(flagPruning, types.PruningOptionNothing)
			},
			expectedOptions: types.PruneNothing,
		},
		{
			name: "custom pruning options",
			initParams: func() {
				viper.Set(flagPruning, types.PruningOptionCustom)
				viper.Set(flagPruningKeepRecent, 1234)
				viper.Set(flagPruningKeepEvery, 4321)
				viper.Set(flagPruningInterval, 10)
			},
			expectedOptions: types.PruningOptions{
				KeepRecent: 1234,
				KeepEvery:  4321,
				Interval:   10,
			},
		},
		{
			name:            "default",
			initParams:      func() {},
			expectedOptions: types.PruneDefault,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(j *testing.T) {
			viper.Reset()
			viper.SetDefault(flagPruning, "syncable")
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
