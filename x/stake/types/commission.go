package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Commission defines a commission parameters for a given validator.
type Commission struct {
	Rate           sdk.Dec   `json:"rate"`             // the commission rate charged to delegators
	MaxRate        sdk.Dec   `json:"max_rate"`         // maximum commission rate which validator can ever charge
	MaxChangeRate  sdk.Dec   `json:"max_change_rate"`  // maximum daily increase of the validator commission
	LastChangeTime time.Time `json:"last_change_time"` // the last time the commission rate was changed
}

// NewCommission returns an initialized validator commission.
func NewCommission(rate, maxRate, maxChangeRate sdk.Dec) Commission {
	return Commission{
		Rate:           rate,
		MaxRate:        maxRate,
		MaxChangeRate:  maxChangeRate,
		LastChangeTime: time.Unix(0, 0).UTC(),
	}
}

// Equal checks if the given Commission object is equal to the receiving
// Commission object.
func (c Commission) Equal(c2 Commission) bool {
	return c.Rate.Equal(c2.Rate) &&
		c.MaxRate.Equal(c2.MaxRate) &&
		c.MaxChangeRate.Equal(c2.MaxChangeRate) &&
		c.LastChangeTime.Equal(c2.LastChangeTime)
}

// String implements the Stringer interface for a Comission.
func (c Commission) String() string {
	return fmt.Sprintf("rate: %s, maxRate: %s, maxChangeRate: %s, lastChangeTime: %s",
		c.Rate, c.MaxRate, c.MaxChangeRate, c.LastChangeTime,
	)
}

// Validate performs basic sanity validation checks of initial commission
// parameters. If validation fails, an SDK error is returned.
func (c Commission) Validate() sdk.Error {
	switch {
	case c.MaxRate.LT(sdk.ZeroDec()):
		// max rate cannot be negative
		return ErrCommissionNegative(DefaultCodespace)

	case c.MaxRate.GT(sdk.OneDec()):
		// max rate cannot be greater than 100%
		return ErrCommissionHuge(DefaultCodespace)

	case c.Rate.LT(sdk.ZeroDec()):
		// rate cannot be negative
		return ErrCommissionNegative(DefaultCodespace)

	case c.Rate.GT(c.MaxRate):
		// rate cannot be greater than the max rate
		return ErrCommissionGTMaxRate(DefaultCodespace)

	case c.MaxChangeRate.LT(sdk.ZeroDec()):
		// change rate cannot be negative
		return ErrCommissionChangeRateNegative(DefaultCodespace)

	case c.MaxChangeRate.GT(c.MaxRate):
		// change rate cannot be greater than the max rate
		return ErrCommissionChangeRateGTMaxRate(DefaultCodespace)
	}

	return nil
}
