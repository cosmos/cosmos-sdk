package operations

// DONTCOVER

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	distrsimops "github.com/cosmos/cosmos-sdk/x/distribution/simulation/operations"
	paramsimops "github.com/cosmos/cosmos-sdk/x/params/simulation/operations"
)

// Simulation parameter constants
const (
	OpWeightSubmitVotingSlashingTextProposal           = "op_weight_submit_voting_slashing_text_proposal"
	OpWeightSubmitVotingSlashingCommunitySpendProposal = "op_weight_submit_voting_slashing_community_spend_proposal"
	OpWeightSubmitVotingSlashingParamChangeProposal    = "op_weight_submit_voting_slashing_param_change_proposal"
	OpWeightMsgDeposit                                 = "op_weight_msg_deposit"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(cdc *codec.Codec, keeper gov.Keeper) simulation.WeightedOperations {
	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightSubmitVotingSlashingTextProposal, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			SimulateSubmittingVotingAndSlashingForProposal(keeper, SimulateTextProposalContent),
		),
		simulation.NewWeigthedOperation(
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightSubmitVotingSlashingCommunitySpendProposal, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			SimulateSubmittingVotingAndSlashingForProposal(keeper, distrsim.SimulateCommunityPoolSpendProposalContent(app.distrKeeper)),
		),
		simulation.NewWeigthedOperation(
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightSubmitVotingSlashingParamChangeProposal, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			SimulateSubmittingVotingAndSlashingForProposal(keeper, paramsim.SimulateParamChangeProposalContent),
		),
		simulation.NewWeigthedOperation(
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgDeposit, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			SimulateMsgDeposit(keeper),
		),
	}
}
