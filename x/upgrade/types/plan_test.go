package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestPlanString(t *testing.T) {
	cases := map[string]struct {
		p      Plan
		expect string
	}{
		"with time": {
			p: Plan{
				Name: "due_time",
				Info: "https://foo.bar",
				Time: mustParseTime("2019-07-08T11:33:55Z"),
			},
			expect: "Upgrade Plan\n  Name: due_time\n  Time: 2019-07-08T11:33:55Z\n  Info: https://foo.bar",
		},
		"with height": {
			p: Plan{
				Name:   "by height",
				Info:   "https://foo.bar/baz",
				Height: 7890,
			},
			expect: "Upgrade Plan\n  Name: by height\n  Height: 7890\n  Info: https://foo.bar/baz",
		},
		"neither": {
			p: Plan{
				Name: "almost-empty",
			},
			expect: "Upgrade Plan\n  Name: almost-empty\n  Height: 0\n  Info: ",
		},
	}

	for name, tc := range cases {
		tc := tc // copy to local variable for scopelint
		t.Run(name, func(t *testing.T) {
			s := tc.p.String()
			require.Equal(t, tc.expect, s)
		})
	}
}

func TestPlanValid(t *testing.T) {
	cases := map[string]struct {
		p     Plan
		valid bool
	}{
		"proper": {
			p: Plan{
				Name: "all-good",
				Info: "some text here",
				Time: mustParseTime("2019-07-08T11:33:55Z"),
			},
			valid: true,
		},
		"proper by height": {
			p: Plan{
				Name:   "all-good",
				Height: 123450000,
			},
			valid: true,
		},
		"no name": {
			p: Plan{
				Height: 123450000,
			},
		},
		"no due at": {
			p: Plan{
				Name: "missing",
				Info: "important",
			},
		},
		"negative height": {
			p: Plan{
				Name:   "minus",
				Height: -12345,
			},
		},
	}

	for name, tc := range cases {
		tc := tc // copy to local variable for scopelint
		t.Run(name, func(t *testing.T) {
			err := tc.p.ValidateBasic()
			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}

}

func TestShouldExecute(t *testing.T) {
	cases := map[string]struct {
		p         Plan
		ctxTime   time.Time
		ctxHeight int64
		expected  bool
	}{
		"past time": {
			p: Plan{
				Name: "do-good",
				Info: "some text here",
				Time: mustParseTime("2019-07-08T11:33:55Z"),
			},
			ctxTime:   mustParseTime("2019-07-08T11:32:00Z"),
			ctxHeight: 100000,
			expected:  false,
		},
		"on time": {
			p: Plan{
				Name: "do-good",
				Time: mustParseTime("2019-07-08T11:33:55Z"),
			},
			ctxTime:   mustParseTime("2019-07-08T11:33:55Z"),
			ctxHeight: 100000,
			expected:  true,
		},
		"future time": {
			p: Plan{
				Name: "do-good",
				Time: mustParseTime("2019-07-08T11:33:55Z"),
			},
			ctxTime:   mustParseTime("2019-07-08T11:33:57Z"),
			ctxHeight: 100000,
			expected:  true,
		},
		"past height": {
			p: Plan{
				Name:   "do-good",
				Height: 1234,
			},
			ctxTime:   mustParseTime("2019-07-08T11:32:00Z"),
			ctxHeight: 1000,
			expected:  false,
		},
		"on height": {
			p: Plan{
				Name:   "do-good",
				Height: 1234,
			},
			ctxTime:   mustParseTime("2019-07-08T11:32:00Z"),
			ctxHeight: 1234,
			expected:  true,
		},
		"future height": {
			p: Plan{
				Name:   "do-good",
				Height: 1234,
			},
			ctxTime:   mustParseTime("2019-07-08T11:32:00Z"),
			ctxHeight: 1235,
			expected:  true,
		},
	}

	for name, tc := range cases {
		tc := tc // copy to local variable for scopelint
		t.Run(name, func(t *testing.T) {
			ctx := sdk.NewContext(nil, abci.Header{Height: tc.ctxHeight, Time: tc.ctxTime}, false, log.NewNopLogger())
			should := tc.p.ShouldExecute(ctx)
			assert.Equal(t, tc.expected, should)
		})
	}
}
