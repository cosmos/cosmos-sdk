package main

import (
	// "encoding/json"
	"reflect"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/client/cli"
)

const (
	billion = int64(1000 * 1000 * 1000)
)

func iso(s string) time.Time {
	t, err := parseIso(s)
	if err != nil {
		panic(err)
	}
	return t
}

func hhmm(s string) time.Time {
	t, err := time.Parse(hhmmFmt, s)
	if err != nil {
		panic(err)
	}
	return t
}

func coins(s string) sdk.Coins {
	c, err := sdk.ParseCoinsNormalized(s)
	if err != nil {
		panic(err)
	}
	return c
}

func evt(ts string, cs string) event {
	tm := iso(ts)
	c := coins(cs)
	return event{Time: tm, Coins: c}
}

func TestDivision(t *testing.T) {
	for _, tt := range []struct {
		name      string
		total     int64
		divisions int
		want      []int64
	}{
		{"zeroparts", 99, 0, nil},
		{"negparts", 99, -3, nil},
		{"negtot", -25, 7, nil},
		{"onepart", 123, 1, []int64{123}},
		{"twoparts_even", 32, 2, []int64{16, 16}},
		{"twoparts_odd", 17, 2, []int64{8, 9}},
		{"hard", 25, 7, []int64{3, 4, 3, 4, 3, 4, 4}},
		{"big", 30 * billion, 3, []int64{
			10 * billion, 10 * billion, 10 * billion,
		}},
		{"huge", billion * billion, 10, []int64{
			100 * 1000 * 1000 * billion,
			100 * 1000 * 1000 * billion,
			100 * 1000 * 1000 * billion,
			100 * 1000 * 1000 * billion,
			100 * 1000 * 1000 * billion,
			100 * 1000 * 1000 * billion,
			100 * 1000 * 1000 * billion,
			100 * 1000 * 1000 * billion,
			100 * 1000 * 1000 * billion,
			100 * 1000 * 1000 * billion,
		}},
		{"huge_hard", billion*billion + 7, 10, []int64{
			100 * 1000 * 1000 * billion,
			100*1000*1000*billion + 1,
			100*1000*1000*billion + 1,
			100 * 1000 * 1000 * billion,
			100*1000*1000*billion + 1,
			100*1000*1000*billion + 1,
			100 * 1000 * 1000 * billion,
			100*1000*1000*billion + 1,
			100*1000*1000*billion + 1,
			100*1000*1000*billion + 1,
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := divide(sdk.NewInt(tt.total), tt.divisions)
			if err != nil {
				if tt.want != nil {
					t.Fatalf("divide got error %v", err)
				}
				return
			}
			if len(got) != tt.divisions || len(got) != len(tt.want) {
				t.Fatalf("divide returned wrong size: got %d, want %d", len(got), len(tt.want))
			}
			for i := 0; i < len(got); i++ {
				if !got[i].Equal(sdk.NewInt(tt.want[i])) {
					t.Errorf("divide got %v, want %v", got, tt.want)
				}
			}
			if got != nil {
				sum := sdk.NewInt(0)
				for _, x := range got {
					sum = sum.Add(x)
				}
				if !sum.Equal(sdk.NewInt(tt.total)) {
					t.Errorf("divide total got %v, want %v", sum, tt.total)
				}
			}
		})
	}
}

