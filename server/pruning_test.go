package server

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store"
)

func TestGetPruningOptionsFromFlags(t *testing.T) {
	tests := []struct {
		name            string
		initParams      func()
		expectedOptions store.PruningOptions
		wantErr         bool
	}{
		{
			name: "pruning",
			initParams: func() {
				viper.Set(flagPruning, store.PruningStrategyNothing)
			},
			expectedOptions: store.PruneNothing,
		},
		{
			name: "granular pruning",
			initParams: func() {
				viper.Set(flagPruning, "custom")
				viper.Set(flagPruningSnapshotEvery, 1234)
				viper.Set(flagPruningKeepEvery, 4321)
			},
			expectedOptions: store.PruningOptions{
				SnapshotEvery: 1234,
				KeepEvery:     4321,
			},
		},
		{
			name:            "default",
			initParams:      func() {},
			expectedOptions: store.PruneSyncable,
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
