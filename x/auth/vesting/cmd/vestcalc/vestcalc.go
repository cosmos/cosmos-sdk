package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/client/cli"
)

// vestcalc is a utility for creating or reading schedule files
// for use in some vesting account types.  See README.md for usage.

// divide returns the division of total as evenly as possible.
// Divisor must be 1 or greater and total must be nonnegative.
func divide(total sdk.Int, divisor int) ([]sdk.Int, error) {
	if divisor < 1 {
		return nil, fmt.Errorf("divisions must be 1 or greater")
	}
	div64 := int64(divisor)
	if total.IsNegative() {
		return nil, fmt.Errorf("total must be nonnegative")
	}
	divisions := make([]sdk.Int, divisor)

	// Ideally we could compute total of the first i divisions as
	//     cumulative(i) = floor((total * i) / divisor)
	// and so
	//     divisions[i] = cumulative(i + 1) - cumulative(i)
	// but this could lead to numeric overflow for large values of total.
	// Instead, we'll compute
	//     truncated = floor(total / divisor)
	// so that
	//     total = truncated * divisor + remainder
	// where remainder < divisor, then divide the remainder via the
	// above algorithm - which now won't overflow - and sum the
	// truncated and slices of the remainder to form the divisions.
	truncated := total.QuoRaw(div64)
	remainder := total.ModRaw(div64)
	cumulative := sdk.NewInt(0) // portion of remainder which has been doled out
	for i := int64(0); i < div64; i++ {
		// multiply will not overflow since remainder and div64 are < 2^63
		nextCumulative := remainder.MulRaw(i + 1).QuoRaw(div64)
		divisions[i] = truncated.Add(nextCumulative.Sub(cumulative))
		cumulative = nextCumulative
	}

	// Integrity check
	sum := sdk.NewInt(0)
	for _, x := range divisions {
		sum = sum.Add(x)
	}
	if !sum.Equal(total) {
		return nil, fmt.Errorf("failed integrity check: divisions of %v sum to %d, should be %d", divisions, sum, total)
	}
	return divisions, nil
}

// divideCoins divides the coins into divisor separate parts as evenly as possible.
// Divisor must be positive. Returns an array holding the division.
func divideCoins(coins sdk.Coins, divisor int) ([]sdk.Coins, error) {
	if divisor < 1 {
		return nil, fmt.Errorf("divisor must be 1 or greater")
	}
	divisions := make([]sdk.Coins, divisor)
	divisionsByDenom := make(map[string][]sdk.Int)
	for _, coin := range coins {
		dividedCoin, err := divide(coin.Amount, divisor)
		if err != nil {
			return nil, fmt.Errorf("cannot divide %s: %v", coin.Denom, err)
		}
		divisionsByDenom[coin.Denom] = dividedCoin
	}
	for i := 0; i < divisor; i++ {
		newCoins := sdk.NewCoins()
		for _, coin := range coins {
			c := sdk.NewCoin(coin.Denom, divisionsByDenom[coin.Denom][i])
			newCoins = newCoins.Add(c)
		}
		divisions[i] = newCoins
	}
	// Integrity check
	sum := sdk.NewCoins()
	for _, c := range divisions {
		sum = sum.Add(c...)
	}
	if !sum.IsEqual(coins) {
		return nil, fmt.Errorf("failed integrity check: divisions of %v sum to %s, should be %s", divisions, sum, coins)
	}
	return divisions, nil
}

// monthlyVestTimes generates timestamps for successive months after startTime.
// The monthly events occur at the given time of day. If the month is not
// long enough for the desired date, the last day of the month is used.
// The number of months must be positive.
func monthlyVestTimes(startTime time.Time, months int, timeOfDay time.Time) ([]time.Time, error) {
	if months < 1 {
		return nil, fmt.Errorf("must have at least one vesting period")
	}
	location := startTime.Location()
	hour := timeOfDay.Hour()
	minute := timeOfDay.Minute()
	second := timeOfDay.Second()
	times := make([]time.Time, months)
	for i := 1; i <= months; i++ {
		tm := startTime.AddDate(0, i, 0)
		if tm.Day() != startTime.Day() {
			// The starting day-of-month cannot fit in this month,
			// and we've wrapped to the next month. Back up to the
			// end of the previous month.
			tm = tm.AddDate(0, 0, -tm.Day())
		}
		times[i-1] = time.Date(tm.Year(), tm.Month(), tm.Day(), hour, minute, second, 0, location)
	}
	// Integrity check: dates must be sequential and 26-33 days apart.
	// (Jan 31 to Feb 28 or Feb 28 to Mar 31, plus slop for DST.)
	lastTime := startTime
	for _, tm := range times {
		duration := tm.Sub(lastTime)
		if duration < 26*24*time.Hour {
			return nil, fmt.Errorf("vesting dates too close: %v and %v", lastTime, tm)
		}
		if duration > 33*24*time.Hour {
			return nil, fmt.Errorf("vesting dates too distant: %v and %v", lastTime, tm)
		}
		lastTime = tm
	}
	return times, nil
}

