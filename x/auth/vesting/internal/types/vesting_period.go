package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VestingPeriod defines a length of time and amount of coins that will vest
type VestingPeriod struct {
	PeriodLength  int64     `json:"period_length" yaml:"period_length"`   // length of the period, in seconds
	VestingAmount sdk.Coins `json:"vesting_amount" yaml:"vesting_amount"` // amount of coins vesting during this period
}

// VestingPeriods stores all vesting periods passed as part of a PeriodicVestingAccount
type VestingPeriods []VestingPeriod

// String VestingPeriod implements stringer interface
func (p VestingPeriod) String() string {
	return fmt.Sprintf(`Period Length: %d
	VestingAmount: %s`, p.PeriodLength, p.VestingAmount)
}

// String VestingPeriods implements stringer interface
func (vp VestingPeriods) String() string {
	var periodsListString []string
	for _, period := range vp {
		periodsListString = append(periodsListString, period.String())
	}
	return strings.TrimSpace(fmt.Sprintf(`Vesting Periods:
		%s`, strings.Join(periodsListString, ", ")))
}
