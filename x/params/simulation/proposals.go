package simulation

import (
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const (
	// OpWeightSubmitParamChangeProposal app params key for param change proposal
	OpWeightSubmitParamChangeProposal = "op_weight_submit_param_change_proposal"
	DefaultWeightParamChangeProposal  = 5
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(paramChanges []simtypes.LegacyParamChange) []simtypes.WeightedProposalContent { //nolint:staticcheck // keep support for simulation gov v1beta1
	return []simtypes.WeightedProposalContent{ //nolint:staticcheck // keep support for simulation gov v1beta1
		simulation.NewWeightedProposalContent(
			OpWeightSubmitParamChangeProposal,
			DefaultWeightParamChangeProposal,
			SimulateParamChangeProposalContent(paramChanges),
		),
	}
}
