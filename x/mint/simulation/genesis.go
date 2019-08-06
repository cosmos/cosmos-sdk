package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation parameter constants
const (
	InflationRateChange = "inflation_rate_change"
	Inflation           = "inflation"
	InflationMax        = "inflation_max"
	InflationMin        = "inflation_min"
	GoalBonded          = "goal_bonded"
)

// GenInflationRateChange randomized InflationRateChange
func GenInflationRateChange(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (inflation sdk.Dec) {
	ap.GetOrGenerate(cdc, Inflation, &inflation, r,
		func(r *rand.Rand) {
			inflation = sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
		})
	return
}

// GenInflation randomized Inflation
func GenInflation(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (inflationRateChange sdk.Dec) {
	ap.GetOrGenerate(cdc, InflationRateChange, &inflationRateChange, r,
		func(r *rand.Rand) {
			inflationRateChange = sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
		})
	return
}

// GenInflationMax randomized InflationMax
func GenInflationMax(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (inflationMax sdk.Dec) {
	ap.GetOrGenerate(cdc, InflationMax, &inflationMax, r,
		func(r *rand.Rand) {
			inflationMax = sdk.NewDecWithPrec(20, 2)
		})
	return
}

// GenInflationMin randomized InflationMin
func GenInflationMin(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (inflationMin sdk.Dec) {
	ap.GetOrGenerate(cdc, InflationMin, &inflationMin, r,
		func(r *rand.Rand) {
			inflationMin = sdk.NewDecWithPrec(7, 2)
		})
	return
}

// GenGoalBonded randomized GoalBonded
func GenGoalBonded(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (goalBonded sdk.Dec) {
	ap.GetOrGenerate(cdc, GoalBonded, &goalBonded, r,
		func(r *rand.Rand) {
			goalBonded = sdk.NewDecWithPrec(67, 2)
		})
	return
}

// GenMintGenesisState generates a random GenesisState for mint
func GenMintGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	// minter
	inflation := GenInflation(cdc, r, ap)
	minter := mint.InitialMinter(inflation)

	// params
	inflationRateChange := GenInflationRateChange(cdc, r, ap)
	mintDenom := sdk.DefaultBondDenom
	inflationMax := GenInflationMax(cdc, r, ap)
	inflationMin := GenInflationMin(cdc, r, ap)
	goalBonded := GenGoalBonded(cdc, r, ap)
	blocksPerYear := uint64(60 * 60 * 8766 / 5)
	params := mint.NewParams(mintDenom, inflationRateChange, inflationMax, inflationMin, goalBonded, blocksPerYear)

	mintGenesis := mint.NewGenesisState(minter, params)

	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, mintGenesis))
	genesisState[mint.ModuleName] = cdc.MustMarshalJSON(mintGenesis)
}
