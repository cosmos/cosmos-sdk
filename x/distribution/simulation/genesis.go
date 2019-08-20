package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Simulation parameter constants
const (
	CommunityTax        = "community_tax"
	BaseProposerReward  = "base_proposer_reward"
	BonusProposerReward = "bonus_proposer_reward"
)

// GenCommunityTax randomized CommunityTax
func GenCommunityTax(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
}

// GenBaseProposerReward randomized BaseProposerReward
func GenBaseProposerReward(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
}

// GenBonusProposerReward randomized BonusProposerReward
func GenBonusProposerReward(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
}

// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(input *module.GeneratorInput) {

	var (
		communityTax        sdk.Dec
		baseProposerReward  sdk.Dec
		bonusProposerReward sdk.Dec
	)

	input.AppParams.GetOrGenerate(input.Cdc, CommunityTax, &communityTax, input.R,
		func(r *rand.Rand) { communityTax = GenCommunityTax(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, BaseProposerReward, &baseProposerReward, input.R,
		func(r *rand.Rand) { baseProposerReward = GenBaseProposerReward(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, BonusProposerReward, &bonusProposerReward, input.R,
		func(r *rand.Rand) { bonusProposerReward = GenBonusProposerReward(input.R) })

	distrGenesis := types.GenesisState{
		FeePool:             types.InitialFeePool(),
		CommunityTax:        communityTax,
		BaseProposerReward:  baseProposerReward,
		BonusProposerReward: bonusProposerReward,
	}

	fmt.Printf("Selected randomly generated distribution parameters:\n%s\n", codec.MustMarshalJSONIndent(input.Cdc, distrGenesis))
	input.GenState[types.ModuleName] = input.Cdc.MustMarshalJSON(distrGenesis)
}
