package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// OpWeightSubmitCommunitySpendProposal app params key for community spend proposal
const OpWeightSubmitCommunitySpendProposal = "op_weight_submit_community_spend_proposal"

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper) []module.WeightedProposalContent {
	return []module.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSubmitCommunitySpendProposal,
			simappparams.DefaultWeightCommunitySpendProposal,
			SimulateCommunityPoolSpendProposalContent(k),
		),
	}
}

// SimulateCommunityPoolSpendProposalContent generates random community-pool-spend proposal content
func SimulateCommunityPoolSpendProposalContent(k keeper.Keeper) module.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []module.Account) module.Content {
		simAccount, _ := module.RandomAcc(r, accs)

		balance := k.GetFeePool(ctx).CommunityPool
		if balance.Empty() {
			return nil
		}

		denomIndex := r.Intn(len(balance))
		amount, err := module.RandPositiveInt(r, balance[denomIndex].Amount.TruncateInt())
		if err != nil {
			return nil
		}

		return types.NewCommunityPoolSpendProposal(
			module.RandStringOfLength(r, 10),
			module.RandStringOfLength(r, 100),
			simAccount.Address,
			sdk.NewCoins(sdk.NewCoin(balance[denomIndex].Denom, amount)),
		)
	}
}
