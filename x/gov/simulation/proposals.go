package simulation

import (
	"math/rand"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// OpWeightSubmitTextProposal app params key for text proposal
const OpWeightSubmitTextProposal = "op_weight_submit_text_proposal"

// ProposalContents defines the module weighted proposals' contents
func ProposalMessages() []simtypes.WeightedProposalMessageSim {
	return []simtypes.WeightedProposalMessageSim{
		simulation.NewWeightedProposalMessageSim(
			OpWeightMsgDeposit,
			simappparams.DefaultWeightTextProposal,
			SimulateSignalProposal,
		),
	}
}

func SimulateSignalProposal(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) []sdk.Msg {
	return []sdk.Msg{types.NewMsgSignal(
		simtypes.RandStringOfLength(r, 140),
		simtypes.RandStringOfLength(r, 5000),
	)}
}
