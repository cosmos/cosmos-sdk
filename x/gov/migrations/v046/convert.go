package v046

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

func convertToNewDeposits(oldDeps v1beta1.Deposits) v1beta2.Deposits {
	newDeps := make([]*v1beta2.Deposit, len(oldDeps))
	for i, oldDep := range oldDeps {
		newDeps[i] = &v1beta2.Deposit{
			ProposalId: oldDep.ProposalId,
			Depositor:  oldDep.Depositor,
			Amount:     oldDep.Amount,
		}
	}

	return newDeps
}

func convertToNewVotes(oldVotes v1beta1.Votes) (v1beta2.Votes, error) {
	newVotes := make([]*v1beta2.Vote, len(oldVotes))
	for i, oldVote := range oldVotes {
		var newWVOs []*v1beta2.WeightedVoteOption

		// We deprecated Vote.Option in v043. However, it might still be set.
		// - if only Options is set, or both Option & Options are set, we read from Options,
		// - if Options is not set, and Option is set, we read from Option,
		// - if none are set, we throw error.
		if oldVote.Options != nil {
			newWVOs = make([]*v1beta2.WeightedVoteOption, len(oldVote.Options))
			for j, oldWVO := range oldVote.Options {
				newWVOs[j] = v1beta2.NewWeightedVoteOption(v1beta2.VoteOption(oldWVO.Option), oldWVO.Weight)
			}
		} else if oldVote.Option != v1beta1.OptionEmpty {
			newWVOs = v1beta2.NewNonSplitVoteOption(v1beta2.VoteOption(oldVote.Option))
		} else {
			return nil, fmt.Errorf("vote does not have neither Options nor Option")
		}

		newVotes[i] = &v1beta2.Vote{
			ProposalId: oldVote.ProposalId,
			Voter:      oldVote.Voter,
			Options:    newWVOs,
		}
	}

	return newVotes, nil
}

func convertToNewDepParams(oldDepParams v1beta1.DepositParams) v1beta2.DepositParams {
	return v1beta2.DepositParams{
		MinDeposit:       oldDepParams.MinDeposit,
		MaxDepositPeriod: &oldDepParams.MaxDepositPeriod,
	}
}

func convertToNewVotingParams(oldVoteParams v1beta1.VotingParams) v1beta2.VotingParams {
	return v1beta2.VotingParams{
		VotingPeriod: &oldVoteParams.VotingPeriod,
	}
}

func convertToNewTallyParams(oldTallyParams v1beta1.TallyParams) v1beta2.TallyParams {
	return v1beta2.TallyParams{
		Quorum:        oldTallyParams.Quorum.String(),
		Threshold:     oldTallyParams.Threshold.String(),
		VetoThreshold: oldTallyParams.VetoThreshold.String(),
	}
}

func convertToNewProposal(oldProp v1beta1.Proposal) (v1beta2.Proposal, error) {
	msg, err := v1beta2.NewLegacyContent(oldProp.GetContent(), authtypes.NewModuleAddress(ModuleName).String())
	if err != nil {
		return v1beta2.Proposal{}, err
	}
	msgAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return v1beta2.Proposal{}, err
	}

	return v1beta2.Proposal{
		ProposalId: oldProp.ProposalId,
		Messages:   []*codectypes.Any{msgAny},
		Status:     v1beta2.ProposalStatus(oldProp.Status),
		FinalTallyResult: &v1beta2.TallyResult{
			Yes:        oldProp.FinalTallyResult.Yes.String(),
			No:         oldProp.FinalTallyResult.No.String(),
			Abstain:    oldProp.FinalTallyResult.Abstain.String(),
			NoWithVeto: oldProp.FinalTallyResult.NoWithVeto.String(),
		},
		SubmitTime:      &oldProp.SubmitTime,
		DepositEndTime:  &oldProp.DepositEndTime,
		TotalDeposit:    oldProp.TotalDeposit,
		VotingStartTime: &oldProp.VotingStartTime,
		VotingEndTime:   &oldProp.VotingEndTime,
	}, nil
}

func convertToNewProposals(oldProps v1beta1.Proposals) (v1beta2.Proposals, error) {
	newProps := make([]*v1beta2.Proposal, len(oldProps))
	for i, oldProp := range oldProps {
		p, err := convertToNewProposal(oldProp)
		if err != nil {
			return nil, err
		}

		newProps[i] = &p
	}

	return newProps, nil
}
