package server

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	pruningTypes "github.com/cosmos/cosmos-sdk/pruning/types"
)

func TestGetPruningOptionsFromFlags(t *testing.T) {
	tests := []struct {
		name            string
		initParams      func() *viper.Viper
		expectedOptions *pruningTypes.PruningOptions
		wantErr         bool
	}{
		{
			name: FlagPruning,
			initParams: func() *viper.Viper {
				v := viper.New()
				v.Set(FlagPruning, pruningTypes.PruningOptionNothing)
				return v
			},
			expectedOptions: pruningTypes.NewPruningOptions(pruningTypes.PruningNothing),
		},
		{
			name: "custom pruning options",
			initParams: func() *viper.Viper {
				v := viper.New()
				v.Set(FlagPruning, pruningTypes.PruningOptionCustom)
				v.Set(FlagPruningKeepRecent, 1234)
				v.Set(FlagPruningInterval, 10)

				return v
			},
			expectedOptions: pruningTypes.NewCustomPruningOptions(1234, 10),
		},
		{
			name: pruningTypes.PruningOptionDefault,
			initParams: func() *viper.Viper {
				v := viper.New()
				v.Set(FlagPruning, pruningTypes.PruningOptionDefault)
				return v
			},
			expectedOptions: pruningTypes.NewPruningOptions(pruningTypes.PruningDefault),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(j *testing.T) {
			viper.Reset()
			viper.SetDefault(FlagPruning, pruningTypes.PruningOptionDefault)
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
