package simulation

import (
	"math/rand"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// OpWeightSubmitCommunitySpendProposal app params key for community spend proposal
const OpWeightSubmitCommunitySpendProposal = "op_weight_submit_community_spend_proposal"

// ProposalContents defines the module weighted proposals' contents
func ProposalMessages(k keeper.Keeper) []simtypes.WeightedProposalMessageSim {
	return []simtypes.WeightedProposalMessageSim{
		simulation.NewWeightedProposalMessageSim(
			OpWeightSubmitCommunitySpendProposal,
			simappparams.DefaultWeightCommunitySpendProposal,
			SimulateCommunityPoolSpendProposalMessage(k),
		),
	}
}

// SimulateCommunityPoolSpendProposalContent generates random community-pool-spend proposal content
func SimulateCommunityPoolSpendProposalMessage(k keeper.Keeper) simtypes.ProposalSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) []sdk.Msg {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		balance := k.GetFeePool(ctx).CommunityPool
		if balance.Empty() {
			return nil
		}

		denomIndex := r.Intn(len(balance))
		amount, err := simtypes.RandPositiveInt(r, balance[denomIndex].Amount.TruncateInt())
		if err != nil {
			return nil
		}

		return []sdk.Msg{types.NewMsgSpendCommunityPool(
			sdk.NewCoins(sdk.NewCoin(balance[denomIndex].Denom, amount)),
			simAccount.Address,
		)}
	}
}
