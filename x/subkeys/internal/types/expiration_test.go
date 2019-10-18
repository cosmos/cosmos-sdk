package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpiresAt(t *testing.T) {
	now := time.Now()

	cases := map[string]struct {
		example ExpiresAt
		valid   bool
		zero    bool
		before  *ExpiresAt
		after   *ExpiresAt
	}{
		"basic": {
			example: ExpiresAtHeight(100),
			valid:   true,
			before:  &ExpiresAt{Height: 50, Time: now},
			after:   &ExpiresAt{Height: 122, Time: now},
		},
		"zero": {
			example: ExpiresAt{},
			zero:    true,
			valid:   true,
			before:  &ExpiresAt{Height: 1},
		},
		"double": {
			example: ExpiresAt{Height: 100, Time: now},
			valid:   false,
		},
		"match height": {
			example: ExpiresAtHeight(1000),
			valid:   true,
			before:  &ExpiresAt{Height: 999, Time: now},
			after:   &ExpiresAt{Height: 1000, Time: now},
		},
		"match time": {
			example: ExpiresAtTime(now),
			valid:   true,
			before:  &ExpiresAt{Height: 43, Time: now.Add(-1 * time.Second)},
			after:   &ExpiresAt{Height: 76, Time: now},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.example.ValidateBasic()
			assert.Equal(t, tc.zero, tc.example.IsZero())
			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tc.before != nil {
				assert.Equal(t, false, tc.example.IsExpired(tc.before.Time, tc.before.Height))
			}
			if tc.after != nil {
				assert.Equal(t, true, tc.example.IsExpired(tc.after.Time, tc.after.Height))
			}
		})
	}
}

func TestPeriodValid(t *testing.T) {
	now := time.Now()

	cases := map[string]struct {
		period       Period
		valid        bool
		compatible   ExpiresAt
		incompatible ExpiresAt
	}{
		"basic height": {
			period:       BlockPeriod(100),
			valid:        true,
			compatible:   ExpiresAtHeight(50),
			incompatible: ExpiresAtTime(now),
		},
		"basic time": {
			period:       ClockPeriod(time.Hour),
			valid:        true,
			compatible:   ExpiresAtTime(now),
			incompatible: ExpiresAtHeight(50),
		},
		"zero": {
			period: Period{},
			valid:  false,
		},
		"double": {
			period: Period{Block: 100, Clock: time.Hour},
			valid:  false,
		},
		"negative clock": {
			period: ClockPeriod(-1 * time.Hour),
			valid:  false,
		},
		"negative block": {
			period: BlockPeriod(-5),
			valid:  false,
		},
	}

	for name, tc := range cases {
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

func TestPeriodStep(t *testing.T) {
	now := time.Now()

	cases := map[string]struct {
		expires ExpiresAt
		period  Period
		valid   bool
		result  ExpiresAt
	}{
		"add height": {
			expires: ExpiresAtHeight(789),
			period:  BlockPeriod(100),
			valid:   true,
			result:  ExpiresAtHeight(889),
		},
		"add time": {
			expires: ExpiresAtTime(now),
			period:  ClockPeriod(time.Hour),
			valid:   true,
			result:  ExpiresAtTime(now.Add(time.Hour)),
		},
		"mismatch": {
			expires: ExpiresAtHeight(789),
			period:  ClockPeriod(time.Hour),
			valid:   false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.period.ValidateBasic()
			require.NoError(t, err)
			err = tc.expires.ValidateBasic()
			require.NoError(t, err)

			err = tc.expires.Step(tc.period)
			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.result, tc.expires)
		})
	}
}
