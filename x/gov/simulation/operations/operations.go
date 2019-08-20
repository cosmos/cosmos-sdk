package operations

// DONTCOVER

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	distrsimops "github.com/cosmos/cosmos-sdk/x/distribution/simulation/operations"
	"github.com/cosmos/cosmos-sdk/x/gov"
	paramsimops "github.com/cosmos/cosmos-sdk/x/params/simulation/operations"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation parameter constants
const (
	OpWeightSubmitVotingSlashingTextProposal           = "op_weight_submit_voting_slashing_text_proposal"
	OpWeightSubmitVotingSlashingCommunitySpendProposal = "op_weight_submit_voting_slashing_community_spend_proposal"
	OpWeightSubmitVotingSlashingParamChangeProposal    = "op_weight_submit_voting_slashing_param_change_proposal"
	OpWeightMsgDeposit                                 = "op_weight_msg_deposit"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, keeper gov.Keeper) simulation.WeightedOperations {

	var (
		weightSubmitVotingSlashingTextProposal           int
		weightSubmitVotingSlashingCommunitySpendProposal int
		weightSubmitVotingSlashingParamChangeProposal    int
		weightMsgDeposit                                 int
	)

	appParams.GetOrGenerate(cdc, OpWeightSubmitVotingSlashingTextProposal, &weightSubmitVotingSlashingTextProposal, nil,
		func(_ *rand.Rand) { weightSubmitVotingSlashingTextProposal = 5 })

	appParams.GetOrGenerate(cdc, OpWeightSubmitVotingSlashingCommunitySpendProposal, &weightSubmitVotingSlashingCommunitySpendProposal, nil,
		func(_ *rand.Rand) { weightSubmitVotingSlashingCommunitySpendProposal = 5 })

	appParams.GetOrGenerate(cdc, OpWeightSubmitVotingSlashingParamChangeProposal, &weightSubmitVotingSlashingParamChangeProposal, nil,
		func(_ *rand.Rand) { weightSubmitVotingSlashingParamChangeProposal = 5 })

	appParams.GetOrGenerate(cdc, OpWeightMsgDeposit, &weightMsgDeposit, nil,
		func(_ *rand.Rand) { weightMsgDeposit = 100 })

	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			weightSubmitVotingSlashingTextProposal,
			SimulateSubmittingVotingAndSlashingForProposal(keeper, SimulateTextProposalContent),
		),
		simulation.NewWeigthedOperation(
			weightSubmitVotingSlashingCommunitySpendProposal,
			SimulateSubmittingVotingAndSlashingForProposal(keeper, distrsimops.SimulateCommunityPoolSpendProposalContent(app.distrKeeper)),
		),
		simulation.NewWeigthedOperation(
			weightSubmitVotingSlashingParamChangeProposal,
			SimulateSubmittingVotingAndSlashingForProposal(keeper, paramsimops.SimulateParamChangeProposalContent(app.sm.ParamChanges)),
		),
		simulation.NewWeigthedOperation(
			weightMsgDeposit,
			SimulateMsgDeposit(keeper),
		),
	}
}
