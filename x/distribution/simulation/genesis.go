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

// Simulation parameter constants
const (
	CommunityTax             = "community_tax"
	BaseProposerReward       = "base_proposer_reward"
	BonusProposerReward      = "bonus_proposer_reward"
)

// GenDistrGenesisState generates a random GenesisState for distribution
func GenDistrGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {

	var communityTax sdk.Dec
	ap.GetOrGenerate(cdc, CommunityTax, &communityTax, r,
		func(r *rand.Rand) {
			communityTax = sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
		})

	var baseProposerReward sdk.Dec
	ap.GetOrGenerate(cdc, BaseProposerReward, &baseProposerReward, r,
		func(r *rand.Rand) {
			baseProposerReward = sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
		})
	
	var bonusProposerReward sdk.Dec
	ap.GetOrGenerate(cdc, BonusProposerReward, &bonusProposerReward, r,
		func(r *rand.Rand) {
			bonusProposerReward = sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
		})

	distrGenesis := distribution.GenesisState{
		FeePool: distribution.InitialFeePool(),
		CommunityTax: communityTax,
		BaseProposerReward: baseProposerReward,
		BonusProposerReward: bonusProposerReward,
	}

	fmt.Printf("Selected randomly generated distribution parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, distrGenesis))
	genesisState[distribution.ModuleName] = cdc.MustMarshalJSON(distrGenesis)
}
