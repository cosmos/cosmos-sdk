package simulation

import (
	simappparams "github.com/Stride-Labs/cosmos-sdk/simapp/params"
	simtypes "github.com/Stride-Labs/cosmos-sdk/types/simulation"
	"github.com/Stride-Labs/cosmos-sdk/x/simulation"
)

// OpWeightSubmitParamChangeProposal app params key for param change proposal
const OpWeightSubmitParamChangeProposal = "op_weight_submit_param_change_proposal"

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(paramChanges []simtypes.ParamChange) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSubmitParamChangeProposal,
			simappparams.DefaultWeightParamChangeProposal,
			SimulateParamChangeProposalContent(paramChanges),
		),
	}
}
