package v046

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
)

// ConvertToLegacyProposal takes a new proposal and attempts to convert it to the
// legacy proposal format. This conversion is best effort. New proposal types that
// don't have a legacy message will return a "nil" content.
func ConvertToLegacyProposal(proposal v1beta2.Proposal) (v1beta1.Proposal, error) {
	legacyProposal := v1beta1.Proposal{
		ProposalId: proposal.ProposalId,
		Content: nil,
		Status: v1beta1.ProposalStatus(proposal.Status),
		FinalTallyResult: ConvertToLegacyTallyResult(proposal.FinalTallyResult),
		SubmitTime: *proposal.SubmitTime,
		DepositEndTime: *proposal.DepositEndTime,
		TotalDeposit: types.NewCoins(proposal.TotalDeposit...),
		VotingStartTime: *proposal.VotingStartTime,
		VotingEndTime: *proposal.VotingEndTime,
	}

	msgs, err := proposal.GetMsgs()
	if err != nil {
		return v1beta1.Proposal{}, err
	}
	msgTypes := make([]string, len(msgs))
	for idx, msg := range msgs {
		msgTypes[idx] = msg.String()
		if legacyMsg, ok := msg.(*v1beta2.MsgExecLegacyContent); ok {
			// check that the content struct can be unmarshalled
			_, err := v1beta2.LegacyContentFromMessage(legacyMsg)
			if err != nil {
				return v1beta1.Proposal{}, err
			}
			legacyProposal.Content = legacyMsg.Content
		}
	}
	return legacyProposal, nil
}

// ConvertToNewProposal takes a legacy proposal and converts it to a new proposal, 
// wrapping the content into a "MsgExecLegacyContent". 
func ConvertToNewProposal(proposal v1beta1.Proposal) v1beta2.Proposal {
	return v1beta2.Proposal{}
}

func ConvertToLegacyTallyResult(tally *v1beta2.TallyResult) v1beta1.TallyResult {
	yes, _ := types.NewIntFromString(tally.Yes)
	no, _ := types.NewIntFromString(tally.No)
	veto, _ := types.NewIntFromString(tally.NoWithVeto)
	abstain, _ := types.NewIntFromString(tally.Abstain)

	return v1beta1.TallyResult{
		Yes: yes,
		No: no,
		NoWithVeto: veto,
		Abstain: abstain,
	}
}