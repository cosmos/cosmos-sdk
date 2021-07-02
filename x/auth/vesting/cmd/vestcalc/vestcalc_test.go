package main

import (
	//"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/auth/vesting/client/cli"
)

const (
	billion = int64(1000 * 1000 * 1000)
)

func TestDivision(t *testing.T) {
	for _, tt := range []struct {
		name      string
		total     int64
		divisions int32
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
			got, _ := divide(tt.total, tt.divisions)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("divide got %v, want %v", got, tt.want)
			}
			if got != nil {
				sum := int64(0)
				for _, x := range got {
					sum = sum + x
				}
				if sum != tt.total {
					t.Errorf("divide total got %v, want %v", sum, tt.total)
				}
			}
		})
	}
}

func iso(s string) time.Time {
	t, _ := parseIso(s)
	return t
}

func hhmm(s string) time.Time {
	t, _ := time.Parse(hhmmFmt, s)
	return t
}

func TestMonthlyVestTimes(t *testing.T) {
	for _, tt := range []struct {
		Name      string
		Start     time.Time
		Months    int32
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
				{iso("2020-03-01T12:00"), 1000},
				{iso("2020-04-01T12:00"), 100},
			},
			want: []event{
				{iso("2020-03-01T12:00"), 1000},
				{iso("2020-04-01T12:00"), 100},
			},
		},
		{
			name:  "after",
			cliff: iso("2021-01-02T09:30"),
			events: []event{
				{iso("2020-03-01T12:00"), 1000},
				{iso("2020-04-01T12:00"), 100},
			},
			want: []event{
				{iso("2021-01-02T09:30"), 1100},
			},
		},
		{
			name:  "mid",
			cliff: iso("2021-06-15T17:00"),
			events: []event{
				{iso("2021-05-15T12:00"), 10},
				{iso("2021-06-15T12:00"), 100},
				{iso("2021-07-15T12:00"), 1000},
			},
			want: []event{
				{iso("2021-06-15T17:00"), 110},
				{iso("2021-07-15T12:00"), 1000},
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
				Amount:    1000000000, // 1000 BLD
				Denom:     "ubld",
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
	} {
		t.Run(tt.name, func(t *testing.T) {
			events, err := tt.config.generateEvents()
			if err != nil {
				t.Fatalf("generateEvents error: %v", err)
			}
			got := tt.config.convertRelative(events)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encoding got %v, want %v", got, tt.want)
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
