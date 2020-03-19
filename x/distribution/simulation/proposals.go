package simulation

import (
	simulation2 "github.com/cosmos/cosmos-sdk/types/simulation"
	"math/rand"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// OpWeightSubmitCommunitySpendProposal app params key for community spend proposal
const OpWeightSubmitCommunitySpendProposal = "op_weight_submit_community_spend_proposal"

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper) []simulation2.WeightedProposalContent {
	return []simulation2.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSubmitCommunitySpendProposal,
			simappparams.DefaultWeightCommunitySpendProposal,
			SimulateCommunityPoolSpendProposalContent(k),
		),
	}
}

// SimulateCommunityPoolSpendProposalContent generates random community-pool-spend proposal content
func SimulateCommunityPoolSpendProposalContent(k keeper.Keeper) simulation2.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simulation2.Account) simulation2.Content {
		simAccount, _ := simulation2.RandomAcc(r, accs)

		balance := k.GetFeePool(ctx).CommunityPool
		if balance.Empty() {
			return nil
		}

		denomIndex := r.Intn(len(balance))
		amount, err := simulation2.RandPositiveInt(r, balance[denomIndex].Amount.TruncateInt())
		if err != nil {
			return nil
		}

		return types.NewCommunityPoolSpendProposal(
			simulation2.RandStringOfLength(r, 10),
			simulation2.RandStringOfLength(r, 100),
			simAccount.Address,
			sdk.NewCoins(sdk.NewCoin(balance[denomIndex].Denom, amount)),
		)
	}
}
