package simulation

import (
	"context"
	"math/rand"

	"cosmossdk.io/x/gov/types/v1beta1"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// OpWeightSubmitTextProposal app params key for text proposal
const (
	OpWeightSubmitTextProposal = "op_weight_submit_text_proposal"
	DefaultWeightTextProposal  = 5
)

// ProposalContents defines the module weighted proposals' contents
//
//nolint:staticcheck // used for legacy testing
func ProposalContents() []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSubmitTextProposal,
			DefaultWeightTextProposal,
			SimulateLegacyTextProposalContent,
		),
	}
}

// SimulateLegacyTextProposalContent returns a random text proposal content.
//
//nolint:staticcheck // used for legacy testing
func SimulateLegacyTextProposalContent(r *rand.Rand, _ context.Context, _ []simtypes.Account) simtypes.Content {
	return v1beta1.NewTextProposal(
		simtypes.RandStringOfLength(r, 140),
		simtypes.RandStringOfLength(r, 5000),
	)
}
