package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Period defines a length of time and amount of coins that will vest
type Period struct {
	Length int64     `json:"length" yaml:"length"` // length of the period, in seconds
	Amount sdk.Coins `json:"amount" yaml:"amount"` // amount of coins vesting during this period
}

// Periods stores all vesting periods passed as part of a PeriodicVestingAccount
type Periods []Period

// String Period implements stringer interface
func (p Period) String() string {
	return fmt.Sprintf(`Length: %d
	Amount: %s`, p.Length, p.Amount)
}

// String Periods implements stringer interface
func (vp Periods) String() string {
	periodsListString := make([]string, len(vp))
	for _, period := range vp {
		periodsListString = append(periodsListString, period.String())
	}
	return strings.TrimSpace(fmt.Sprintf(`Vesting Periods:
		%s`, strings.Join(periodsListString, ", ")))
}
