package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
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
func GenInflationRateChange(cdc *codec.Codec, r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenInflation randomized Inflation
func GenInflation(cdc *codec.Codec, r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenInflationMax randomized InflationMax
func GenInflationMax(cdc *codec.Codec, r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(20, 2)
}

// GenInflationMin randomized InflationMin
func GenInflationMin(cdc *codec.Codec, r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(7, 2)
}

// GenGoalBonded randomized GoalBonded
func GenGoalBonded(cdc *codec.Codec, r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(67, 2)
}

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(cdc *codec.Codec, r *rand.Rand, genesisState map[string]json.RawMessage) {
	// minter
	inflation := GenInflation(cdc, r)
	minter := types.InitialMinter(inflation)

	// params
	inflationRateChange := GenInflationRateChange(cdc, r)
	mintDenom := sdk.DefaultBondDenom
	inflationMax := GenInflationMax(cdc, r)
	inflationMin := GenInflationMin(cdc, r)
	goalBonded := GenGoalBonded(cdc, r)
	blocksPerYear := uint64(60 * 60 * 8766 / 5)
	params := types.NewParams(mintDenom, inflationRateChange, inflationMax, inflationMin, goalBonded, blocksPerYear)

	mintGenesis := types.NewGenesisState(minter, params)

	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, mintGenesis))
	genesisState[types.ModuleName] = cdc.MustMarshalJSON(mintGenesis)
}
