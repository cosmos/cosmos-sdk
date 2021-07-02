package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/x/auth/vesting/client/cli"
)

// divide returns the division of total as evenly as possible.
// Divisions must be 1 or greater and total must be nonnegative.
func divide(total int64, divisions int32) ([]int64, error) {
	if divisions < 1 {
		return nil, fmt.Errorf("divisions must be 1 or greater")
	}
	if total < 0 {
		return nil, fmt.Errorf("total must be nonnegative")
	}
	a := make([]int64, divisions)
	// Figure out the truncated division and the amount left over.
	// Fact: remainder < divisions
	truncated := total / int64(divisions)
	remainder := total - truncated*int64(divisions)
	runningTot := int64(0)
	for i := int32(0); i < divisions; i++ {
		// restrictiong divisions to int32 prevents overflow
		nextTot := remainder * int64(i+1) / int64(divisions)
		a[i] = truncated + nextTot - runningTot
		runningTot = nextTot
	}
	// Integrity check
	sum := int64(0)
	for _, x := range a {
		sum = sum + x
	}
	if sum != total {
		return nil, fmt.Errorf("failed integrity check: divisions sum to %d, should be %d", sum, total)
	}
	return a, nil
}

// monthlyVestTimes generates timestamps for successive months after startTime.
// The monthly events occur at the given time of day. If the month is not
// long enough for the desired date, the last day of the month is used.
func monthlyVestTimes(startTime time.Time, months int32, timeOfDay time.Time) ([]time.Time, error) {
	if months < 1 {
		return nil, fmt.Errorf("must have at least one vesting period")
	}
	location := startTime.Location()
	hour := timeOfDay.Hour()
	minute := timeOfDay.Minute()
	second := timeOfDay.Second()
	times := make([]time.Time, months)
	for i := 1; i <= int(months); i++ {
		tm := startTime.AddDate(0, int(i), 0)
		if tm.Day() != startTime.Day() {
			// The starting day-of-month cannot fit in this month,
			// and we've wrapped to the next month. Back up to the
			// end of the previous month.
			tm = tm.AddDate(0, 0, -tm.Day())
		}
		times[i-1] = time.Date(tm.Year(), tm.Month(), tm.Day(), hour, minute, second, 0, location)
	}
	// Integrity check: dates must be sequential and 26-33 days apart.
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

// encodeCoins encodes the given amount and denomination in coin format.
// TODO: use sdk standard coin parsing and formatting.
func encodeCoins(amount int64, denom string) string {
	return fmt.Sprint(amount) + denom
}

// parseCoins decodes the coin format into an amount and denomination
func parseCoins(coins string) (int64, string) {
	var amount int64
	var denom string
	fmt.Sscanf(coins, "%d%s", &amount, &denom)
	return amount, denom
}

// marshalPeriods gives the JSON encoding.
func marshalPeriods(periods []cli.InputPeriod) ([]byte, error) {
	return json.MarshalIndent(periods, "", "  ")
}

// unmarshalPeriods parses an array of periods in JSON.
func unmarshalPeriods(bz []byte) ([]cli.InputPeriod, error) {
	periods := []cli.InputPeriod{}
	err := json.Unmarshal(bz, &periods)
	if err != nil {
		return nil, err
	}
	return periods, nil
}

// event represents a single vesting event with an absolute time.
// The denomination must be understood by context.
// TODO: switch to sdk.Coins - doesn't need to be just one denom.
type event struct {
	Time   time.Time
	Amount int64 // TODO replace int64 with sdk.Int
}

// zipEvents generates events by zipping corresponding amounts and times.
func zipEvents(amounts []int64, times []time.Time) ([]event, error) {
	n := len(amounts)
	if len(times) != n {
		return nil, fmt.Errorf("amount and time arrays are unequal")
	}
	events := make([]event, n)
	for i := 0; i < n; i++ {
		events[i] = event{Time: times[i], Amount: amounts[i]}
	}
	return events, nil
}

// marshalEvents returns a printed representation of events.
func marshalEvents(events []event) ([]byte, error) {
	var b strings.Builder
	b.WriteString("[\n")
	for _, e := range events {
		b.WriteString("    ")
		b.WriteString(formatIso(e.Time))
		b.WriteString(": ")
		b.WriteString(fmt.Sprint(e.Amount))
		b.WriteString("\n")
	}
	b.WriteString("]")
	return []byte(b.String()), nil
}

// applyCliff accumulates vesting events before or at the cliff time
// into a single event, leaving subsequent events unchanged.
func applyCliff(events []event, cliff time.Time) ([]event, error) {
	newEvents := []event{}
	amount := int64(0)
	for _, e := range events {
		if !e.Time.After(cliff) {
			amount = amount + e.Amount
			continue
		}
		if amount != 0 {
			cliffEvent := event{Time: cliff, Amount: amount}
			newEvents = append(newEvents, cliffEvent)
			amount = 0
		}
		newEvents = append(newEvents, e)
	}
	if amount != 0 {
		// special case if all events are before the cliff
		cliffEvent := event{Time: cliff, Amount: amount}
		newEvents = append(newEvents, cliffEvent)
	}
	// integrity check
	oldTotal := int64(0)
	for _, e := range events {
		oldTotal = oldTotal + e.Amount
	}
	newTotal := int64(0)
	for _, e := range newEvents {
		newTotal = newTotal + e.Amount
	}
	if oldTotal != newTotal {
		return nil, fmt.Errorf("applying vesting cliff changed total from %d to %d", oldTotal, newTotal)
	}
	return newEvents, nil
}

// eventsToPeriods converts the events to periods with the given start time
// and denomination.
func eventsToPeriods(startTime time.Time, events []event, denom string) []cli.InputPeriod {
	periods := []cli.InputPeriod{}
	lastTime := startTime
	for _, e := range events {
		dur := e.Time.Sub(lastTime)
		p := cli.InputPeriod{
			Coins:  encodeCoins(e.Amount, denom),
			Length: int64(dur.Seconds()),
		}
		periods = append(periods, p)
		lastTime = e.Time
	}
	return periods
}

// periodsToEvents converts periods to events with the given start time.
func periodsToEvents(startTime time.Time, periods []cli.InputPeriod) []event {
	events := []event{}
	lastTime := startTime
	for _, p := range periods {
		amount, _ := parseCoins(p.Coins)
		newTime := lastTime.Add(time.Duration(p.Length) * time.Second)
		e := event{
			Time:   newTime,
			Amount: amount,
		}
		events = append(events, e)
		lastTime = newTime
	}
	return events
}

// Time utilities

const day = 24 * time.Hour

// formatDuration returns a duration in a string like "3d4h3m0.5s".
// It follows time.Duration.String() except that it includes 24-hour days.
// NOTE: Does not reflect daylight savings changes.
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
	days := int64(d / day)
	r := d % day // remainder
	s = s + fmt.Sprint(days) + "d"
	hours := int64(r / time.Hour)
	r = r % time.Hour
	s = s + fmt.Sprint(hours) + "h"
	minutes := int64(r / time.Minute)
	r = r % time.Minute
	s = s + fmt.Sprint(minutes) + "m"
	seconds := int64(r / time.Second)
	r = r % time.Second
	s = s + fmt.Sprint(seconds) // no suffix yet
	if r != 0 {
		// Follow normal Duration formatting, but need to avoid
		// the special handling of fractional seconds.
		r = r + time.Second
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

var (
	validDenoms = map[string]bool{"ubld": true} // TODO replace with cosmos-sdk denom validation
)

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
	flagAmount = flag.Int64("amount", 0, "Total amount to vest.")
	flagDenom  = flag.String("denom", "ubld", "Denomination of amount.")
	flagTime   = isoTimeFlag("time", "00:00", "Time of day for vesting, e.g. 15:04.")
	flagCliffs = isoDateListFlag("cliffs", "Vesting cliffs in format 2006-01-02T15:04 (local time).")
	flagRead   = flag.Bool("read", false, "Read periods file from stdin and print dates relative to start.")
	flagWrite  = flag.Bool("write", false, "Write periods file to stdout.")
)

// readConfig bundles data needed for the read operation.
type readConfig struct {
	startTime time.Time
}

// genReadConfig creates a readConfig from flag setings and validates it.
func genReadConfig() (readConfig, error) {
	rc := readConfig{}
	if flagStart.IsZero() {
		return rc, fmt.Errorf("must set start time with --start")
	}
	rc.startTime = *flagStart
	return rc, nil
}

// convertAbsolute converts relative periods to absolute events.
func (rc readConfig) convertAbsolute(periods []cli.InputPeriod) []event {
	return periodsToEvents(rc.startTime, periods)
}

// readCmd reads periods in JSON from stdin and writes a sequence of vesting
// events in local time to stdout.
func readCmd() {
	rc, err := genReadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad read configuration: %v", err)
		return
	}
	bzIn, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read stdin: %v", err)
		return
	}
	periods, err := unmarshalPeriods(bzIn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot decode periods: %v", err)
		return
	}
	events := rc.convertAbsolute(periods)
	bzOut, err := marshalEvents(events)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot encode events: %v", err)
		return
	}
	fmt.Println(string(bzOut))
}

