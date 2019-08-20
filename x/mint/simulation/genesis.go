package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

// Simulation parameter constants
const (
	Inflation           = "inflation"
	InflationRateChange = "inflation_rate_change"
	InflationMax        = "inflation_max"
	InflationMin        = "inflation_min"
	GoalBonded          = "goal_bonded"
)

// GenInflation randomized Inflation
func GenInflation(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenInflationRateChange randomized InflationRateChange
func GenInflationRateChange(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenInflationMax randomized InflationMax
func GenInflationMax(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(20, 2)
}

// GenInflationMin randomized InflationMin
func GenInflationMin(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(7, 2)
}

// GenGoalBonded randomized GoalBonded
func GenGoalBonded(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(67, 2)
}

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(input *module.GeneratorInput) {
	var (
		inflation           sdk.Dec
		inflationRateChange sdk.Dec
		inflationMax        sdk.Dec
		inflationMin        sdk.Dec
		goalBonded          sdk.Dec
	)

	// minter
	input.AppParams.GetOrGenerate(input.Cdc, Inflation, &inflation, input.R,
		func(r *rand.Rand) { inflation = GenInflation(input.R) })

	minter := types.InitialMinter(inflation)

	// params
	input.AppParams.GetOrGenerate(input.Cdc, InflationRateChange, &inflationRateChange, input.R,
		func(r *rand.Rand) { inflationRateChange = GenInflationRateChange(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, InflationMax, &inflationMax, input.R,
		func(r *rand.Rand) { inflationMax = GenInflationMax(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, InflationMin, &inflationMin, input.R,
		func(r *rand.Rand) { inflationMin = GenInflationMin(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, GoalBonded, &goalBonded, input.R,
		func(r *rand.Rand) { goalBonded = GenGoalBonded(input.R) })

	mintDenom := sdk.DefaultBondDenom
	blocksPerYear := uint64(60 * 60 * 8766 / 5)
	params := types.NewParams(mintDenom, inflationRateChange, inflationMax, inflationMin, goalBonded, blocksPerYear)

	mintGenesis := types.NewGenesisState(minter, params)

	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", codec.MustMarshalJSONIndent(input.Cdc, mintGenesis))
	input.GenState[types.ModuleName] = input.Cdc.MustMarshalJSON(mintGenesis)
}
