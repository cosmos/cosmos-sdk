package server

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestPruningOptions(t *testing.T) {
	startCommand := StartCmd(nil, nil)

	tests := []struct {
		name        string
		paramInit   func()
		returnsErr  bool
		expectedErr error
	}{
		{
			name:        "none set, returns nil and will use default from flags",
			paramInit:   func() {},
			returnsErr:  false,
			expectedErr: nil,
		},
		{
			name: "only keep-every provided",
			paramInit: func() {
				viper.Set(flagPruningKeepEvery, 12345)
			},
			returnsErr:  true,
			expectedErr: errPruningGranularOptions,
		},
		{
			name: "only snapshot-every provided",
			paramInit: func() {
				viper.Set(flagPruningSnapshotEvery, 12345)
			},
			returnsErr:  true,
			expectedErr: errPruningGranularOptions,
		},
		{
			name: "pruning flag with other granular options 1",
			paramInit: func() {
				viper.Set(flagPruning, "set")
				viper.Set(flagPruningSnapshotEvery, 1234)
			},
			returnsErr:  true,
			expectedErr: errPruningWithGranularOptions,
		},
		{
			name: "pruning flag with other granular options 2",
			paramInit: func() {
				viper.Set(flagPruning, "set")
				viper.Set(flagPruningKeepEvery, 1234)
			},
			returnsErr:  true,
			expectedErr: errPruningWithGranularOptions,
		},
		{
			name: "pruning flag with other granular options 3",
			paramInit: func() {
				viper.Set(flagPruning, "set")
				viper.Set(flagPruningKeepEvery, 1234)
				viper.Set(flagPruningSnapshotEvery, 1234)
			},
			returnsErr:  true,
			expectedErr: errPruningWithGranularOptions,
		},
		{
			name: "only prunning set",
			paramInit: func() {
				viper.Set(flagPruning, "set")
			},
			returnsErr:  false,
			expectedErr: nil,
		},
		{
			name: "only granular set",
			paramInit: func() {
				viper.Set(flagPruningSnapshotEvery, 12345)
				viper.Set(flagPruningKeepEvery, 12345)
			},
			returnsErr:  false,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			tt.paramInit()

			err := startCommand.PreRunE(nil, nil)

			if tt.returnsErr {
				require.EqualError(t, err, tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