// marshalVestingData writes the vesting data as JSON.
func marshalVestingData(data cli.VestingData) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// unmarshalVestingData parses the vesting data from JSON.
func unmarshalVestingData(bz []byte) (cli.VestingData, error) {
	data := cli.VestingData{}
	err := json.Unmarshal(bz, &data)
	return data, err
}

// event represents a single vesting event with an absolute time.
type event struct {
	Time  time.Time
	Coins sdk.Coins
}

// zipEvents generates events by zipping corresponding amounts and times
// from equal-sized arrays, returning an event array of the same size.
func zipEvents(divisions []sdk.Coins, times []time.Time) ([]event, error) {
	n := len(divisions)
	if len(times) != n {
		return nil, fmt.Errorf("amount and time arrays are unequal")
	}
	events := make([]event, n)
	for i := 0; i < n; i++ {
		events[i] = event{Time: times[i], Coins: divisions[i]}
	}
	return events, nil
}

// marshalEvents returns a printed representation of events.
// nolint:unparam
func marshalEvents(events []event) ([]byte, error) {
	var b strings.Builder
	b.WriteString("[\n")
	for _, e := range events {
		b.WriteString("    ")
		b.WriteString(formatIso(e.Time))
		b.WriteString(": ")
		b.WriteString(e.Coins.String())
		b.WriteString("\n")
	}
	b.WriteString("]")
	return []byte(b.String()), nil
}

// applyCliff accumulates vesting events before or at the cliff time
// into a single event, leaving subsequent events unchanged.
func applyCliff(events []event, cliff time.Time) ([]event, error) {
	newEvents := []event{}
	preCliffAmount := sdk.NewCoins()
	i := 0
	for ; i < len(events) && !events[i].Time.After(cliff); i++ {
		preCliffAmount = preCliffAmount.Add(events[i].Coins...)
	}
	if !preCliffAmount.IsZero() {
		cliffEvent := event{Time: cliff, Coins: preCliffAmount}
		newEvents = append(newEvents, cliffEvent)
	}
	for ; i < len(events); i++ {
		newEvents = append(newEvents, events[i])
	}

	// integrity check
	oldTotal := sdk.NewCoins()
	for _, e := range events {
		oldTotal = oldTotal.Add(e.Coins...)
	}
	newTotal := sdk.NewCoins()
	for _, e := range newEvents {
		newTotal = newTotal.Add(e.Coins...)
	}
	if !oldTotal.IsEqual(newTotal) {
		return nil, fmt.Errorf("applying vesting cliff changed total from %s to %s", oldTotal, newTotal)
	}

	return newEvents, nil
}

// eventsToVestingData converts the events to VestingData with the given start time.
func eventsToVestingData(startTime time.Time, events []event) cli.VestingData {
	periods := []cli.InputPeriod{}
	lastTime := startTime
	for _, e := range events {
		dur := e.Time.Sub(lastTime)
		p := cli.InputPeriod{
			Coins:  e.Coins.String(),
			Length: int64(dur.Seconds()),
		}
		periods = append(periods, p)
		lastTime = e.Time
	}
	return cli.VestingData{
		StartTime: startTime.Unix(),
		Periods:   periods,
	}
}

// vestingDataToEvents converts the vesting data to absolute-timestamped events.
func vestingDataToEvents(data cli.VestingData) ([]event, error) {
	startTime := time.Unix(data.StartTime, 0)
	events := []event{}
	lastTime := startTime
	for _, p := range data.Periods {
		coins, err := sdk.ParseCoinsNormalized(p.Coins)
		if err != nil {
			return nil, err
		}
		newTime := lastTime.Add(time.Duration(p.Length) * time.Second)
		e := event{
			Time:  newTime,
			Coins: coins,
		}
		events = append(events, e)
		lastTime = newTime
	}
	return events, nil
}

// Time utilities

// nolint:unused
const day = 24 * time.Hour