// writeConfig bundles data needed for the write operation.
type writeConfig struct {
	Amount    int64
	Denom     string
	Months    int32
	TimeOfDay time.Time
	Start     time.Time
	Cliffs    []time.Time
}

// genWriteConfig generates a writeConfig from flag settings and validates it.
func genWriteConfig() (writeConfig, error) {
	wc := writeConfig{}
	if *flagAmount < 1 {
		return wc, fmt.Errorf("must have a postive amount")
	}
	wc.Amount = *flagAmount
	if _, ok := validDenoms[*flagDenom]; !ok {
		return wc, fmt.Errorf("must use a valid denomination (%v)", validDenoms)
	}
	wc.Denom = *flagDenom
	if *flagMonths < 1 || *flagMonths > math.MaxInt32 {
		return wc, fmt.Errorf("must use a positive number of months")
	}
	wc.Months = int32(*flagMonths)
	wc.TimeOfDay = *flagTime
	wc.Start = *flagStart
	wc.Cliffs = *flagCliffs
	return wc, nil
}

// generateEvents generates vesting events for the given amount and
// denomination across the given monthly vesting events with the given start
// time and subject to the vesting cliff times, if any.
func (wc writeConfig) generateEvents() ([]event, error) {
	divisions, err := divide(wc.Amount, wc.Months)
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

// convertRelative converts absolute-time events to relative periods.
func (wc writeConfig) convertRelative(events []event) []cli.InputPeriod {
	return eventsToPeriods(wc.Start, events, wc.Denom)
}

// writeCmd generates a set of vesting events based on flags and writes a
// sequences of periods in JSON format to stdout.
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
	periods := wc.convertRelative(events)
	bz, err := marshalPeriods(periods)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot marshal periods: %v", err)
		return
	}
	fmt.Println(string(bz))
}

// main executes either readCmd() or writeCmd() based on flags.
func main() {
	flag.Parse()
	if *flagRead && !*flagWrite {
		readCmd()
	} else if *flagWrite && !*flagRead {
		writeCmd()
	} else {
		fmt.Fprintln(os.Stderr, "Must specify one of --read or --write")
		flag.Usage()
	}
}
