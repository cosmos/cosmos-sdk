package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Simulation parameter constants
const (
	InflationRateChange = "inflation_rate_change"
	Inflation           = "inflation"
	InflationMax        = "inflation_max"
	InflationMin        = "inflation_min"
	GoalBonded          = "goal_bonded"
)

// GenParams generates random gov parameters
func GenParams(paramSims map[string]func(r *rand.Rand) interface{}) {
	paramSims[InflationRateChange] = func(r *rand.Rand) interface{} {
		return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
	}

	paramSims[Inflation] = func(r *rand.Rand) interface{} {
		return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
	}

	paramSims[InflationMax] = func(r *rand.Rand) interface{} {
		return sdk.NewDecWithPrec(20, 2)
	}

	paramSims[InflationMin] = func(r *rand.Rand) interface{} {
		return sdk.NewDecWithPrec(7, 2)
	}

	paramSims[GoalBonded] = func(r *rand.Rand) interface{} {
		return sdk.NewDecWithPrec(67, 2)
	}
}
