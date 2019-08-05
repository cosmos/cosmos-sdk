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

// GenMintGenesisState generates a random GenesisState for mint
func GenMintGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	// minter
	var inflation sdk.Dec
	ap.GetOrGenerate(cdc, Inflation, &inflation, r,
		func(r *rand.Rand) {
			inflation = sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
		})
	
	minter := mint.InitialMinter(inflation)

	// params
	var inflationRateChange sdk.Dec
	ap.GetOrGenerate(cdc, InflationRateChange, &inflationRateChange, r,
		func(r *rand.Rand) {
			inflationRateChange = sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
		})

	mintDenom := sdk.DefaultBondDenom

	var inflationMax sdk.Dec
	ap.GetOrGenerate(cdc, InflationMax, &inflationMax, r,
		func(r *rand.Rand) {
			inflationMax = sdk.NewDecWithPrec(20, 2)
		})

	var inflationMin sdk.Dec
	ap.GetOrGenerate(cdc, InflationMin, &inflationMin, r,
		func(r *rand.Rand) {
			inflationMin = sdk.NewDecWithPrec(7, 2)
		})

	var goalBonded sdk.Dec
	ap.GetOrGenerate(cdc, GoalBonded, &goalBonded, r,
		func(r *rand.Rand) {
			goalBonded = sdk.NewDecWithPrec(67, 2)
		})
	
	blocksPerYear := uint64(60*60*8766/5)
	
	params := mint.NewParams(mintDenom, inflationRateChange, inflationMax, inflationMin, goalBonded, blocksPerYear)
	
	mintGenesis := mint.NewGenesisState(minter, params)
 
	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, mintGenesis))
	genesisState[mint.ModuleName] = cdc.MustMarshalJSON(mintGenesis)
}
