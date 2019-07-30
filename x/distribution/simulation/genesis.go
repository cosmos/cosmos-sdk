package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// GenDistrGenesisState generates a random GenesisState for distribution
func GenDistrGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	distrGenesis := distribution.GenesisState{
		FeePool: distribution.InitialFeePool(),
		CommunityTax: func(r *rand.Rand) sdk.Dec {
			var v sdk.Dec
			ap.GetOrGenerate(cdc, simulation.CommunityTax, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.CommunityTax](r).(sdk.Dec)
				})
			return v
		}(r),
		BaseProposerReward: func(r *rand.Rand) sdk.Dec {
			var v sdk.Dec
			ap.GetOrGenerate(cdc, simulation.BaseProposerReward, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.BaseProposerReward](r).(sdk.Dec)
				})
			return v
		}(r),
		BonusProposerReward: func(r *rand.Rand) sdk.Dec {
			var v sdk.Dec
			ap.GetOrGenerate(cdc, simulation.BonusProposerReward, &v, r,
				func(r *rand.Rand) {
					v = simulation.ModuleParamSimulator[simulation.BonusProposerReward](r).(sdk.Dec)
				})
			return v
		}(r),
	}

	fmt.Printf("Selected randomly generated distribution parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, distrGenesis))
	genesisState[distribution.ModuleName] = cdc.MustMarshalJSON(distrGenesis)
}
