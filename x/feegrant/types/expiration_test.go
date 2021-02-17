package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

func TestExpiresAt(t *testing.T) {
	now := time.Now()

	cases := map[string]struct {
		example types.ExpiresAt
		valid   bool
		zero    bool
		before  types.ExpiresAt
		after   types.ExpiresAt
	}{
		"basic": {
			example: types.ExpiresAtHeight(100),
			valid:   true,
			before:  types.ExpiresAtHeight(50),
			after:   types.ExpiresAtHeight(122),
		},
		"zero": {
			example: types.ExpiresAt{},
			zero:    true,
			valid:   true,
			before:  types.ExpiresAtHeight(1),
		},
		"match height": {
			example: types.ExpiresAtHeight(1000),
			valid:   true,
			before:  types.ExpiresAtHeight(999),
			after:   types.ExpiresAtHeight(1000),
		},
		"match time": {
			example: types.ExpiresAtTime(now),
			valid:   true,
			before:  types.ExpiresAtTime(now.Add(-1 * time.Second)),
			after:   types.ExpiresAtTime(now.Add(1 * time.Second)),
		},
	}

	for name, stc := range cases {
		tc := stc // to make scopelint happy
		t.Run(name, func(t *testing.T) {
			err := tc.example.ValidateBasic()
			assert.Equal(t, tc.zero, tc.example.Undefined())
			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if !tc.before.Undefined() {
				assert.Equal(t, false, tc.example.IsExpired(tc.before.GetTime(), tc.before.GetHeight()))
			}
			if !tc.after.Undefined() {
				assert.Equal(t, true, tc.example.IsExpired(tc.after.GetTime(), tc.after.GetHeight()))
			}
		})
	}
}

func TestDurationValid(t *testing.T) {
	now := time.Now()

	cases := map[string]struct {
		period       types.Duration
		valid        bool
		compatible   types.ExpiresAt
		incompatible types.ExpiresAt
	}{
		"basic height": {
			period:       types.BlockDuration(100),
			valid:        true,
			compatible:   types.ExpiresAtHeight(50),
			incompatible: types.ExpiresAtTime(now),
		},
		"basic time": {
			period:       types.ClockDuration(time.Hour),
			valid:        true,
			compatible:   types.ExpiresAtTime(now),
			incompatible: types.ExpiresAtHeight(50),
		},
		"zero": {
			period: types.Duration{},
			valid:  false,
		},
		"negative clock": {
			period: types.ClockDuration(-1 * time.Hour),
			valid:  false,
		},
	}

	for name, stc := range cases {
		tc := stc // to make scopelint happy
		t.Run(name, func(t *testing.T) {
			err := tc.period.ValidateBasic()
			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, true, tc.compatible.IsCompatible(tc.period))
			assert.Equal(t, false, tc.incompatible.IsCompatible(tc.period))
		})
	}
}

func TestDurationStep(t *testing.T) {
	now := time.Now()

	cases := map[string]struct {
		expires types.ExpiresAt
		period  types.Duration
		valid   bool
		result  types.ExpiresAt
	}{
		"add height": {
			expires: types.ExpiresAtHeight(789),
			period:  types.BlockDuration(100),
			valid:   true,
			result:  types.ExpiresAtHeight(889),
		},
		"add time": {
			expires: types.ExpiresAtTime(now),
			period:  types.ClockDuration(time.Hour),
			valid:   true,
			result:  types.ExpiresAtTime(now.Add(time.Hour)),
		},
		"mismatch": {
			expires: types.ExpiresAtHeight(789),
			period:  types.ClockDuration(time.Hour),
			valid:   false,
		},
	}

	for name, stc := range cases {
		tc := stc // to make scopelint happy
		t.Run(name, func(t *testing.T) {
			err := tc.period.ValidateBasic()
			require.NoError(t, err)
			err = tc.expires.ValidateBasic()
			require.NoError(t, err)

			next, err := tc.expires.Step(tc.period)
			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.result, next)
		})
	}
}