// formatDuration returns a duration in a string like "3d4h3m0.5s".
// It follows time.Duration.String() except that it includes 24-hour days.
// NOTE: Does not reflect daylight savings changes.
// nolint:deadcode,unused
func formatDuration(d time.Duration) string {
	s := ""
	if d < 0 {
		d = -d
		s = "-"
	}
	if d < day {
		// handle several special cases
		return s + d.String()
	}
	// Now we know days are the most significant unit,
	// so all other units should be present
	r := d
	days := int64(r / day)
	r %= day // remainder
	s += fmt.Sprint(days) + "d"
	hours := int64(r / time.Hour)
	r %= time.Hour
	s += fmt.Sprint(hours) + "h"
	minutes := int64(r / time.Minute)
	r %= time.Minute
	s += fmt.Sprint(minutes) + "m"
	seconds := int64(r / time.Second)
	r %= time.Second
	s += fmt.Sprint(seconds) // no suffix yet
	if r != 0 {
		// Follow normal Duration formatting, but need to avoid
		// the special handling of fractional seconds.
		r += time.Second
		frac := r.String()
		// append skipping the leading "1"
		return s + frac[1:] // adds the suffix
	}
	return s + "s" // now the suffix
}

// maxTime gives the maximum of a set of times, or the zero time if empty.
func maxTime(cliffs []time.Time) time.Time {
	tm := time.Time{}
	for _, c := range cliffs {
		if c.After(tm) {
			tm = c
		}
	}
	return tm
}

// shortIsoFmt specifies ISO 8601 without seconds or timezone.
// Note: when parsing, timezone is UTC unless overridden.
const shortIsoFmt = "2006-01-02T15:04"

// Common ISO-8601 formats for local day/time.
var localIsoFormats = []string{
	"2006-01-02",
	"2006-01-02T15:04",
	"2006-01-02T15:04:05",
}

// parseIso tries to parse the given string as some common prefix of ISO-8601.
// "Common" means the least significant unit is day, minute, or second.
// The time will be in local time unless a timezone specifier is given.
func parseIso(s string) (time.Time, error) {
	// Try local (no explicit timezone) formats first.
	for _, fmt := range localIsoFormats {
		tm, err := time.ParseInLocation(fmt, s, time.Local)
		if err == nil {
			return tm, nil
		}
	}
	// Now try the full format.
	return time.Parse(time.RFC3339, s)
}

// formatIso formats the time in shortIso format in local time.
func formatIso(tm time.Time) string {
	return tm.Format(shortIsoFmt)
}

// hhmmFmt specifies an HH:MM time format to generate time.Time values
// where only hours, minutes, seconds are used.
const hhmmFmt = "15:04"

// Custom flag types

// isoDate is time.Time as a flag.Value in shortIsoFmt.
type isoDate struct{ time.Time }

var _ flag.Value = &isoDate{}

// Set implements flag.Value.Set().
func (id *isoDate) Set(s string) error {
	t, err := parseIso(s)
	if err != nil {
		return err
	}
	id.Time = t
	return nil
}

// String implements flag.Value.String().
func (id *isoDate) String() string {
	return formatIso(id.Time)
}

// isoDateFlag makes a new isoDate flag, accessed as a time.Time.
func isoDateFlag(name string, usage string) *time.Time {
	id := isoDate{time.Time{}}
	flag.CommandLine.Var(&id, name, usage)
	return &id.Time
}

// isoDateList is []time.Time as a flag.Value in repeated or comma-separated shortIsoFmt.
type isoDateList []time.Time

var _ flag.Value = &isoDateList{}

// Set implements flag.Value.Set().
func (dates *isoDateList) Set(s string) error {
	for _, ds := range strings.Split(s, ",") {
		d, err := parseIso(ds)
		if err != nil {
			return err
		}
		*dates = append(*dates, d) // accumulates repeated flag arguments
	}
	return nil
}

// String implements flag.Value.String().
func (dates *isoDateList) String() string {
	s := ""
	for _, t := range *dates {
		if s == "" {
			s = formatIso(t)
		} else {
			s = s + "," + formatIso(t)
		}
	}
	return s
}

// isoDateListFlag makes a new isoDateList flag.
func isoDateListFlag(name string, usage string) *isoDateList {
	dates := isoDateList([]time.Time{})
	flag.Var(&dates, name, usage)
	return &dates
}

// isoTime is time.Time as a flagValue in HH:MM format.
type isoTime struct{ time.Time }

var _ flag.Value = &isoTime{}

// Set implements flag.Value.Set().
func (it *isoTime) Set(s string) error {
	t, err := time.Parse(hhmmFmt, s)
	if err != nil {
		return err
	}
	it.Time = t
	return nil
}

// String implements flag.Value.String().
func (it *isoTime) String() string {
	return it.Format(hhmmFmt)
}

