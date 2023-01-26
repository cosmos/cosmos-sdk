package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// OpWeightSubmitTextProposal app params key for text proposal
const OpWeightSubmitTextProposal = "op_weight_submit_text_proposal"

// ProposalMsgs defines the module weighted proposals' contents
func ProposalMsgs() []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			OpWeightSubmitTextProposal,
			DefaultWeightTextProposal,
			SimulateTextProposal,
		),
	}
}

// SimulateTextProposal returns a random text proposal content.
// A text proposal is a proposal that contains no msgs.
func SimulateTextProposal(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
	return nil
}
