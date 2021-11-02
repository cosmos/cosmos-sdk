package vesting

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

func TestMergePeriods(t *testing.T) {
	mkper := func(length int64, amount int64) types.Period {
		return types.Period{
			Length: length,
			Amount: sdk.NewCoins(sdk.NewInt64Coin("test", amount)),
		}
	}
	for _, tt := range []struct {
		name      string
		startP    int64
		p         []types.Period
		startQ    int64
		q         []types.Period
		wantStart int64
		wantEnd   int64
		want      []types.Period
	}{
		{
			name:      "empty_empty",
			startP:    0,
			p:         []types.Period{},
			startQ:    0,
			q:         []types.Period{},
			wantStart: 0,
			want:      []types.Period{},
		},
		{
			name:      "some_empty",
			startP:    -123,
			p:         []types.Period{mkper(45, 8), mkper(67, 13)},
			startQ:    -124,
			q:         []types.Period{},
			wantStart: -124,
			wantEnd:   -11,
			want:      []types.Period{mkper(46, 8), mkper(67, 13)},
		},
		{
			name:      "one_one",
			startP:    0,
			p:         []types.Period{mkper(12, 34)},
			startQ:    0,
			q:         []types.Period{mkper(25, 68)},
			wantStart: 0,
			wantEnd:   25,
			want:      []types.Period{mkper(12, 34), mkper(13, 68)},
		},
		{
			name:      "tied",
			startP:    12,
			p:         []types.Period{mkper(24, 3)},
			startQ:    0,
			q:         []types.Period{mkper(36, 7)},
			wantStart: 0,
			wantEnd:   36,
			want:      []types.Period{mkper(36, 10)},
		},
		{
			name:      "residual",
			startP:    105,
			p:         []types.Period{mkper(45, 309), mkper(80, 243), mkper(30, 401)},
			startQ:    120,
			q:         []types.Period{mkper(40, 823)},
			wantStart: 105,
			wantEnd:   260,
			want:      []types.Period{mkper(45, 309), mkper(10, 823), mkper(70, 243), mkper(30, 401)},
		},
		{
			name:      "typical",
			startP:    1000,
			p:         []types.Period{mkper(100, 25), mkper(100, 25), mkper(100, 25), mkper(100, 25)},
			startQ:    1200,
			q:         []types.Period{mkper(100, 10), mkper(100, 10), mkper(100, 10), mkper(100, 10)},
			wantStart: 1000,
			wantEnd:   1600,
			want:      []types.Period{mkper(100, 25), mkper(100, 25), mkper(100, 35), mkper(100, 35), mkper(100, 10), mkper(100, 10)},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Function is commutative in its arguments, so get two tests
			// for the price of one.  TODO: sub-t.Run() for distinct names.
			for i := 0; i < 2; i++ {
				var gotStart, gotEnd int64
				var got []types.Period
				if i == 0 {
					gotStart, gotEnd, got = mergePeriods(tt.startP, tt.startQ, tt.p, tt.q)
				} else {
					gotStart, gotEnd, got = mergePeriods(tt.startQ, tt.startP, tt.q, tt.p)
				}
				if gotStart != tt.wantStart {
					t.Errorf("wrong start time: got %d, want %d", gotStart, tt.wantStart)
				}
				if gotEnd != tt.wantEnd {
					t.Errorf("wrong end time: got %d, want %d", gotEnd, tt.wantEnd)
				}
				if len(got) != len(tt.want) {
					t.Fatalf("wrong number of periods: got %v, want %v", got, tt.want)
				}
				for i, gotPeriod := range got {
					wantPeriod := tt.want[i]
					if gotPeriod.Length != wantPeriod.Length {
						t.Errorf("period %d length: got %d, want %d", i, gotPeriod.Length, wantPeriod.Length)
					}
					if !gotPeriod.Amount.IsEqual(wantPeriod.Amount) {
						t.Errorf("period %d amount: got %v, want %v", i, gotPeriod.Amount, wantPeriod.Amount)
					}
				}
			}
		})
	}
}
