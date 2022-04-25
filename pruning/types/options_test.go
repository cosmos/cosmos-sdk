package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPruningOptions_Validate(t *testing.T) {
	testCases := []struct {
		opts      PruningOptions
		expectErr error
	}{
		{NewPruningOptions(PruningDefault), nil},
		{NewPruningOptions(PruningEverything), nil},
		{NewPruningOptions(PruningNothing), nil},
		{NewCustomPruningOptions(10, 10), nil},
		{NewPruningOptions(PruningCustom), nil},
		{NewCustomPruningOptions(100, 15), nil},
		{NewCustomPruningOptions(9, 10), ErrPruningKeepRecentTooSmall},
		{NewCustomPruningOptions(10, 9), ErrPruningIntervalTooSmall},
		{NewCustomPruningOptions(10, 0), ErrPruningIntervalZero},
		{NewCustomPruningOptions(9, 0), ErrPruningIntervalZero},
	}

	for _, tc := range testCases {
		err := tc.opts.Validate()
		require.Equal(t, tc.expectErr, err, "options: %v, err: %s", tc.opts, err)
	}
}

func TestPruningOptions_GetStrategy(t *testing.T) {
	testCases := []struct {
		opts             PruningOptions
		expectedStrategy PruningStrategy
	}{
		{NewPruningOptions(PruningDefault), PruningDefault},
		{NewPruningOptions(PruningEverything), PruningEverything},
		{NewPruningOptions(PruningNothing), PruningNothing},
		{NewPruningOptions(PruningCustom), PruningDefault},
		{NewCustomPruningOptions(2, 10), PruningCustom},
	}

	for _, tc := range testCases {
		actualStrategy := tc.opts.GetPruningStrategy()
		require.Equal(t, tc.expectedStrategy, actualStrategy)
	}
}

func TestNewPruningOptionsFromString(t *testing.T) {
	testCases := []struct {
		optString string
		expect    PruningOptions
	}{
		{PruningOptionDefault, NewPruningOptions(PruningDefault)},
		{PruningOptionEverything, NewPruningOptions(PruningEverything)},
		{PruningOptionNothing, NewPruningOptions(PruningNothing)},
		{"invalid", NewPruningOptions(PruningDefault)},
	}

	for _, tc := range testCases {
		actual := NewPruningOptionsFromString(tc.optString)
		require.Equal(t, tc.expect, actual)
	}
}
