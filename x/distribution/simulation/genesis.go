package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Simulation parameter constants
const (
	CommunityTax        = "community_tax"
	BaseProposerReward  = "base_proposer_reward"
	BonusProposerReward = "bonus_proposer_reward"
)

// GenCommunityTax randomized CommunityTax
func GenCommunityTax(cdc *codec.Codec, r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
}

// GenBaseProposerReward randomized BaseProposerReward
func GenBaseProposerReward(cdc *codec.Codec, r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
}

// GenBonusProposerReward randomized BonusProposerReward
func GenBonusProposerReward(cdc *codec.Codec, r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
}

// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(cdc *codec.Codec, r *rand.Rand, genesisState map[string]json.RawMessage) {

	communityTax := GenCommunityTax(cdc, r)
	baseProposerReward := GenBaseProposerReward(cdc, r)
	bonusProposerReward := GenBonusProposerReward(cdc, r)

	distrGenesis := types.GenesisState{
		FeePool:             types.InitialFeePool(),
		CommunityTax:        communityTax,
		BaseProposerReward:  baseProposerReward,
		BonusProposerReward: bonusProposerReward,
	}

	fmt.Printf("Selected randomly generated distribution parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, distrGenesis))
	genesisState[types.ModuleName] = cdc.MustMarshalJSON(distrGenesis)
}
