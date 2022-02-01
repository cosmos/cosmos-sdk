package v046

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
)

// ConvertToLegacyProposal takes a new proposal and attempts to convert it to the
// legacy proposal format. This conversion is best effort. New proposal types that
// don't have a legacy message will return a "nil" content.
func ConvertToLegacyProposal(proposal v1beta2.Proposal) (v1beta1.Proposal, error) {
	var err error
	legacyProposal := v1beta1.Proposal{
		ProposalId:   proposal.ProposalId,
		Status:       v1beta1.ProposalStatus(proposal.Status),
		TotalDeposit: types.NewCoins(proposal.TotalDeposit...),
	}

	legacyProposal.FinalTallyResult, err = ConvertToLegacyTallyResult(proposal.FinalTallyResult)
	if err != nil {
		return v1beta1.Proposal{}, err
	}

	if proposal.VotingStartTime != nil {
		legacyProposal.VotingStartTime = *proposal.VotingStartTime
	}

	if proposal.VotingEndTime != nil {
		legacyProposal.VotingEndTime = *proposal.VotingEndTime
	}

	if proposal.SubmitTime != nil {
		legacyProposal.SubmitTime = *proposal.SubmitTime
	}

	if proposal.DepositEndTime != nil {
		legacyProposal.DepositEndTime = *proposal.DepositEndTime
	}

	msgs, err := proposal.GetMsgs()
	if err != nil {
		return v1beta1.Proposal{}, err
	}
	for _, msg := range msgs {
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

func ConvertToLegacyTallyResult(tally *v1beta2.TallyResult) (v1beta1.TallyResult, error) {
	yes, ok := types.NewIntFromString(tally.Yes)
	if !ok {
		return v1beta1.TallyResult{}, fmt.Errorf("unable to convert yes tally string (%s) to int", tally.Yes)
	}
	no, ok := types.NewIntFromString(tally.No)
	if !ok {
		return v1beta1.TallyResult{}, fmt.Errorf("unable to convert no tally string (%s) to int", tally.No)
	}
	veto, ok := types.NewIntFromString(tally.NoWithVeto)
	if !ok {
		return v1beta1.TallyResult{}, fmt.Errorf("unable to convert no with veto tally string (%s) to int", tally.NoWithVeto)
	}
	abstain, ok := types.NewIntFromString(tally.Abstain)
	if !ok {
		return v1beta1.TallyResult{}, fmt.Errorf("unable to convert abstain tally string (%s) to int", tally.Abstain)
	}

	return v1beta1.TallyResult{
		Yes:        yes,
		No:         no,
		NoWithVeto: veto,
		Abstain:    abstain,
	}, nil
}

func ConvertToLegacyVote(vote v1beta2.Vote) (v1beta1.Vote, error) {
	options, err := ConvertToLegacyVoteOptions(vote.Options)
	if err != nil {
		return v1beta1.Vote{}, err
	}
	return v1beta1.Vote{
		ProposalId: vote.ProposalId,
		Voter:      vote.Voter,
		Options:    options,
	}, nil
}

func ConvertToLegacyVoteOptions(voteOptions []*v1beta2.WeightedVoteOption) ([]v1beta1.WeightedVoteOption, error) {
	options := make([]v1beta1.WeightedVoteOption, len(voteOptions))
	for i, option := range voteOptions {
		weight, err := types.NewDecFromStr(option.Weight)
		if err != nil {
			return options, err
		}
		options[i] = v1beta1.WeightedVoteOption{
			Option: v1beta1.VoteOption(option.Option),
			Weight: weight,
		}
	}
	return options, nil
}

func ConvertToLegacyDeposit(deposit *v1beta2.Deposit) v1beta1.Deposit {
	return v1beta1.Deposit{
		ProposalId: deposit.ProposalId,
		Depositor:  deposit.Depositor,
		Amount:     types.NewCoins(deposit.Amount...),
	}
}
