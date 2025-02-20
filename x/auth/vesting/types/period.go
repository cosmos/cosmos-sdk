package types

import (
	"fmt"
	"strings"
	"time"

	"sigs.k8s.io/yaml"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Periods stores all vesting periods passed as part of a PeriodicVestingAccount
type Periods []Period

// Duration is converts the period Length from seconds to a time.Duration
func (p Period) Duration() time.Duration {
	return time.Duration(p.Length) * time.Second
}

// String implements the fmt.Stringer interface
func (p Period) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// TotalLength return the total length in seconds for a period
func (p Periods) TotalLength() int64 {
	var total int64
	for _, period := range p {
		total += period.Length
	}
	return total
}

// TotalDuration returns the total duration of the period
func (p Periods) TotalDuration() time.Duration {
	len := p.TotalLength()
	return time.Duration(len) * time.Second
}

// TotalDuration returns the sum of coins for the period
func (p Periods) TotalAmount() sdk.Coins {
	total := sdk.Coins{}
	for _, period := range p {
		total = total.Add(period.Amount...)
	}
	return total
}

// String implements the fmt.Stringer interface
func (p Periods) String() string {
	periodsListString := make([]string, len(p))
	for _, period := range p {
		periodsListString = append(periodsListString, period.String())
	}

	return strings.TrimSpace(fmt.Sprintf(`Vesting Periods:
		%s`, strings.Join(periodsListString, ", ")))
}

// A "schedule" is an increasing step function of Coins over time.
// It's specified as an absolute start time and a sequence of relative
// periods, with each step at the end of a period. A schedule may also
// give the time and total value at the last step, which can speed
// evaluation of the step function after the last step.

// ReadSchedule returns the value of a schedule at readTime.
func ReadSchedule(startTime, endTime int64, periods []Period, totalCoins sdk.Coins, readTime int64) sdk.Coins {
	if readTime <= startTime {
		return sdk.NewCoins()
	}
	if readTime >= endTime {
		return totalCoins
	}

	coins := sdk.NewCoins() // sum of amounts for events before readTime
	time := startTime

	for _, period := range periods {
		if readTime < time+period.Length {
			// we're reading before the next event
			break
		}
		coins = coins.Add(period.Amount...)
		time += period.Length
	}

	return coins
}

// max64 returns the maximum of its inputs.
func max64(i, j int64) int64 {
	if i > j {
		return i
	}
	return j
}

// min64 returns the minimum of its inputs.
func min64(i, j int64) int64 {
	if i < j {
		return i
	}
	return j
}

// coinsMin returns the minimum of its inputs for all denominations.
func coinsMin(a, b sdk.Coins) sdk.Coins {
	min := sdk.NewCoins()
	for _, coinA := range a {
		denom := coinA.Denom
		bAmt := b.AmountOfNoDenomValidation(denom)
		minAmt := coinA.Amount
		if minAmt.GT(bAmt) {
			minAmt = bAmt
		}
		if minAmt.IsPositive() {
			min = min.Add(sdk.NewCoin(denom, minAmt))
		}
	}
	return min
}

// DisjunctPeriods returns the union of two vesting period schedules.
// The returned schedule is the union of the vesting events, with simultaneous
// events combined into a single event.
// Input schedules P and Q are defined by their start times and periods.
// Returns new start time, new end time, and merged vesting events, relative to
// the new start time.
func DisjunctPeriods(startP, startQ int64, periodsP, periodsQ []Period) (int64, int64, []Period) {
	timeP := startP // time of last merged p event, next p event is relative to this time
	timeQ := startQ // time of last merged q event, next q event is relative to this time
	iP := 0         // p indexes before this have been merged
	iQ := 0         // q indexes before this have been merged
	lenP := len(periodsP)
	lenQ := len(periodsQ)
	startTime := min64(startP, startQ) // we pick the earlier time
	time := startTime                  // time of last merged event, or the start time
	merged := []Period{}

	// emit adds an output period and updates the last event time
	emit := func(nextTime int64, amount sdk.Coins) {
		period := Period{
			Length: nextTime - time,
			Amount: amount,
		}
		merged = append(merged, period)
		time = nextTime
	}

	// consumeP emits the next period from p, updating indexes
	consumeP := func(nextP int64) {
		emit(nextP, periodsP[iP].Amount)
		timeP = nextP
		iP++
	}

	// consumeQ emits the next period from q, updating indexes
	consumeQ := func(nextQ int64) {
		emit(nextQ, periodsQ[iQ].Amount)
		timeQ = nextQ
		iQ++
	}

	// consumeBoth emits a merge of the next periods from p and q, updating indexes
	consumeBoth := func(nextTime int64) {
		emit(nextTime, periodsP[iP].Amount.Add(periodsQ[iQ].Amount...))
		timeP = nextTime
		timeQ = nextTime
		iP++
		iQ++
	}

	// while there are more events in both schedules, handle the next one, merge if concurrent
	for iP < lenP && iQ < lenQ {
		nextP := timeP + periodsP[iP].Length // next p event in absolute time
		nextQ := timeQ + periodsQ[iQ].Length // next q event in absolute time
		switch {
		case nextP < nextQ:
			consumeP(nextP)
		case nextP > nextQ:
			consumeQ(nextQ)
		default: // nextP == nextQ
			consumeBoth(nextP)
		}
	}
	// consume remaining events in schedule P
	for iP < lenP {
		nextP := timeP + periodsP[iP].Length
		consumeP(nextP)
	}
	// consume remaining events in schedule Q
	for iQ < lenQ {
		nextQ := timeQ + periodsQ[iQ].Length
		consumeQ(nextQ)
	}
	return startTime, time, merged
}

// ConjunctPeriods returns the combination of two period schedules where the result is the minimum of the two schedules.
func ConjunctPeriods(startP, startQ int64, periodsP, periodsQ []Period) (int64, int64, []Period) {
	startTime := min64(startP, startQ)
	time := startTime

	timeP := startP
	timeQ := startQ
	iP := 0
	iQ := 0
	lenP := len(periodsP)
	lenQ := len(periodsQ)
	merged := []Period{}
	amount := sdk.NewCoins()
	amountP := amount
	amountQ := amount

	// emit adds an output period and updates the last event time
	emit := func(nextTime int64, coins sdk.Coins) {
		period := Period{
			Length: nextTime - time,
			Amount: coins,
		}
		merged = append(merged, period)
		time = nextTime
		amount = amount.Add(coins...)
	}

	// consumeP processes the next event in P and emits an event
	// if the minimum of P and Q changes
	consumeP := func(nextTime int64) {
		amountP = amountP.Add(periodsP[iP].Amount...)
		min := coinsMin(amountP, amountQ)
		if amount.IsAllLTE(min) {
			diff := min.Sub(amount...)
			if !diff.IsZero() {
				emit(nextTime, diff)
			}
		}
		timeP = nextTime
		iP++
	}

	// consumeQ processes the next event in Q and emits an event
	// if the minimum of P and Q changes
	consumeQ := func(nextTime int64) {
		amountQ = amountQ.Add(periodsQ[iQ].Amount...)
		min := coinsMin(amountP, amountQ)
		if amount.IsAllLTE(min) {
			diff := min.Sub(amount...)
			if !diff.IsZero() {
				emit(nextTime, diff)
			}
		}
		timeQ = nextTime
		iQ++
	}

	// consumeBoth processes simultaneous events in P and Q and emits an
	// event if the minimum of P and Q changes
	consumeBoth := func(nextTime int64) {
		amountP = amountP.Add(periodsP[iP].Amount...)
		amountQ = amountQ.Add(periodsQ[iQ].Amount...)
		min := coinsMin(amountP, amountQ)
		if amount.IsAllLTE(min) {
			diff := min.Sub(amount...)
			if !diff.IsZero() {
				emit(nextTime, diff)
			}
		}
		timeP = nextTime
		timeQ = nextTime
		iP++
		iQ++
	}

	// while there are events left in both schedules, process the next one
	for iP < lenP && iQ < lenQ {
		nextP := timeP + periodsP[iP].Length // next p event in absolute time
		nextQ := timeQ + periodsQ[iQ].Length // next q event in absolute time
		switch {
		case nextP < nextQ:
			consumeP(nextP)
		case nextP > nextQ:
			consumeQ(nextQ)
		default: // nextP == nextQ
			consumeBoth(nextP)
		}
	}

	// consume remaining events in schedule P
	for iP < lenP {
		nextP := timeP + periodsP[iP].Length
		consumeP(nextP)
	}

	// consume remaining events in schedule Q
	for iQ < lenQ {
		nextQ := timeQ + periodsQ[iQ].Length
		consumeQ(nextQ)
	}

	return startTime, time, merged
}

// AlignSchedules rewrites the first period length to align the two arguments to the same start time.
func AlignSchedules(startP, startQ int64, p, q []Period) (startTime, endTime int64) {
	startTime = min64(startP, startQ)

	if len(p) > 0 {
		p[0].Length += startP - startTime
	}
	if len(q) > 0 {
		q[0].Length += startQ - startTime
	}

	endP := startTime
	for _, period := range p {
		endP += period.Length
	}
	endQ := startTime
	for _, period := range q {
		endQ += period.Length
	}
	endTime = max64(endP, endQ)
	return
}
