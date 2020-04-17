package server

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestPruningOptions(t *testing.T) {
	tests := []struct {
		name        string
		paramInit   func()
		returnsErr  bool
		expectedErr error
	}{
		{
			name:        "default",
			paramInit:   func() {},
			returnsErr:  false,
			expectedErr: nil,
		},
		{
			name:        "unknown strategy",
			paramInit:   func() { viper.Set(flagPruning, "unknown") },
			returnsErr:  true,
			expectedErr: fmt.Errorf("unknown pruning strategy unknown"),
		},
		{
			name: "only keep-every provided",
			paramInit: func() {
				viper.Set(flagPruning, "custom")
				viper.Set(flagPruningKeepEvery, 12345)
			},
			returnsErr:  false,
			expectedErr: nil,
		},
		{
			name: "only snapshot-every provided",
			paramInit: func() {
				viper.Set(flagPruning, "custom")
				viper.Set(flagPruningSnapshotEvery, 12345)
			},
			returnsErr:  true,
			expectedErr: fmt.Errorf("invalid granular options"),
		},
		{
			name: "pruning flag with other granular options 3",
			paramInit: func() {
				viper.Set(flagPruning, "custom")
				viper.Set(flagPruningKeepEvery, 1234)
				viper.Set(flagPruningSnapshotEvery, 1234)
			},
			returnsErr:  false,
			expectedErr: nil,
		},
		{
			name: "nothing strategy",
			paramInit: func() {
				viper.Set(flagPruning, "nothing")
			},
			returnsErr:  false,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.SetDefault(flagPruning, "syncable")
			startCommand := StartCmd(nil, nil)
			tt.paramInit()
			err := startCommand.PreRunE(startCommand, nil)

			if tt.returnsErr {
				require.EqualError(t, err, tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
