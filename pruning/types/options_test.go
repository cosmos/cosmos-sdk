package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPruningOptions_Validate(t *testing.T) {
	testCases := []struct {
		opts   *PruningOptions
		expectErr  error
	}{
		{NewPruningOptions(PruningDefault), nil},
		{NewPruningOptions(PruningEverything), nil},
		{NewPruningOptions(PruningNothing), nil},
		{NewCustomPruningOptions(2, 10), nil},
		{NewCustomPruningOptions(100, 15), nil},
		{NewCustomPruningOptions(1, 10), ErrPruningKeepRecentTooSmall},
		{NewCustomPruningOptions(2, 9), ErrPruningIntervalTooSmall},
		{NewCustomPruningOptions(2, 0), ErrPruningIntervalZero},
		{NewCustomPruningOptions(2, 0), ErrPruningIntervalZero},
	}

	for _, tc := range testCases {
		err := tc.opts.Validate()
		require.Equal(t, tc.expectErr, err, "options: %v, err: %s", tc.opts, err)
	}
}
