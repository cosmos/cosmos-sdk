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
		initParams      func() *viper.Viper
		expectedOptions types.PruningOptions
		wantErr         bool
	}{
		{
			name: FlagPruning,
			initParams: func() *viper.Viper {
				v := viper.New()
				v.Set(FlagPruning, types.PruningOptionNothing)
				return v
			},
			expectedOptions: types.PruneNothing,
		},
		{
			name: "custom pruning options",
			initParams: func() *viper.Viper {
				v := viper.New()
				v.Set(FlagPruning, types.PruningOptionCustom)
				v.Set(FlagPruningKeepRecent, 1234)
				v.Set(FlagPruningKeepEvery, 4321)
				v.Set(FlagPruningInterval, 10)

				return v
			},
			expectedOptions: types.PruningOptions{
				KeepRecent: 1234,
				KeepEvery:  4321,
				Interval:   10,
			},
		},
		{
			name: types.PruningOptionDefault,
			initParams: func() *viper.Viper {
				v := viper.New()
				v.Set(FlagPruning, types.PruningOptionDefault)
				return v
			},
			expectedOptions: types.PruneDefault,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(j *testing.T) {
			viper.Reset()
			viper.SetDefault(FlagPruning, types.PruningOptionDefault)
			v := tt.initParams()

			opts, err := GetPruningOptionsFromFlags(v)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.Equal(t, tt.expectedOptions, opts)
		})
	}
}
