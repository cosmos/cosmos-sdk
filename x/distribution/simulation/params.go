package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
)


// Simulation parameter constants
const (
	CommunityTax             = "community_tax"
	BaseProposerReward       = "base_proposer_reward"
	BonusProposerReward      = "bonus_proposer_reward"
)

// GenParams generates random distribution parameters
func GenParams(paramSims map[string]func(r *rand.Rand) interface{}) {
	paramSims[CommunityTax] = func(r *rand.Rand) interface{} {
		return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
	}
	
	paramSims[BaseProposerReward] = func(r *rand.Rand) interface{} {
		return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
	}
	
	paramSims[BonusProposerReward] = func(r *rand.Rand) interface{} {
		return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2))
	}
}