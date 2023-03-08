package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func mkper(length int64, amount int64) Period {
	return Period{
		Length: length,
		Amount: sdk.NewCoins(sdk.NewInt64Coin("test", amount)),
	}
}

func TestReadSchedule(t *testing.T) {
	periods := []Period{
		mkper(10, 10), mkper(20, 20), mkper(40, 40),
	}
	total := sdk.NewCoins(sdk.NewInt64Coin("test", 70))
	for _, tt := range []struct {
		time int64
		want int64
	}{
		{0, 0},
		{100, 0},
		{105, 0},
		{110, 10},
		{120, 10},
		{130, 30},
		{150, 30},
		{170, 70},
		{180, 70},
	} {
		gotCoins := ReadSchedule(100, 170, periods, total, tt.time)
		got := gotCoins.AmountOf("test").Int64()
		if got != tt.want {
			t.Errorf("ReadSchedule at %d want %d, got %d", tt.time, tt.want, got)
		}
	}
}

func TestDisjunctPeriods(t *testing.T) {
	for _, tt := range []struct {
		name      string
		startP    int64
		p         []Period
		startQ    int64
		q         []Period
		wantStart int64
		wantEnd   int64
		want      []Period
	}{
		{
			name:      "empty_empty",
			startP:    0,
			p:         []Period{},
			startQ:    0,
			q:         []Period{},
			wantStart: 0,
			want:      []Period{},
		},
		{
			name:      "some_empty",
			startP:    -123,
			p:         []Period{mkper(45, 8), mkper(67, 13)},
			startQ:    -124,
			q:         []Period{},
			wantStart: -124,
			wantEnd:   -11,
			want:      []Period{mkper(46, 8), mkper(67, 13)},
		},
		{
			name:      "one_one",
			startP:    0,
			p:         []Period{mkper(12, 34)},
			startQ:    0,
			q:         []Period{mkper(25, 68)},
			wantStart: 0,
			wantEnd:   25,
			want:      []Period{mkper(12, 34), mkper(13, 68)},
		},
		{
			name:      "tied",
			startP:    12,
			p:         []Period{mkper(24, 3)},
			startQ:    0,
			q:         []Period{mkper(36, 7)},
			wantStart: 0,
			wantEnd:   36,
			want:      []Period{mkper(36, 10)},
		},
		{
			name:      "residual",
			startP:    105,
			p:         []Period{mkper(45, 309), mkper(80, 243), mkper(30, 401)},
			startQ:    120,
			q:         []Period{mkper(40, 823)},
			wantStart: 105,
			wantEnd:   260,
			want:      []Period{mkper(45, 309), mkper(10, 823), mkper(70, 243), mkper(30, 401)},
		},
		{
			name:      "typical",
			startP:    1000,
			p:         []Period{mkper(100, 25), mkper(100, 25), mkper(100, 25), mkper(100, 25)},
			startQ:    1200,
			q:         []Period{mkper(100, 10), mkper(100, 10), mkper(100, 10), mkper(100, 10)},
			wantStart: 1000,
			wantEnd:   1600,
			want:      []Period{mkper(100, 25), mkper(100, 25), mkper(100, 35), mkper(100, 35), mkper(100, 10), mkper(100, 10)},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Function is commutative in its arguments, so get two tests
			// for the price of one.  TODO: sub-t.Run() for distinct names.
			for i := 0; i < 2; i++ {
				var gotStart, gotEnd int64
				var got []Period
				if i == 0 {
					gotStart, gotEnd, got = DisjunctPeriods(tt.startP, tt.startQ, tt.p, tt.q)
				} else {
					gotStart, gotEnd, got = DisjunctPeriods(tt.startQ, tt.startP, tt.q, tt.p)
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

func TestConjunctPeriods(t *testing.T) {
	for _, tt := range []struct {
		name      string
		startP    int64
		p         []Period
		startQ    int64
		q         []Period
		wantStart int64
		wantEnd   int64
		want      []Period
	}{
		{
			name:      "empty_empty",
			startP:    0,
			p:         []Period{},
			startQ:    0,
			q:         []Period{},
			wantStart: 0,
			wantEnd:   0,
			want:      []Period{},
		},
		{
			name:      "some_empty",
			startP:    -123,
			p:         []Period{mkper(45, 8), mkper(67, 13)},
			startQ:    -124,
			q:         []Period{},
			wantStart: -124,
			wantEnd:   -124,
			want:      []Period{},
		},
		{
			name:      "one_one",
			startP:    0,
			p:         []Period{mkper(12, 34)},
			startQ:    0,
			q:         []Period{mkper(25, 68)},
			wantStart: 0,
			wantEnd:   25,
			want:      []Period{mkper(25, 34)},
		},
		{
			name:      "tied",
			startP:    12,
			p:         []Period{mkper(24, 3)},
			startQ:    0,
			q:         []Period{mkper(36, 7)},
			wantStart: 0,
			wantEnd:   36,
			want:      []Period{mkper(36, 3)},
		},
		{
			name:      "residual",
			startP:    105,
			p:         []Period{mkper(45, 309), mkper(80, 243), mkper(30, 401)},
			startQ:    120,
			q:         []Period{mkper(40, 823)},
			wantStart: 105,
			wantEnd:   260,
			want:      []Period{mkper(55, 309), mkper(70, 243), mkper(30, 271)},
		},
		{
			name:      "overlap",
			startP:    1000,
			p:         []Period{mkper(100, 25), mkper(100, 25), mkper(100, 25), mkper(100, 25)},
			startQ:    1200,
			q:         []Period{mkper(100, 10), mkper(100, 10), mkper(100, 10), mkper(100, 10)},
			wantStart: 1000,
			wantEnd:   1600,
			want:      []Period{mkper(300, 10), mkper(100, 10), mkper(100, 10), mkper(100, 10)},
		},
		{
			name:      "clip",
			startP:    100,
			p:         []Period{mkper(10, 1), mkper(10, 1), mkper(10, 1), mkper(10, 1), mkper(10, 1)},
			startQ:    100,
			q:         []Period{mkper(1, 3)},
			wantStart: 100,
			wantEnd:   130,
			want:      []Period{mkper(10, 1), mkper(10, 1), mkper(10, 1)},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Function is commutative in its arguments, so get two tests
			// for the price of one.  TODO: sub-t.Run() for distinct names.
			for i := 0; i < 2; i++ {
				var gotStart, gotEnd int64
				var got []Period
				if i == 0 {
					gotStart, gotEnd, got = ConjunctPeriods(tt.startP, tt.startQ, tt.p, tt.q)
				} else {
					gotStart, gotEnd, got = ConjunctPeriods(tt.startQ, tt.startP, tt.q, tt.p)
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

func TestAlignSchedules(t *testing.T) {
	p := []Period{mkper(10, 50), mkper(30, 7)}
	q := []Period{mkper(40, 6), mkper(10, 8), mkper(5, 3)}
	start, end := AlignSchedules(100, 200, p, q)
	if start != 100 {
		t.Errorf("want start 100, got %d", start)
	}
	if end != 255 {
		t.Errorf("want end 255, got %d", end)
	}
	if p[0].Length != 10 {
		t.Errorf("want p first length unchanged, got %d", p[0].Length)
	}
	if q[0].Length != 140 {
		t.Errorf("want q first length 140, got %d", q[0].Length)
	}
}