// isoTimeFlag makes a new isoTime flag, accessed as a time.Time.
func isoTimeFlag(name string, value string, usage string) *time.Time {
	t, err := time.Parse(hhmmFmt, value)
	if err != nil {
		t = time.Time{}
	}
	it := isoTime{t}
	flag.CommandLine.Var(&it, name, usage)
	return &it.Time
}

var (
	flagStart  = isoDateFlag("start", "Start date for the vesting in format 2006-01-02T15:04 (local time).")
	flagMonths = flag.Int("months", 1, "Number of months to vest over.")
	flagCoins  = flag.String("coins", "", "Total coins to vest.")
	flagTime   = isoTimeFlag("time", "00:00", "Time of day for vesting, e.g. 15:04.")
	flagCliffs = isoDateListFlag("cliffs", "Vesting cliffs in format 2006-01-02T15:04 (local time).")
	flagRead   = flag.Bool("read", false, "Read periods file from stdin and print dates relative to start.")
	flagWrite  = flag.Bool("write", false, "Write periods file to stdout.")
)

// readCmd reads a schedule file from stdin and writes a sequence of vesting
// events in local time to stdout. See README.md for the format.
func readCmd() {
	bzIn, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read stdin: %v", err)
		return
	}
	vestingData, err := unmarshalVestingData(bzIn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot decode vesting data: %v", err)
		return
	}
	events, err := vestingDataToEvents(vestingData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot convert vesting data: %v", err)
	}
	bzOut, err := marshalEvents(events)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot encode events: %v", err)
		return
	}
	fmt.Println(string(bzOut))
}

// writeConfig bundles data needed for the write operation.
type writeConfig struct {
	// Coins is the total amount to be vested.
	Coins sdk.Coins
	// Months is the number of months to vest over. Must be positive.
	Months    int
	TimeOfDay time.Time
	Start     time.Time
	Cliffs    []time.Time
}

// genWriteConfig generates a writeConfig from flag settings and validates it.
func genWriteConfig() (writeConfig, error) {
	wc := writeConfig{}
	coins, err := sdk.ParseCoinsNormalized(*flagCoins)
	if err != nil {
		return wc, fmt.Errorf("cannot parse --coins: %v", err)
	}
	wc.Coins = coins
	if *flagMonths < 1 {
		return wc, fmt.Errorf("must use a positive number of months")
	}
	wc.Months = *flagMonths
	wc.TimeOfDay = *flagTime
	wc.Start = *flagStart
	wc.Cliffs = *flagCliffs
	return wc, nil
}

// generateEvents generates vesting events from the writeConfig.
func (wc writeConfig) generateEvents() ([]event, error) {
	divisions, err := divideCoins(wc.Coins, wc.Months)
	if err != nil {
		return nil, fmt.Errorf("vesting amount division failed: %v", err)
	}
	times, err := monthlyVestTimes(wc.Start, wc.Months, wc.TimeOfDay)
	if err != nil {
		return nil, fmt.Errorf("vesting time calcuation failed: %v", err)
	}
	events, err := zipEvents(divisions, times)
	if err != nil {
		return nil, fmt.Errorf("vesting event generation failed: %v", err)
	}
	if len(wc.Cliffs) > 0 {
		last := maxTime(wc.Cliffs)
		events, err = applyCliff(events, last)
		if err != nil {
			return nil, fmt.Errorf("vesting cliff failed: %v", err)
		}
	}
	return events, nil
}

// convertRelative converts absolute-time events to VestingData relative to the Start time.
func (wc writeConfig) convertRelative(events []event) cli.VestingData {
	return eventsToVestingData(wc.Start, events)
}

// writeCmd generates a set of vesting events based on parsed flags
// and writes a schedule file to stdout.
func writeCmd() {
	wc, err := genWriteConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad write configuration: %v", err)
		return
	}
	events, err := wc.generateEvents()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot generate events: %v", err)
		return
	}
	vestingData := wc.convertRelative(events)
	bz, err := marshalVestingData(vestingData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot marshal vesting data: %v", err)
		return
	}
	fmt.Println(string(bz))
}

// main parses the flags and executes a subcommand based on flags.
// See README.md for flags and subcommands.
func main() {
	flag.Parse()
	switch {
	case *flagRead && !*flagWrite:
		readCmd()
	case *flagWrite && !*flagRead:
		writeCmd()
	default:
		fmt.Fprintln(os.Stderr, "Must specify one of --read or --write")
		flag.Usage()
		os.Exit(1)
	}
}
