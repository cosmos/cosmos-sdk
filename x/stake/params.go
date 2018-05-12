package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Params defines the high level settings for staking
type Params struct {
	InflationRateChange sdk.Rat `json:"inflation_rate_change"` // maximum annual change in inflation rate
	InflationMax        sdk.Rat `json:"inflation_max"`         // maximum inflation rate
	InflationMin        sdk.Rat `json:"inflation_min"`         // minimum inflation rate
	GoalBonded          sdk.Rat `json:"goal_bonded"`           // Goal of percent bonded atoms

	MaxValidators uint16 `json:"max_validators"` // maximum number of validators
	BondDenom     string `json:"bond_denom"`     // bondable coin denomination
}

func (p Params) equal(p2 Params) bool {
	return p.InflationRateChange.Equal(p2.InflationRateChange) &&
		p.InflationMax.Equal(p2.InflationMax) &&
		p.InflationMin.Equal(p2.InflationMin) &&
		p.GoalBonded.Equal(p2.GoalBonded) &&
		p.MaxValidators == p2.MaxValidators &&
		p.BondDenom == p2.BondDenom
}

func defaultParams() Params {
	return Params{
		InflationRateChange: sdk.NewRat(13, 100),
		InflationMax:        sdk.NewRat(20, 100),
		InflationMin:        sdk.NewRat(7, 100),
		GoalBonded:          sdk.NewRat(67, 100),
		MaxValidators:       100,
		BondDenom:           "steak",
	}
}
