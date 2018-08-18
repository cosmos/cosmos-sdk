package types

import (
	"bytes"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// defaultUnbondingTime reflects three weeks in seconds as the default
// unbonding time.
const defaultUnbondingTime time.Duration = 60 * 60 * 24 * 3 * time.Second

// Params defines the high level settings for staking
type Params struct {
	InflationRateChange sdk.Dec `json:"inflation_rate_change"` // maximum annual change in inflation rate
	InflationMax        sdk.Dec `json:"inflation_max"`         // maximum inflation rate
	InflationMin        sdk.Dec `json:"inflation_min"`         // minimum inflation rate
	GoalBonded          sdk.Dec `json:"goal_bonded"`           // Goal of percent bonded atoms

	UnbondingTime time.Duration `json:"unbonding_time"`

	MaxValidators uint16 `json:"max_validators"` // maximum number of validators
	BondDenom     string `json:"bond_denom"`     // bondable coin denomination
}

// Equal returns a boolean determining if two Param types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := MsgCdc.MustMarshalBinary(&p)
	bz2 := MsgCdc.MustMarshalBinary(&p2)
	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		InflationRateChange: sdk.NewDecWithPrec(13, 2),
		InflationMax:        sdk.NewDecWithPrec(20, 2),
		InflationMin:        sdk.NewDecWithPrec(7, 2),
		GoalBonded:          sdk.NewDecWithPrec(67, 2),
		UnbondingTime:       defaultUnbondingTime,
		MaxValidators:       100,
		BondDenom:           "steak",
	}
}