func TestDivideCoins(t *testing.T) {
	for _, tt := range []struct {
		name      string
		coins     sdk.Coins
		divisions int
		want      []sdk.Coins
	}{
		{
			name:      "one_denom",
			coins:     coins("5ubld"),
			divisions: 3,
			want:      []sdk.Coins{coins("1ubld"), coins("2ubld"), coins("2ubld")},
		},
		{
			name:      "mixed_gaps",
			coins:     coins("3xxx,2yyy"),
			divisions: 6,
			want: []sdk.Coins{
				coins(""),
				coins("1xxx"),
				coins("1yyy"),
				coins("1xxx"),
				coins(""),
				coins("1xxx,1yyy"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := divideCoins(tt.coins, tt.divisions)
			if err != nil {
				t.Fatalf("division error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("division got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonthlyVestTimes(t *testing.T) {
	for _, tt := range []struct {
		Name      string
		Start     time.Time
		Months    int
		TimeOfDay time.Time
		Want      []time.Time
		WantErr   bool
	}{
		{
			Name:      "first",
			Start:     iso("2020-01-01"),
			Months:    12,
			TimeOfDay: hhmm("12:00"),
			Want: []time.Time{
				iso("2020-02-01T12:00"),
				iso("2020-03-01T12:00"),
				iso("2020-04-01T12:00"),
				iso("2020-05-01T12:00"),
				iso("2020-06-01T12:00"),
				iso("2020-07-01T12:00"),
				iso("2020-08-01T12:00"),
				iso("2020-09-01T12:00"),
				iso("2020-10-01T12:00"),
				iso("2020-11-01T12:00"),
				iso("2020-12-01T12:00"),
				iso("2021-01-01T12:00"),
			},
		},
		{
			Name:      "clip to end of month",
			Start:     iso("2021-01-31"),
			Months:    12,
			TimeOfDay: hhmm("17:00"),
			Want: []time.Time{
				iso("2021-02-28T17:00"),
				iso("2021-03-31T17:00"),
				iso("2021-04-30T17:00"),
				iso("2021-05-31T17:00"),
				iso("2021-06-30T17:00"),
				iso("2021-07-31T17:00"),
				iso("2021-08-31T17:00"),
				iso("2021-09-30T17:00"),
				iso("2021-10-31T17:00"),
				iso("2021-11-30T17:00"),
				iso("2021-12-31T17:00"),
				iso("2022-01-31T17:00"),
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := monthlyVestTimes(tt.Start, tt.Months, tt.TimeOfDay)
			if err != nil {
				if tt.WantErr {
					return
				}
				t.Fatalf("vestingTimes failed: %v", err)
			}
			if tt.WantErr {
				t.Fatalf("vestingTimes didn't fail!")
			}
			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("vestingTimes got %v, want %v", got, tt.Want)
			}
		})
	}
}

func TestApplyCliff(t *testing.T) {
	for _, tt := range []struct {
		name   string
		cliff  time.Time
		events []event
		want   []event
	}{
		{
			name:  "before",
			cliff: iso("1999-12-31T23:59"),
			events: []event{
				evt("2020-03-01T12:00", "1000ubld"),
				evt("2020-04-01T12:00", "100ubld"),
			},
			want: []event{
				evt("2020-03-01T12:00", "1000ubld"),
				evt("2020-04-01T12:00", "100ubld"),
			},
		},
		{
			name:  "after",
			cliff: iso("2021-01-02T09:30"),
			events: []event{
				evt("2020-03-01T12:00", "1000ubld"),
				evt("2020-04-01T12:00", "100ubld"),
			},
			want: []event{
				evt("2021-01-02T09:30", "1100ubld"),
			},
		},
		{
			name:  "mid",
			cliff: iso("2021-06-15T17:00"),
			events: []event{
				evt("2021-05-15T12:00", "10ubld"),
				evt("2021-06-15T12:00", "100ubld"),
				evt("2021-07-15T12:00", "1000ubld"),
			},
			want: []event{
				evt("2021-06-15T17:00", "110ubld"),
				evt("2021-07-15T12:00", "1000ubld"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyCliff(tt.events, tt.cliff)
			if err != nil {
				t.Fatalf("applyCliff error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applyCliff got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	// Use an explicit timezone for consistant intervals between timestamps.
	oldLoc := time.Local
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("cannot load timezone: %v", err)
	}
	time.Local = loc

	for _, tt := range []struct {
		name   string
		config writeConfig
		want   cli.VestingData
	}{
		{
			name: "simple_2y",
			config: writeConfig{
				Coins:     coins("1000000000ubld"), // 1000 BLD
				Months:    24,
				TimeOfDay: hhmm("09:00"),
				Start:     iso("2021-01-01T09:30"),
				Cliffs: []time.Time{
					iso("2022-01-15T00:00"),
				},
			},
			want: cli.VestingData{
				StartTime: 1609522200,
				Periods: []cli.InputPeriod{
					{Coins: "500000000ubld", Length: 32711400},
					{Coins: "41666666ubld", Length: 1501200},
					{Coins: "41666667ubld", Length: 2419200},
					{Coins: "41666667ubld", Length: 2674800},
					{Coins: "41666666ubld", Length: 2592000},
					{Coins: "41666667ubld", Length: 2678400}, // DST begins
					{Coins: "41666667ubld", Length: 2592000},
					{Coins: "41666666ubld", Length: 2678400},
					{Coins: "41666667ubld", Length: 2678400},
					{Coins: "41666667ubld", Length: 2592000},
					{Coins: "41666666ubld", Length: 2678400},
					{Coins: "41666667ubld", Length: 2595600}, // DST ends
					{Coins: "41666667ubld", Length: 2678400},
				},
			},
		},
		{
			name: "mixed_denom",
			config: writeConfig{
				Coins:     coins("201ubld,1002urun"),
				Months:    4,
				TimeOfDay: hhmm("14:30"),
				Start:     iso("2021-07-01"),
			},
			want: cli.VestingData{
				StartTime: 1625122800,
				Periods: []cli.InputPeriod{
					{Coins: "50ubld,250urun", Length: 2730600},
					{Coins: "50ubld,251urun", Length: 2678400},
					{Coins: "50ubld,250urun", Length: 2592000},
					{Coins: "51ubld,251urun", Length: 2678400},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			events, err := tt.config.generateEvents()
			if err != nil {
				t.Fatalf("generateEvents error: %v", err)
			}
			got := tt.config.convertRelative(events)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateEvents got %v, want %v", got, tt.want)
			}
		})
	}

	time.Local = oldLoc
}

func TestFormatDuration(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  string
	}{
		{"zero", "0s", "0s"},
		{"small", "23h", "23h0m0s"},
		{"whole", "72h", "3d0h0m0s"},
		{"mixed", "127h", "5d7h0m0s"},
		{"fracsec", "76h3m7.501s", "3d4h3m7.501s"},
		{"fracsec_harder", "76h3m0.501s", "3d4h3m0.501s"},
		{"neg", "-76h3m7.501s", "-3d4h3m7.501s"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := time.ParseDuration(tt.input)
			if err != nil {
				t.Fatalf("bad duration: %v", err)
			}
			got := formatDuration(duration)
			if got != tt.want {
				t.Errorf(`got "%s", want "%s"`, got, tt.want)
			}
		})
	}
}
