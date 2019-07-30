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

// GenMintGenesisState generates a random GenesisState for mint
func GenMintGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	mintGenesis := mint.NewGenesisState(
		mint.InitialMinter(
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.Inflation, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.Inflation](r).(sdk.Dec)
					})
				return v
			}(r),
		),
		mint.NewParams(
			sdk.DefaultBondDenom,
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.InflationRateChange, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.InflationRateChange](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.InflationMax, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.InflationMax](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.InflationMin, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.InflationMin](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.GoalBonded, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.GoalBonded](r).(sdk.Dec)
					})
				return v
			}(r),
			uint64(60*60*8766/5),
		),
	)

	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, mintGenesis.Params))
	genesisState[mint.ModuleName] = cdc.MustMarshalJSON(mintGenesis)
}
