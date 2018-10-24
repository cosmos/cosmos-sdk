package mint

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// mint parameters
type Params struct {
	MintDenom           string  `json:"mint_denom"`            // type of coin to mint
	InflationRateChange sdk.Dec `json:"inflation_rate_change"` // maximum annual change in inflation rate
	InflationMax        sdk.Dec `json:"inflation_max"`         // maximum inflation rate
	InflationMin        sdk.Dec `json:"inflation_min"`         // minimum inflation rate
	GoalBonded          sdk.Dec `json:"goal_bonded"`           // goal of percent bonded atoms
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:           "steak",
		InflationRateChange: sdk.NewDecWithPrec(13, 2),
		InflationMax:        sdk.NewDecWithPrec(20, 2),
		InflationMin:        sdk.NewDecWithPrec(7, 2),
		GoalBonded:          sdk.NewDecWithPrec(67, 2),
	}
}

func validateParams(params Params) error {
	if params.GoalBonded.LT(sdk.ZeroDec()) {
		return fmt.Errorf("mint parameter GoalBonded should be positive, is %s ", params.GoalBonded.String())
	}
	if params.GoalBonded.GT(sdk.OneDec()) {
		return fmt.Errorf("mint parameter GoalBonded must be <= 1, is %s", params.GoalBonded.String())
	}
	if params.InflationMax.LT(params.InflationMin) {
		return fmt.Errorf("mint parameter Max inflation must be greater than or equal to min inflation")
	}
	if params.MintDenom == "" {
		return fmt.Errorf("mint parameter MintDenom can't be an empty string")
	}
	return nil
}
