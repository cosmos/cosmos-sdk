package main

import (
	//"encoding/json"
	"reflect"
	"testing"
	"time"
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
	t, _ := time.Parse(shortIsoFmt, s)
	return t
}

func hhmmFmt(s string) time.Time {
	t, _ := time.Parse(hhmm, s)
	return t
}

func TestVestingTimes(t *testing.T) {
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
			Start:     iso("2020-01-01T16:00"),
			Months:    12,
			TimeOfDay: hhmmFmt("12:00"),
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
			got, err := vestingTimes(tt.Start, tt.Months, tt.TimeOfDay)
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

/*
func TestEncode(t *testing.T) {
	for _, tt := range []struct{
		name string
		params encodeParams
		want string
	}{
		{
			name: "simple_2y",
			params: encodeParams{
				Amount: 1000000000,  // 1000 BLD
				Denom: "ubld",
				Months: 24,
				DayOfMonth: 1,
				TimeOfDay: "09:00",
				Location: "America/Los_Angeles",
				Start: "2021-01-01T09:30",
				Cliffs: []string{
					"2022-01-01T00:00:00",
				},
			},
			want: `[
				{ "coins": "500000000ubld", "length_seconds": 31536000 },
				{ "coins": "41666666ubld", "length_seconds": 2678400 },
				{ "coins": "41666667ubld", "length_seconds": 2419200 },
				{ "coins": "41666667ubld", "length_seconds": 2678400 },
				{ "coins": "41666666ubld", "length_seconds": 2592000 },
				{ "coins": "41666667ubld", "length_seconds": 2678400 },
				{ "coins": "41666667ubld", "length_seconds": 2592000 },
				{ "coins": "41666666ubld", "length_seconds": 2678400 },
				{ "coins": "41666667ubld", "length_seconds": 2678400 },
				{ "coins": "41666667ubld", "length_seconds": 2592000 },
				{ "coins": "41666666ubld", "length_seconds": 2678400 },
				{ "coins": "41666667ubld", "length_seconds": 2592000 },
				{ "coins": "41666667ubld", "length_seconds": 2678400 }
			]`,
		},
	}{
		t.Run(tt.name, func(t *testing.T) {
			gotRaw, err := encode(tt.params)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}
			got := []filePeriod{}
			err = json.Unmarshal([]byte(gotRaw), &got)
			if err != nil {
				t.Fatalf("error decoding got JSON: %v", err)
			}
			want := []filePeriod{}
			err = json.Unmarshal([]byte(tt.want), &want)
			if err != nil {
				t.Fatalf("error decoding want JSON: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("encoding got %v, want %v", got, want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	for _, tt := range []struct{
		name string
		input string
		want string
	}{
		{"zero", "0s", "0s"},
		{"small", "23h", "23h0m0s"},
		{"whole", "72h", "3d0h0m0s"},
		{"mixed", "127h", "5d7h0m0s"},
		{"fracsec", "76h3m7.501s", "3d4h3m7.501s"},
		{"fracsec_harder", "76h3m0.501s", "3d4h3m0.501s"},
		{"neg", "-76h3m7.501s", "-3d4h3m7.501s"},
	}{
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

func TestParseDate(t *testing.T) {
	for _, tt := range []struct {
		name string
		date string
		want string
	}{
		{"ref", "2006-001-02", "Mon Jan  2 00:00:00 2006"},
	}{
		t.Run(tt.name, func(t *testing.T) {
			tm, err := parseDate(tt.date)
			if err != nil {
				t.Fatalf("parseDate error: %v", err)
			}
			got := tm.Format(time.ANSIC)
			if got != tt.want {
				t.Errorf(`parseDate got "%s", want "%s"`, got , tt.want)
			}
		})
	}
}
*/
