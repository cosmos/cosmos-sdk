package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	// MinSelfBond defines a minimum self bond parameters for a given validator.
	MinSelfBond struct {
		Amount               sdk.Int   `json:"amount"`                  // minimum self bond that this validator can have
		MaxDailyDecreaseRate sdk.Int   `json:"max_daily_decrease_rate"` // maximum rate at which validator can decrease MinSelfBond
		UpdateTime           time.Time `json:"update_time"`             // the last time the MinSelfBond rate was changed
	}

	// MinSelfBondMsg defines a minselfbond message to be used for creating or editing a
	// validator.
	MinSelfBondMsg struct {
		Amount               sdk.Int `json:"amount"`            // minimum self bond that this validator can have
		MaxDailyDecreaseRate sdk.Int `json:"max_decrease_rate"` // maximum rate at which validator can decrease MinSelfBond
	}
)

// NewSelfBondMsg returns an initialized validator SelfBond message.
func NewMinSelfBondMsg(minSelfBond, maxDailyDecreaseRate sdk.Int) MinSelfBondMsg {
	return MinSelfBondMsg{
		Amount:               minSelfBond,
		MaxDailyDecreaseRate: maxDailyDecreaseRate,
	}
}

// NewSelfBond returns an initialized validator SelfBond.
func NewMinSelfBond(minSelfBond, maxDailyDecreaseRate sdk.Int) MinSelfBond {
	return MinSelfBond{
		Amount:               minSelfBond,
		MaxDailyDecreaseRate: maxDailyDecreaseRate,
		UpdateTime:           time.Unix(0, 0).UTC(),
	}
}

// NewSelfBond returns an initialized validator SelfBond with a specified
// update time which should be the current block BFT time.
func NewMinSelfBondWithTime(minSelfBond, maxDailyDecreaseRate sdk.Int, updatedAt time.Time) MinSelfBond {
	return MinSelfBond{
		Amount:               minSelfBond,
		MaxDailyDecreaseRate: maxDailyDecreaseRate,
		UpdateTime:           updatedAt,
	}
}

// Equal checks if the given SelfBond object is equal to the receiving
// SelfBond object.
func (msb MinSelfBond) Equal(msb2 MinSelfBond) bool {
	return msb.Amount.Equal(msb2.Amount) &&
		msb.MaxDailyDecreaseRate.Equal(msb2.MaxDailyDecreaseRate) &&
		msb.UpdateTime.Equal(msb2.UpdateTime)
}

// String implements the Stringer interface for a Commission.
func (msb MinSelfBond) String() string {
	return fmt.Sprintf("minSelfBond: %s, maxDailyDecreaseRate: %s, updateTime: %s",
		msb.Amount, msb.MaxDailyDecreaseRate, msb.UpdateTime,
	)
}

// Validate performs basic sanity validation checks of initial commission
// parameters. If validation fails, an SDK error is returned.
func (msb MinSelfBond) Validate() sdk.Error {
	switch {
	case !msb.Amount.GT(sdk.ZeroInt()):
		return ErrMinSelfBondInvalid(DefaultCodespace)
	case msb.MaxDailyDecreaseRate.LT(sdk.ZeroInt()):
		return ErrMinSelfBondMaxDailyDecreaseRateInvalid(DefaultCodespace)
	default:
		return nil
	}
}

// Validate performs basic sanity validation checks of initial commission
// parameters. If validation fails, an SDK error is returned.
func (msb MinSelfBond) UpdateMinSelfBond(newMSB MinSelfBondMsg, currTime time.Time) (MinSelfBond, sdk.Error) {
	err := msb.Validate()
	if err != nil {
		return msb, err
	}

	switch {
	case currTime.Sub(msb.UpdateTime).Hours() < 24:
		return msb, ErrMinSelfBondUpdateTime(DefaultCodespace)
	case !newMSB.Amount.LT(msb.Amount.Sub(msb.MaxDailyDecreaseRate)):
		return msb, ErrMinSelfBondTooLow(DefaultCodespace)
	case newMSB.MaxDailyDecreaseRate.GT(msb.MaxDailyDecreaseRate):
		return msb, ErrMinSelfBondMaxDailyDecreaseRateIncreased(DefaultCodespace)
	default:
		msb.Amount = newMSB.Amount
		msb.MaxDailyDecreaseRate = newMSB.MaxDailyDecreaseRate
		msb.UpdateTime = currTime
		return msb, nil
	}
}
