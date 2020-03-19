package simulation

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"math/rand"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// OpWeightSubmitTextProposal app params key for text proposal
const OpWeightSubmitTextProposal = "op_weight_submit_text_proposal"

// ProposalContents defines the module weighted proposals' contents
func ProposalContents() []module.WeightedProposalContent {
	return []simulation.WeightedProposalContent{
		{
			appParamsKey:       OpWeightSubmitTextProposal,
			defaultWeight:      simappparams.DefaultWeightTextProposal,
			contentSimulatorFn: SimulateTextProposalContent,
		},
	}
}

// SimulateTextProposalContent returns a random text proposal content.
func SimulateTextProposalContent(r *rand.Rand, _ sdk.Context, _ []simulation.Account) types.Content {
	return types.NewTextProposal(
		simulation.RandStringOfLength(r, 140),
		simulation.RandStringOfLength(r, 5000),
	)
}
